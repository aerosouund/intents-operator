package linkerdmanager

import (
	"context"
	"fmt"
	"testing"

	authpolicy "github.com/linkerd/linkerd2/controller/gen/apis/policy/v1alpha1"
	linkerdserver "github.com/linkerd/linkerd2/controller/gen/apis/server/v1beta1"
	otterizev1alpha3 "github.com/otterize/intents-operator/src/operator/api/v1alpha3"
	"github.com/otterize/intents-operator/src/shared/injectablerecorder"
	"github.com/otterize/intents-operator/src/shared/testbase"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapiv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

type policyWrapper struct {
	*authpolicy.AuthorizationPolicy
}

type routeWrapper struct {
	*authpolicy.HTTPRoute
}

func (rw routeWrapper) String() string {
	return fmt.Sprintf("string httproute wrapper")
}

func (rw routeWrapper) Matches(x interface{}) bool {
	valueAsRoute, ok := x.(*authpolicy.HTTPRoute)
	if !ok {
		fmt.Println("value passed cannot be casted into a httproute")
		return false
	}

	if rw.Namespace != valueAsRoute.Namespace {
		return false
	}

	if string(*rw.Spec.Rules[0].Matches[0].Path.Value) != string(*valueAsRoute.Spec.Rules[0].Matches[0].Path.Value) {
		return false
	}
	if string(*rw.Spec.ParentRefs[0].Kind) != string(*valueAsRoute.Spec.ParentRefs[0].Kind) {
		return false
	}
	if string(rw.Spec.ParentRefs[0].Name) != string(valueAsRoute.Spec.ParentRefs[0].Name) {
		return false
	}

	return true
}

func (pw policyWrapper) String() string {
	return fmt.Sprintf("string policy wrapper")
}

func (pw policyWrapper) Matches(x interface{}) bool {
	valueAsPolicy, ok := x.(*authpolicy.AuthorizationPolicy)
	if !ok {
		fmt.Println("value passed cannot be casted into an authorization policy")
		return false
	}

	if pw.Namespace != valueAsPolicy.Namespace {
		return false
	}

	if string(pw.Spec.TargetRef.Kind) != string(valueAsPolicy.Spec.TargetRef.Kind) {
		return false
	}

	if string(pw.Spec.RequiredAuthenticationRefs[0].Name) != string(valueAsPolicy.Spec.RequiredAuthenticationRefs[0].Name) {
		return false
	}

	if string(pw.Spec.RequiredAuthenticationRefs[0].Kind) != string(valueAsPolicy.Spec.RequiredAuthenticationRefs[0].Kind) {
		return false
	}
	return true
}

type LinkerdManagerTestSuite struct {
	testbase.MocksSuiteBase
	admin *LinkerdManager
}

func (s *LinkerdManagerTestSuite) SetupTest() {
	s.MocksSuiteBase.SetupTest()
	s.admin = NewLinkerdManager(s.Client, []string{}, &injectablerecorder.InjectableRecorder{Recorder: s.Recorder}, true, true)
}

func (s *LinkerdManagerTestSuite) TearDownTest() {
	s.admin = nil
	s.MocksSuiteBase.TearDownTest()
}

func (s *LinkerdManagerTestSuite) TestShouldntCreateServer() {
	ns := "test-namespace"
	var port int32 = 3000

	intentsSpec := &otterizev1alpha3.IntentsSpec{
		Service: otterizev1alpha3.Service{Name: "service-that-calls"},
		Calls: []otterizev1alpha3.Intent{
			{
				Name: "test-service",
			},
		},
	}

	intents := otterizev1alpha3.ClientIntents{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intents-object",
			Namespace: ns,
		},
		Spec: intentsSpec,
	}

	podSelector := s.admin.BuildPodLabelSelectorFromIntent(intents.Spec.Calls[0], intents.Namespace)
	serversList := &linkerdserver.ServerList{
		Items: []linkerdserver.Server{
			{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "policy.linkerd.io/v1beta1",
					Kind:       "Server",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "server-for-test-service-port-3000",
					Namespace: ns,
				},
				Spec: linkerdserver.ServerSpec{
					PodSelector: &podSelector,
					Port:        intstr.FromInt32(port),
				},
			},
		},
	}

	s.Client.EXPECT().List(gomock.Any(), gomock.Any(), &client.ListOptions{Namespace: ns}).Do(func(_ context.Context, emptyServersList *linkerdserver.ServerList, _ ...client.ListOption) {
		emptyServersList.Items = append(emptyServersList.Items, serversList.Items...)
	}).Return(nil)
	_, shouldCreateServer, _ := s.admin.shouldCreateServer(context.Background(), intents, intents.Spec.Calls[0], port)
	s.Equal(false, shouldCreateServer)
}

