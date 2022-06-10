package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	templatev1beta1 "github.com/oconnormi/kathisto/api/v1beta1"
)

var _ = Describe("Template controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		TemplateName      = "test-template"
		TemplateNamespace = "default"
	)

	Context("When rendering a single template with a single input", func() {
		It("Should create a new secret", func() {
			By("Converting each key from the input into a template variable")
			ctx := context.Background()
			inputSecret := corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-input",
					Namespace: TemplateNamespace,
				},
				Data: map[string][]byte{
					"foo": []byte("bar"),
					"baz": []byte("qux"),
				},
			}
			Expect(k8sClient.Create(ctx, &inputSecret)).Should(Succeed())
			source := corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-source",
					Namespace: TemplateNamespace,
				},
				Data: map[string]string{
					"foo.yaml": "{{ .test-input.foo }}-{{ .test-input.baz }}",
				},
			}
			Expect(k8sClient.Create(ctx, &source)).Should(Succeed())
			template := &templatev1beta1.Template{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kathisto.oconnormi.io/v1beta1",
					Kind:       "Template",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      TemplateName,
					Namespace: TemplateNamespace,
				},
				Spec: templatev1beta1.TemplateSpec{
					Source: templatev1beta1.Source{
						Name: "test-source",
					},
					Inputs: []templatev1beta1.Input{
						{
							Name: "test-input",
							Type: "secret",
						},
					},
					Output: "test-output",
				},
			}
			Expect(k8sClient.Create(ctx, template)).Should(Succeed())
			outputFilter := types.NamespacedName{Namespace: TemplateNamespace, Name: "test-output"}
			var output corev1.Secret
			Expect(k8sClient.Get(ctx, outputFilter, &output)).Should(Succeed())
			expexted := map[string]string{}
			expexted["foo.yaml"] = "bar-qux"
			Expect(output.Data).Should(ConsistOf(expexted))
		})
	})
})
