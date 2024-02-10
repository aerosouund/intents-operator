package intents_reconcilers

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/otterize/intents-operator/src/operator/api/v1alpha3"
	otterizev1alpha3 "github.com/otterize/intents-operator/src/operator/api/v1alpha3"
	mocks "github.com/otterize/intents-operator/src/operator/controllers/intents_reconcilers/mocks"
	"github.com/otterize/intents-operator/src/shared/testbase"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type LinkerdReconcilerTestSuite struct {
	testbase.MocksSuiteBase
	admin           *LinkerdReconciler
	ldm             *mocks.MockLinkerdPolicyManager
	serviceResolver *mocks.MockServiceResolver
}

func (s *LinkerdReconcilerTestSuite) SetupTest() {
	s.MocksSuiteBase.SetupTest()
	s.admin = NewLinkerdReconciler(s.Client, &runtime.Scheme{}, []string{}, true, true)
}

func (s *LinkerdReconcilerTestSuite) TearDownTest() {
	s.admin = nil
	s.MocksSuiteBase.TearDownTest()
}

func (s *LinkerdReconcilerTestSuite) TestCreatePolicy() {
	clientIntentsName := "client-intents"
	serviceName := "test-client"
	serverNamespace := "far-far-away"

	namespacedName := types.NamespacedName{
		Namespace: testNamespace,
		Name:      clientIntentsName,
	}
	req := ctrl.Request{
		NamespacedName: namespacedName,
	}

	serverName := fmt.Sprintf("test-server.%s", serverNamespace)
	intentsSpec := &otterizev1alpha3.IntentsSpec{
		Service: otterizev1alpha3.Service{Name: serviceName},
		Calls: []otterizev1alpha3.Intent{
			{
				Name: serverName,
			},
		},
	}

	intentsWithoutFinalizer := otterizev1alpha3.ClientIntents{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clientIntentsName,
			Namespace: testNamespace,
		},
		Spec: intentsSpec,
	}

	// Initial call to get the ClientIntents object when reconciler starts
	emptyIntents := &otterizev1alpha3.ClientIntents{}
	s.Client.EXPECT().Get(gomock.Any(), req.NamespacedName, gomock.Eq(emptyIntents)).DoAndReturn(
		func(ctx context.Context, name types.NamespacedName, intents *otterizev1alpha3.ClientIntents, options ...client.ListOption) error {
			intentsWithoutFinalizer.DeepCopyInto(intents)
			return nil
		})

	intentsObj := otterizev1alpha3.ClientIntents{}
	intentsWithoutFinalizer.DeepCopyInto(&intentsObj)

	clientServiceAccount := "test-server-sa"
	clientPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-client-fdae32",
			Namespace: serverNamespace,
		},
		Spec: v1.PodSpec{
			ServiceAccountName: clientServiceAccount,
			Containers: []v1.Container{
				{
					Name: "real-application-who-does-something",
				},
				{
					Name: "linkerd-proxy",
				},
			},
		},
	}

	serverPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-server-2b5e0d",
			Namespace: serverNamespace,
		},
		Spec: v1.PodSpec{
			ServiceAccountName: "test-server-sa",
			Containers: []v1.Container{
				{
					Name: "server-who-listens",
				},
				{
					Name: "linkerd-proxy",
				},
			},
		},
	}
	s.serviceResolver.EXPECT().ResolveClientIntentToPod(gomock.Any(), gomock.Eq(intentsObj)).Return(clientPod, nil)
	s.serviceResolver.EXPECT().ResolveIntentServerToPod(gomock.Any(), gomock.Eq(intentsObj.Spec.Calls[0]), serverNamespace).Return(serverPod, nil)
	s.ldm.EXPECT().Create(gomock.Any(), gomock.Eq(&intentsObj), clientServiceAccount).Return(nil)
	res, err := s.admin.Reconcile(context.Background(), req)
	s.NoError(err)
	s.Empty(res)
}

func (s *LinkerdReconcilerTestSuite) TestAnything() {
	_ = &v1alpha3.ClientIntents{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: &v1alpha3.IntentsSpec{
			Service: v1alpha3.Service{
				Name: clientName,
			},
			Calls: []v1alpha3.Intent{
				{
					Name: "test",
				},
			},
		},
	}

	// This object will be returned to the reconciler's Get call when it calls for a CRD named "servers.policy.linkerd.io"
	crd := &v12.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "servers.policy.linkerd.io",
		},
	}

	// The matchers here make it check for a CRD called "servers.policy.linkerd.io"
	s.Client.EXPECT().Get(gomock.Any(), types.NamespacedName{Name: "servers.policy.linkerd.io"}, gomock.AssignableToTypeOf(crd)).Do(
		func(ctx context.Context, key types.NamespacedName, obj *v12.CustomResourceDefinition, opts ...metav1.GetOptions) error {
			// Copy the struct into the target pointer struct
			*obj = *crd
			return nil
		})

	res, err := s.admin.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "test",
		},
	})
	s.Require().NoError(err)
	s.Require().Empty(res)
}

func TestLinkerdReconcilerSuite(t *testing.T) {
	suite.Run(t, new(LinkerdReconcilerTestSuite))
}