func (s *LinkerdManagerTestSuite) TestCreateResourcesNonHTTPIntent() {
	ns := "test-namespace"

	intentsSpec := &otterizev1alpha3.IntentsSpec{
		Service: otterizev1alpha3.Service{Name: "service-that-calls"},
		Calls: []otterizev1alpha3.Intent{
			{
				Name: "test-service",
			},
		},
	}

	intents := otterizev1alpha3.ClientIntents{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intents-object",
			Namespace: ns,
		},
		Spec: intentsSpec,
	}

	formattedTargetServer := otterizev1alpha3.GetFormattedOtterizeIdentity(intents.Spec.Calls[0].GetTargetServerName(), ns)
	linkerdServerServiceFormattedIdentity := otterizev1alpha3.GetFormattedOtterizeIdentity(intents.GetServiceName(), intents.Namespace)
	podSelector := s.admin.BuildPodLabelSelectorFromIntent(intents.Spec.Calls[0], intents.Namespace)

	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				otterizev1alpha3.OtterizeServerLabelKey: formattedTargetServer,
			},
			Name:      "example-pod",
			Namespace: ns,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Ports: []v1.ContainerPort{
						{
							ContainerPort: 8000,
						},
					},
				},
			},
		},
	}

	netAuth := &authpolicy.NetworkAuthentication{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy.linkerd.io/v1alpha1",
			Kind:       "NetworkAuthentication",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      NetworkAuthenticationNameTemplate,
			Namespace: intents.Namespace,
			Labels: map[string]string{
				otterizev1alpha3.OtterizeLinkerdServerAnnotationKey: linkerdServerServiceFormattedIdentity,
			},
		},
		Spec: authpolicy.NetworkAuthenticationSpec{
			Networks: []*authpolicy.Network{
				{
					Cidr: "0.0.0.0/0",
				},
				{
					Cidr: "::0",
				},
			},
		},
	}

	mtlsAuth := &authpolicy.MeshTLSAuthentication{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy.linkerd.io/v1alpha1",
			Kind:       "MeshTLSAuthentication",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(OtterizeLinkerdMeshTLSNameTemplate, intents.Spec.Service.Name),
			Namespace: intents.Namespace,
			Labels: map[string]string{
				otterizev1alpha3.OtterizeLinkerdServerAnnotationKey: linkerdServerServiceFormattedIdentity,
			},
		},
		Spec: authpolicy.MeshTLSAuthenticationSpec{
			Identities: []string{"default.test-namespace.serviceaccount.identity.linkerd.cluster.local"},
		},
	}

	server := &linkerdserver.Server{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy.linkerd.io/v1beta1",
			Kind:       "Server",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "server-for-test-service-port-8000",
			Namespace: ns,
			Labels: map[string]string{
				otterizev1alpha3.OtterizeLinkerdServerAnnotationKey: linkerdServerServiceFormattedIdentity,
			},
		},
		Spec: linkerdserver.ServerSpec{
			PodSelector: &podSelector,
			Port:        intstr.FromInt32(8000),
		},
	}

	policy := &authpolicy.AuthorizationPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy.linkerd.io/v1alpha1",
			Kind:       "AuthorizationPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "authpolicy-to-test-service-port-8000-from-client-service-that-calls-knng3afe",
			Namespace: intents.Namespace,
			Labels: map[string]string{
				otterizev1alpha3.OtterizeLinkerdServerAnnotationKey: linkerdServerServiceFormattedIdentity,
			},
		},
		Spec: authpolicy.AuthorizationPolicySpec{
			TargetRef: gatewayapiv1alpha2.PolicyTargetReference{
				Group: "policy.linkerd.io",
				Kind:  v1beta1.Kind("Server"),
				Name:  v1beta1.ObjectName("server-for-test-service-port-8000"),
			},
			RequiredAuthenticationRefs: []gatewayapiv1alpha2.PolicyTargetReference{
				{
					Group: "policy.linkerd.io",
					Kind:  v1beta1.Kind("MeshTLSAuthentication"),
					Name:  "meshtls-for-client-service-that-calls",
				},
			},
		},
	}

	s.Client.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any()).Do(func(_ context.Context, podList *v1.PodList, _ ...client.ListOption) {
		podList.Items = append(podList.Items, pod)
	})
	s.Client.EXPECT().List(gomock.Any(), gomock.Any(), &client.ListOptions{Namespace: ns}).Return(nil)
	s.Client.EXPECT().Create(gomock.Any(), netAuth).Return(nil)
	s.Client.EXPECT().List(gomock.Any(), gomock.Any(), &client.ListOptions{Namespace: ns}).Return(nil)
	s.Client.EXPECT().Create(gomock.Any(), mtlsAuth).Return(nil)
	s.Client.EXPECT().List(gomock.Any(), gomock.Any(), &client.ListOptions{Namespace: ns}).Return(nil)
	s.Client.EXPECT().Create(gomock.Any(), server)
	s.Client.EXPECT().List(gomock.Any(), gomock.Any(), &client.ListOptions{Namespace: ns}).Return(nil)
	s.Client.EXPECT().Create(gomock.Any(), policyWrapper{policy})

	_, err := s.admin.createResources(context.Background(), &intents, "default")
	s.NoError(err)
}

