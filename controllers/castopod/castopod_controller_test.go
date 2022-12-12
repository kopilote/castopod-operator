package castopod

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kopilote/castopod-operator/apis/castopod/v1beta1"
	pkgapiv1beta1 "github.com/kopilote/castopod-operator/pkg/apis/v1beta1"
	. "github.com/kopilote/castopod-operator/pkg/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Castopod Controller", func() {
	mutator := NewCastopodMutator(GetClient(), GetScheme())
	WithMutator(mutator, func() {
		var (
			configuration *v1beta1.Configuration
			version       *v1beta1.Version
		)
		BeforeEach(func() {
			configuration = NewDumbConfiguration()
			version = NewDumbVersions()
		})
		When("Creating a configuration", func() {
			BeforeEach(func() {
				Expect(Create(configuration)).To(Succeed())
				Expect(Create(version)).To(Succeed())
			})
			Context("Then creating a Castopod", func() {
				var (
					castopod v1beta1.Castopod
				)
				BeforeEach(func() {
					name := uuid.NewString()
					castopod = v1beta1.NewStack(name, v1beta1.CastopodSpec{
						ConfigurationSpec: configuration.Name,
						VersionSpec:       version.Name,
						Activated:         true,
					})
					Expect(Create(&castopod)).To(Succeed())
					Eventually(ConditionStatus(&castopod, pkgapiv1beta1.ConditionTypeReady)).
						Should(Equal(metav1.ConditionTrue))
				})
				It("Should create a new namespace", func() {
					Expect(Get(types.NamespacedName{
						Name: fmt.Sprintf("castopod-%s", castopod.Name),
					}, &v1.Namespace{})).To(BeNil())
				})
			})
		})
	})
})
