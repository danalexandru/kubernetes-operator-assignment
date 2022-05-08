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

	cmacme "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	certmgv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/redhat-cop/operator-utils/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	// v1beta1 "k8s.io/api/extensions/v1beta1"
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
	key := types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Namespace,
	}

	customDeployment := &crdsv1.CustomDeployment{}
	deployment := &appsv1.Deployment{}
	service := &corev1.Service{}
	issuer := &certmgv1.ClusterIssuer{}

	// Get Custom Deployment
	err := r.Get(ctx, key, customDeployment)
	if err != nil && !errors.IsNotFound(err) {
		return requeue, err
	}

	customDeployment.ReplaceEmptyFieldsWithDefaultValues()

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

		// Get/Create Issuer
		err = r.Client.Get(ctx, key, issuer)
		if err != nil {
			if errors.IsNotFound(err) {
				if err = r.createNewIssuer(ctx, customDeployment); err != nil {
					return requeue, err
				}
			} else {
				return requeue, err
			}
		}
	} else {
		// Delete Deployment
		err = r.Client.Get(ctx, key, deployment)
		if err == nil {
			err = r.Client.Delete(ctx, deployment)
			if err != nil {
				return requeue, err
			}
		} else if err != nil && !errors.IsNotFound(err) {
			return requeue, err
		}

		// Delete Service
		err = r.Get(ctx, key, service)
		if err == nil {
			err = r.Client.Delete(ctx, service)
			if err != nil {
				return requeue, err
			}
		} else if err != nil && !errors.IsNotFound(err) {
			return requeue, err
		}

		// // Delete Issuer
		err = r.Client.Get(ctx, key, issuer)
		if err == nil {
			err = r.Client.Delete(ctx, issuer)
			if err != nil {
				return requeue, err
			}
		} else if err != nil && !errors.IsNotFound(err) {
			return requeue, err
		}
	}

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
	return r.Create(ctx, &corev1.Service{
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
			ExternalIPs: []string{"86.120.240.137"},
		},
	})
}

func (r *CustomDeploymentReconciler) createNewIssuer(ctx context.Context, customDeployment *crdsv1.CustomDeployment) error {
	var issuerClass string = "nginx"
	return r.Create(ctx, &certmgv1.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      customDeployment.Name,
			Namespace: customDeployment.Namespace,
		},
		Spec: certmgv1.IssuerSpec{
			IssuerConfig: certmgv1.IssuerConfig{
				ACME: &cmacme.ACMEIssuer{
					Server: "https://acme-v02.api.letsencrypt.org/directory",
					PrivateKey: cmmeta.SecretKeySelector{
						LocalObjectReference: cmmeta.LocalObjectReference{
							Name: fmt.Sprintf("%s-issuer-key", customDeployment.Name),
						},
					},
					Solvers: []cmacme.ACMEChallengeSolver{{
						HTTP01: &cmacme.ACMEChallengeSolverHTTP01{
							Ingress: &cmacme.ACMEChallengeSolverHTTP01Ingress{
								Class: &issuerClass,
							},
						},
					}},
				},
			},
		},
	})

}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.CustomDeployment{}).
		Complete(r)
}