func (s *LinkerdManagerTestSuite) TestCreateResourcesHTTPIntent() {
	ns := "test-namespace"

	intentsSpec := &otterizev1alpha3.IntentsSpec{
		Service: otterizev1alpha3.Service{Name: "service-that-calls"},
		Calls: []otterizev1alpha3.Intent{
			{
				Name: "test-service",
				Type: "http",
				HTTPResources: []otterizev1alpha3.HTTPResource{
					{
						Path: "/api",
					},
				},
			},
		},
	}

	intents := otterizev1alpha3.ClientIntents{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intents-object",
			Namespace: ns,
		},
		Spec: intentsSpec,
	}

	formattedTargetServer := otterizev1alpha3.GetFormattedOtterizeIdentity(intents.Spec.Calls[0].GetTargetServerName(), ns)
	linkerdServerServiceFormattedIdentity := otterizev1alpha3.GetFormattedOtterizeIdentity(intents.GetServiceName(), intents.Namespace)
	podSelector := s.admin.BuildPodLabelSelectorFromIntent(intents.Spec.Calls[0], intents.Namespace)

	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				otterizev1alpha3.OtterizeServerLabelKey: formattedTargetServer,
			},
			Name:      "example-pod",
			Namespace: ns,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Ports: []v1.ContainerPort{
						{
							ContainerPort: 8000,
						},
					},
				},
			},
		},
	}

	netAuth := &authpolicy.NetworkAuthentication{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy.linkerd.io/v1alpha1",
			Kind:       "NetworkAuthentication",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      NetworkAuthenticationNameTemplate,
			Namespace: intents.Namespace,
			Labels: map[string]string{
				otterizev1alpha3.OtterizeLinkerdServerAnnotationKey: linkerdServerServiceFormattedIdentity,
			},
		},
		Spec: authpolicy.NetworkAuthenticationSpec{
			Networks: []*authpolicy.Network{
				{
					Cidr: "0.0.0.0/0",
				},
				{
					Cidr: "::0",
				},
			},
		},
	}

	mtlsAuth := &authpolicy.MeshTLSAuthentication{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy.linkerd.io/v1alpha1",
			Kind:       "MeshTLSAuthentication",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(OtterizeLinkerdMeshTLSNameTemplate, intents.Spec.Service.Name),
			Namespace: intents.Namespace,
			Labels: map[string]string{
				otterizev1alpha3.OtterizeLinkerdServerAnnotationKey: linkerdServerServiceFormattedIdentity,
			},
		},
		Spec: authpolicy.MeshTLSAuthenticationSpec{
			Identities: []string{"default.test-namespace.serviceaccount.identity.linkerd.cluster.local"},
		},
	}

	server := &linkerdserver.Server{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy.linkerd.io/v1beta1",
			Kind:       "Server",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "server-for-test-service-port-8000",
			Namespace: ns,
			Labels: map[string]string{
				otterizev1alpha3.OtterizeLinkerdServerAnnotationKey: linkerdServerServiceFormattedIdentity,
			},
		},
		Spec: linkerdserver.ServerSpec{
			PodSelector: &podSelector,
			Port:        intstr.FromInt32(8000),
		},
	}

	route := &authpolicy.HTTPRoute{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy.linkerd.io/v1beta3",
			Kind:       "HTTPRoute",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "http-route-for-test-service-port-8000-knng3afe",
			Namespace: ns,
			Labels: map[string]string{
				otterizev1alpha3.OtterizeLinkerdServerAnnotationKey: linkerdServerServiceFormattedIdentity,
			},
		},
		Spec: authpolicy.HTTPRouteSpec{
			CommonRouteSpec: v1beta1.CommonRouteSpec{
				ParentRefs: []v1beta1.ParentReference{
					{
						Group: (*v1beta1.Group)(StringPtr("policy.linkerd.io")),
						Kind:  (*v1beta1.Kind)(StringPtr("Server")),
						Name:  v1beta1.ObjectName("server-for-test-service-port-8000"),
					},
				},
			},
			Rules: []authpolicy.HTTPRouteRule{
				{
					Matches: []authpolicy.HTTPRouteMatch{
						{
							Path: &authpolicy.HTTPPathMatch{
								Type:  getPathMatchPointer(authpolicy.PathMatchPathPrefix),
								Value: StringPtr("/api"),
							},
						},
					},
				},
			},
		},
	}

	policy := &authpolicy.AuthorizationPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy.linkerd.io/v1alpha1",
			Kind:       "AuthorizationPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bs",
			Namespace: ns,
			Labels: map[string]string{
				otterizev1alpha3.OtterizeLinkerdServerAnnotationKey: linkerdServerServiceFormattedIdentity,
			},
		},
		Spec: authpolicy.AuthorizationPolicySpec{
			TargetRef: gatewayapiv1alpha2.PolicyTargetReference{
				Group: "policy.linkerd.io",
				Kind:  v1beta1.Kind("HTTPRoute"),
				Name:  v1beta1.ObjectName("http-route-for-test-service-port-8000-knng3afe"),
			},
			RequiredAuthenticationRefs: []gatewayapiv1alpha2.PolicyTargetReference{
				{
					Group: "policy.linkerd.io",
					Kind:  v1beta1.Kind("MeshTLSAuthentication"),
					Name:  "meshtls-for-client-service-that-calls",
				},
			},
		},
	}

	s.Client.EXPECT().List(gomock.Any(), gomock.Any(), client.MatchingLabels{otterizev1alpha3.OtterizeServerLabelKey: formattedTargetServer},
		client.InNamespace(ns)).Do(func(_ context.Context, podList *v1.PodList, _ ...client.ListOption) {
		podList.Items = append(podList.Items, pod)
	})
	s.Client.EXPECT().List(gomock.Any(), gomock.Any(), &client.ListOptions{Namespace: ns}).Return(nil)
	s.Client.EXPECT().Create(gomock.Any(), netAuth).Return(nil)
	s.Client.EXPECT().List(gomock.Any(), gomock.Any(), &client.ListOptions{Namespace: ns}).Return(nil)
	s.Client.EXPECT().Create(gomock.Any(), mtlsAuth).Return(nil)
	s.Client.EXPECT().List(gomock.Any(), gomock.Any(), &client.ListOptions{Namespace: ns}).Return(nil)
	s.Client.EXPECT().Create(gomock.Any(), server)

	s.Client.EXPECT().List(gomock.Any(), gomock.Any(), &client.ListOptions{Namespace: ns}).Return(nil)
	s.Client.EXPECT().Create(gomock.Any(), routeWrapper{route})
	s.Client.EXPECT().List(gomock.Any(), gomock.Any(), &client.ListOptions{Namespace: ns}).Return(nil)
	s.Client.EXPECT().Create(gomock.Any(), policyWrapper{policy})

	_, err := s.admin.createResources(context.Background(), &intents, "default")
	_ = route
	s.NoError(err)
}

