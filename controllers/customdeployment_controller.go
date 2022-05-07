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
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crdsv1 "kubernetes-operator-assignment/api/v1"

	"github.com/redhat-cop/operator-utils/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CustomDeploymentReconciler reconciles a CustomDeployment object
type CustomDeploymentReconciler struct {
	util.ReconcilerBase
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
	// logger := log.FromContext(ctx)

	// logger.Info(fmt.Sprintf("\n---\nreq = %+v\n---\n", req))
	key := types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Namespace,
	}

	customDeployment := &crdsv1.CustomDeployment{}
	deployment := &appsv1.Deployment{}
	service := &corev1.Service{}

	// Get Custom Deployment
	err := r.Get(ctx, key, customDeployment)
	if err != nil && !errors.IsNotFound(err) {
		return requeue, err
	}

	customDeployment.ReplaceEmptyFieldsWithDefaultValues()
	// logger.Info(fmt.Sprintf("\n---\ncustomDeployment = %+v\n---\n", customDeployment))

	if customDeployment.GetDeletionTimestamp() == nil {
		err = r.Client.Get(ctx, key, deployment)
		if err != nil {
			if errors.IsNotFound(err) {
				if err = r.createNewDeployment(ctx, customDeployment); err != nil {
					return requeue, err
				}
			} else {
				return requeue, err
			}
		}

		// Get/Create Service
		err = r.Client.Get(ctx, key, service)
		if err != nil {
			if errors.IsNotFound(err) {
				if err = r.createNewService(ctx, customDeployment); err != nil {
					return requeue, err
				}
			} else {
				return requeue, err
			}
		}
	} else {
		err = r.Client.Get(ctx, key, deployment)
		if err == nil {
			err = r.Client.Delete(ctx, deployment)
			if err != nil {
				return requeue, err
			}
		} else if err != nil && !errors.IsNotFound(err) {
			return requeue, err
		}

		err = r.Get(ctx, key, service)
		if err == nil {
			err = r.Client.Delete(ctx, service)
			if err != nil {
				return requeue, err
			}
		} else if err != nil && !errors.IsNotFound(err) {
			return requeue, err
		}
	}

	// if util.IsBeingDeleted(customDeployment) {
	// 	if !util.HasFinalizer(customDeployment, "CustomDeploymentReconciler") {
	// 		logger.Info("GREAT")
	// 		return reconcile.Result{}, nil
	// 	}

	// 	logger.Info("STILL GOOD BUT NOT AS GREAT")
	// } else {
	// 	logger.Info("NOT GOOD")
	// }

	// isApplicationMarkedToBeDeleted := customDeployment.GetDeletionTimestamp() != nil
	// if isApplicationMarkedToBeDeleted {
	// 	logger.Info("YAY")
	// } else {
	// 	logger.Info("BOO")
	// }

	// logger.Info(fmt.Sprintf("\n---\nclient = %+v\nscheme=%+v\n---\n", r.Client, r.Scheme))
	// Get/Create Regular Deployment

	return done, nil
}

func (r *CustomDeploymentReconciler) createNewDeployment(ctx context.Context, customDeployment *crdsv1.CustomDeployment) error {
	return r.Create(ctx, &appsv1.Deployment{
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
							ContainerPort: 80,
						}},
					}},
				},
			},
		},
	})
}

func (r *CustomDeploymentReconciler) createNewService(ctx context.Context, customDeployment *crdsv1.CustomDeployment) error {

	err := r.Create(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      customDeployment.Name,
			Namespace: customDeployment.Namespace,
			Annotations: map[string]string{
				"metallb.universe.tf/address-pool": "production-public-ips",
			},
		},
		Spec: corev1.ServiceSpec{
			// Type: "NodePort",
			Type: "LoadBalancer",
			Selector: map[string]string{
				"app": customDeployment.Spec.Image.Name,
			},
			Ports: []corev1.ServicePort{{
				Protocol:   "TCP",
				Port:       80,
				TargetPort: intstr.FromInt(int(80)),
				NodePort:   *customDeployment.Spec.Port,
			}},
			ExternalIPs: []string{customDeployment.Spec.Host},
		},
	})

	if err != nil {
		return err
	}

	return r.Create(ctx, &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      customDeployment.Name,
			Namespace: customDeployment.Namespace,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{{
				Host: customDeployment.Spec.Host,
				IngressRuleValue: v1beta1.IngressRuleValue{
					HTTP: &v1beta1.HTTPIngressRuleValue{
						Paths: []v1beta1.HTTPIngressPath{{
							Backend: v1beta1.IngressBackend{
								ServiceName: customDeployment.Spec.Image.Name,
								ServicePort: intstr.FromInt(int(*customDeployment.Spec.Port)),
							},
						}},
					},
				},
			}},
		},
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.CustomDeployment{}).
		Complete(r)
}
