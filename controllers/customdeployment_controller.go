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
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	crdsv1 "kubernetes-operator-assignment/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CustomDeploymentReconciler reconciles a CustomDeployment object
type CustomDeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	requeue = ctrl.Result{Requeue: true}
	done    = ctrl.Result{}
)

//+kubebuilder:rbac:groups=crds.k8s.op.asgn,resources=customdeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.k8s.op.asgn,resources=customdeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.k8s.op.asgn,resources=customdeployments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CustomDeployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *CustomDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	key := types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Namespace,
	}

	customDeployment := &crdsv1.CustomDeployment{}
	deployment := &appsv1.Deployment{}

	err := r.Get(ctx, key, customDeployment)
	if err != nil && !errors.IsNotFound(err) {
		return requeue, err
	}

	err = r.Get(ctx, key, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			r.createNewDeployment(ctx, customDeployment)
		}
	}

	return done, nil
}

func (r *CustomDeploymentReconciler) createNewDeployment(ctx context.Context, customDeployment *crdsv1.CustomDeployment) {
	customDeployment.ReplaceEmptyFieldsWithDefaultValues()

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      customDeployment.Name,
			Namespace: customDeployment.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: customDeployment.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": customDeployment.Spec.Image.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": customDeployment.Spec.Image.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  customDeployment.Spec.Image.Name,
						Image: fmt.Sprintf("%s:%s", customDeployment.Spec.Image.Name, customDeployment.Spec.Image.Tag),
						Ports: []corev1.ContainerPort{{
							ContainerPort: *customDeployment.Spec.Port,
						}},
					}},
				},
			},
		},
	}

	r.Create(ctx, deployment)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.CustomDeployment{}).
		Complete(r)
}