func (s *LinkerdManagerTestSuite) TestDeleteAll() {
	ns := "test-namespace"

	intents := otterizev1alpha3.ClientIntents{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "intents-object-1",
			Namespace: ns,
		},
		Spec: &otterizev1alpha3.IntentsSpec{
			Service: otterizev1alpha3.Service{Name: "service-that-calls"},
			Calls: []otterizev1alpha3.Intent{
				{
					Name: "test-service",
				},
			},
		},
	}
	linkerdServerServiceFormattedIdentity := otterizev1alpha3.GetFormattedOtterizeIdentity(intents.GetServiceName(), intents.Namespace)

	authPoliciesOfIntent := authpolicy.AuthorizationPolicyList{
		Items: []authpolicy.AuthorizationPolicy{
			{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "policy.linkerd.io/v1alpha1",
					Kind:       "AuthorizationPolicy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-policy",
					Namespace: intents.Namespace,
					Labels: map[string]string{
						otterizev1alpha3.OtterizeLinkerdServerAnnotationKey: linkerdServerServiceFormattedIdentity,
					},
				},
				Spec: authpolicy.AuthorizationPolicySpec{
					TargetRef: gatewayapiv1alpha2.PolicyTargetReference{
						Group: "policy.linkerd.io",
						Kind:  v1beta1.Kind("HTTPRoute"),
						Name:  v1beta1.ObjectName("some-route"),
					},
					RequiredAuthenticationRefs: []gatewayapiv1alpha2.PolicyTargetReference{
						{
							Group: "policy.linkerd.io",
							Kind:  v1beta1.Kind("MeshTLSAuthentication"),
							Name:  "meshtls-for-client-service-that-calls",
						},
					},
				},
			},
		},
	}

	authPoliciesOfEveryone := authpolicy.AuthorizationPolicyList{
		Items: []authpolicy.AuthorizationPolicy{
			{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "policy.linkerd.io/v1alpha1",
					Kind:       "AuthorizationPolicy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-policy",
					Namespace: intents.Namespace,
					Labels: map[string]string{
						otterizev1alpha3.OtterizeLinkerdServerAnnotationKey: linkerdServerServiceFormattedIdentity,
					},
				},
				Spec: authpolicy.AuthorizationPolicySpec{
					TargetRef: gatewayapiv1alpha2.PolicyTargetReference{
						Group: "policy.linkerd.io",
						Kind:  v1beta1.Kind("HTTPRoute"),
						Name:  v1beta1.ObjectName("some-route"),
					},
					RequiredAuthenticationRefs: []gatewayapiv1alpha2.PolicyTargetReference{
						{
							Group: "policy.linkerd.io",
							Kind:  v1beta1.Kind("MeshTLSAuthentication"),
							Name:  "meshtls-for-client-service-that-calls",
						},
					},
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "policy.linkerd.io/v1alpha1",
					Kind:       "AuthorizationPolicy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-other-policy",
					Namespace: intents.Namespace,
					Labels: map[string]string{
						otterizev1alpha3.OtterizeLinkerdServerAnnotationKey: "someone-else",
					},
				},
				Spec: authpolicy.AuthorizationPolicySpec{
					TargetRef: gatewayapiv1alpha2.PolicyTargetReference{
						Group: "policy.linkerd.io",
						Kind:  v1beta1.Kind("HTTPRoute"),
						Name:  v1beta1.ObjectName("some-route"),
					},
					RequiredAuthenticationRefs: []gatewayapiv1alpha2.PolicyTargetReference{
						{
							Group: "policy.linkerd.io",
							Kind:  v1beta1.Kind("MeshTLSAuthentication"),
							Name:  "meshtls-for-client-someone-else",
						},
					},
				},
			},
		},
	}

	routes := authpolicy.HTTPRouteList{
		Items: []authpolicy.HTTPRoute{
			{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "policy.linkerd.io/v1beta3",
					Kind:       "HTTPRoute",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-route",
					Namespace: intents.Namespace,
					Labels: map[string]string{
						otterizev1alpha3.OtterizeLinkerdServerAnnotationKey: linkerdServerServiceFormattedIdentity,
					},
				},
				Spec: authpolicy.HTTPRouteSpec{
					CommonRouteSpec: v1beta1.CommonRouteSpec{
						ParentRefs: []v1beta1.ParentReference{
							{
								Group: (*v1beta1.Group)(StringPtr("policy.linkerd.io")),
								Kind:  (*v1beta1.Kind)(StringPtr("Server")),
								Name:  v1beta1.ObjectName("server-for-test-service-port-6969"),
							},
						},
					},
				},
			},
		},
	}
	fmt.Println(routes)

	s.Client.EXPECT().List(gomock.Any(), gomock.Any(), &client.MatchingLabels{otterizev1alpha3.OtterizeLinkerdServerAnnotationKey: linkerdServerServiceFormattedIdentity}).Do(func(_ context.Context, policiesOfIntent *authpolicy.AuthorizationPolicyList, _ ...client.ListOption) {
		policiesOfIntent.Items = append(policiesOfIntent.Items, authPoliciesOfIntent.Items...)
	}).Return(nil)
	s.Client.EXPECT().List(gomock.Any(), gomock.Any(), &client.ListOptions{Namespace: ns}).Do(func(_ context.Context, allPoliciesInNS *authpolicy.AuthorizationPolicyList, _ ...client.ListOption) {
		allPoliciesInNS.Items = append(allPoliciesInNS.Items, authPoliciesOfEveryone.Items...)
	}).Return(nil)

	s.admin.DeleteAll(context.Background(), &intents)

}

func TestLinkerdManagerTestSuite(t *testing.T) {
	suite.Run(t, new(LinkerdManagerTestSuite))
}
