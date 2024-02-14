package linkerdmanager

import (
	"context"
	"fmt"
	"testing"

	linkerdserver "github.com/linkerd/linkerd2/controller/gen/apis/server/v1beta1"
	otterizev1alpha3 "github.com/otterize/intents-operator/src/operator/api/v1alpha3"
	"github.com/otterize/intents-operator/src/shared/injectablerecorder"
	"github.com/otterize/intents-operator/src/shared/testbase"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func TestLinkerdManagerTestSuite(t *testing.T) {
	suite.Run(t, new(LinkerdManagerTestSuite))
}
