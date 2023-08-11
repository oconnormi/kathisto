/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"bytes"
	"context"
	"strings"
	tpl "text/template"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	templatev1beta1 "github.com/oconnormi/kathisto/api/v1beta1"
)

// TemplateReconciler reconciles a Template object
type TemplateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kathisto.oconnormi.io,resources=templates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kathisto.oconnormi.io,resources=templates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kathisto.oconnormi.io,resources=templates/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Template object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *TemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	var template templatev1beta1.Template
	if err := r.Get(ctx, req.NamespacedName, &template); err != nil {
		log.Error(err, "unable to fetch Template")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	inputContents := make(map[string]interface{})

	// Get all template inputs
	for _, input := range template.Spec.Inputs {
		inputFilter := types.NamespacedName{Namespace: req.Namespace, Name: input.Name}
		if input.Type == "secret" {
			var secretInput corev1.Secret
			if err := r.Get(ctx, inputFilter, &secretInput); err != nil {
				log.Error(err, "unable to fetch secret input")
				return ctrl.Result{}, err
			}
			vars := secretToVars(secretInput)
			inputContents[strings.ReplaceAll(input.Name, "-", "_")] = vars
		}
		if input.Type == "configmap" {
			var configmapInput corev1.ConfigMap
			if err := r.Get(ctx, inputFilter, &configmapInput); err != nil {
				log.Error(err, "unable to fetch input configmap")
				return ctrl.Result{}, err
			}
			vars := configMapToVars(configmapInput)
			inputContents[configmapInput.Name] = vars
		}
	}

	// Get source
	var source corev1.ConfigMap
	sourceFilter := types.NamespacedName{Name: template.Spec.Source.Name, Namespace: req.Namespace}
	if err := r.Get(ctx, sourceFilter, &source); err != nil {
		log.Error(err, "unable to fetch source configmap")
		return ctrl.Result{}, err
	}

	var output corev1.Secret
	outputFilter := types.NamespacedName{Name: template.Spec.Output, Namespace: req.Namespace}
	if err := r.Get(ctx, outputFilter, &output); err != nil {
		log.Info("output secret {secret} doesn't exist, creating")

		output = corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      template.Spec.Output,
				Namespace: req.Namespace,
			},
			Data: make(map[string][]byte),
		}
		if err := r.Create(ctx, &output); err != nil {
			log.Error(err, "unable to create output secret")
			return ctrl.Result{}, err
		}
	}

	for name, sourceTemplate := range source.Data {
		current, err := tpl.New(name).Parse(sourceTemplate)
		if err != nil {
			log.Error(err, "unable to parse template")
			return ctrl.Result{}, err
		}
		tplOutput := &bytes.Buffer{}
		if err := current.Execute(tplOutput, &inputContents); err != nil {
			log.Error(err, "unable to execute template")
		}
		if output.Data == nil {
			output.Data = make(map[string][]byte)
		}
		output.Data[name] = tplOutput.Bytes()
	}

	if err := r.Update(ctx, &output); err != nil {
		log.Error(err, "unable to update output secret")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func secretToVars(secret corev1.Secret) map[string]interface{} {
	vars := make(map[string]interface{})
	for key, value := range secret.Data {
		vars[key] = string(value)
	}
	return vars
}

func configMapToVars(configMap corev1.ConfigMap) map[string]interface{} {
	vars := make(map[string]interface{})
	for key, value := range configMap.Data {
		vars[key] = string(value)
	}
	return vars
}

// SetupWithManager sets up the controller with the Manager.
func (r *TemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&templatev1beta1.Template{}).
		Complete(r)
}
