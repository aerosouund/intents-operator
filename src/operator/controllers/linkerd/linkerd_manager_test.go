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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapiv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

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
	fmt.Println(intents)
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
