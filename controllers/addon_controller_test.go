package controllers

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/keikoproj/addon-manager/api/v1alpha1"
)

var (
	addonNamespace = "addon-manager-system"
	addonName      = "addon-bad"
	addonKey       = types.NamespacedName{Name: addonName, Namespace: addonNamespace}
)

const timeout = time.Second * 5

var _ = Describe("AddonController", func() {

	Describe("Addon CR can be reconciled", func() {
		var instance *v1alpha1.Addon

		It("instance should be parsable", func() {
			addonYaml, err := ioutil.ReadFile("../hack/tests/addon-bad.yaml")
			Expect(err).ToNot(HaveOccurred())

			instance, err = parseAddonYaml(addonYaml)
			Expect(err).ToNot(HaveOccurred())
			Expect(instance).To(BeAssignableToTypeOf(&v1alpha1.Addon{}))
			Expect(instance.GetName()).To(Equal(addonName))
		})

		It("instance should be reconciled", func() {
			instance.SetNamespace(addonNamespace)
			err := k8sClient.Create(context.TODO(), instance)
			if apierrors.IsInvalid(err) {
				log.Error(err, "failed to create object, got an invalid object error")
				return
			}
			Expect(err).NotTo(HaveOccurred())
			defer k8sClient.Delete(context.TODO(), instance)

			Eventually(func() error {
				if err := k8sClient.Get(context.TODO(), addonKey, instance); err != nil {
					return err
				}

				if instance.Status.Checksum == "" {
					return fmt.Errorf("checksum was empty")
				}
				return nil
			}, timeout).Should(Succeed())

			By("Verify addon has been reconciled by checking for checksum status")
			Expect(instance.Status.Checksum).ShouldNot(BeEmpty())
		})
	})
})

func parseAddonYaml(data []byte) (*v1alpha1.Addon, error) {
	var err error
	o := &unstructured.Unstructured{}
	err = yaml.Unmarshal(data, &o.Object)
	if err != nil {
		return nil, err
	}
	a := &v1alpha1.Addon{}
	err = scheme.Scheme.Convert(o, a, 0)
	if err != nil {
		return nil, err
	}

	return a, nil
}
