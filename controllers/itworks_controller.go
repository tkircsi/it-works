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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	webappv1alpha1 "tkircsi/it-works/api/v1alpha1"
)

var log = logf.Log.WithName("controller_itworks")

// ItWorksReconciler reconciles a ItWorks object
type ItWorksReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.tkircsi,resources=itworks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.tkircsi,resources=itworks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.tkircsi,resources=itworks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ItWorks object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ItWorksReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling ItWorks")

	// Get the ItWorks instance
	instance := &webappv1alpha1.ItWorks{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found. Clean-up and don't requeue
			return ctrl.Result{}, nil
		}
		// Error getting instance, requeue
		return ctrl.Result{}, err
	}

	// Check Deployment and create if not exists
	res, err := r.ensureDeployment(ctx, instance, r.createDeployment(instance))
	if res != nil {
		return *res, err
	}

	// Check Service and create if not exists
	res, err = r.ensureService(ctx, instance, r.createService(instance))
	if res != nil {
		return *res, err
	}

	// Check Ingress and create if not exists
	res, err = r.ensureIngress(ctx, instance, r.createIngress(instance))
	if res != nil {
		return *res, err
	}

	reqLogger.Info("Deployment and Service exists.")
	return ctrl.Result{}, nil
}

func (r *ItWorksReconciler) ensureIngress(
	ctx context.Context,
	instance *webappv1alpha1.ItWorks,
	ing *netv1.Ingress,
) (*ctrl.Result, error) {
	found := &netv1.Ingress{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      ing.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the service
		log.Info("Creating a new Ingress", "Ingress.Namespace", ing.Namespace, "Ingress.Name", ing.Name)
		err = r.Client.Create(ctx, ing)

		if err != nil {
			// Service creation failed
			log.Error(err, "Failed to create new Ingress", "Ingress.Namespace", ing.Namespace, "Ingress.Name", ing.Name)
			return &ctrl.Result{}, err
		} else {
			// Service creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		log.Error(err, "Failed to get Ingress")
		return &ctrl.Result{}, err
	}
	return nil, nil
}

func (r *ItWorksReconciler) ensureService(
	ctx context.Context,
	instance *webappv1alpha1.ItWorks,
	s *corev1.Service,
) (*ctrl.Result, error) {
	// See if service already exists and create if it doesn't
	found := &corev1.Service{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the service
		log.Info("Creating a new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
		err = r.Client.Create(ctx, s)

		if err != nil {
			// Service creation failed
			log.Error(err, "Failed to create new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
			return &ctrl.Result{}, err
		} else {
			// Service creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		log.Error(err, "Failed to get Service")
		return &ctrl.Result{}, err
	}

	return nil, nil
}

func (r *ItWorksReconciler) ensureDeployment(
	ctx context.Context,
	instance *webappv1alpha1.ItWorks,
	dep *appsv1.Deployment,
) (*ctrl.Result, error) {
	found := &appsv1.Deployment{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      dep.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Deployment not found.
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Client.Create(ctx, dep)

		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return &ctrl.Result{}, err
		}

		// Deployment exists
		return nil, nil
	} else if err != nil {
		// Error but not NotFound
		log.Error(err, "Failed to get Deployment")
		return &ctrl.Result{}, err
	}
	return nil, nil
}

func (r *ItWorksReconciler) createIngress(v *webappv1alpha1.ItWorks) *netv1.Ingress {
	labels := map[string]string{
		"app": v.Name,
	}

	pathTypePrefix := netv1.PathTypePrefix

	ingress := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      v.Name,
			Namespace: v.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"cert-manager.io/issuer": "selfsigned-issuer",
			},
		},
		Spec: netv1.IngressSpec{
			TLS: []netv1.IngressTLS{{
				Hosts:      []string{v.Spec.Host},
				SecretName: "it-works-secret-tls",
			}},
			Rules: []netv1.IngressRule{{
				Host: v.Spec.Host,
				IngressRuleValue: netv1.IngressRuleValue{
					HTTP: &netv1.HTTPIngressRuleValue{
						Paths: []netv1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &pathTypePrefix,
								Backend: netv1.IngressBackend{
									Service: &netv1.IngressServiceBackend{
										Name: "it-works-service",
										Port: netv1.ServiceBackendPort{
											Number: 8080,
										},
									},
								},
							},
						},
					},
				},
			}},
		},
	}
	ctrl.SetControllerReference(v, ingress, r.Scheme)
	return ingress
}

func (r *ItWorksReconciler) createService(v *webappv1alpha1.ItWorks) *corev1.Service {
	labels := map[string]string{
		"app": v.Name,
	}

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "it-works-service",
			Namespace: v.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       8080,
				TargetPort: intstr.FromInt(80),
			}},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	ctrl.SetControllerReference(v, s, r.Scheme)
	return s
}

func (r *ItWorksReconciler) createDeployment(v *webappv1alpha1.ItWorks) *appsv1.Deployment {
	replicas := v.Spec.Replicas
	labels := map[string]string{
		"app": v.Name,
	}

	dep := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{
		Name:      v.Name,
		Namespace: v.Namespace,
	},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           v.Spec.Image,
						ImagePullPolicy: corev1.PullAlways,
						Name:            v.Name,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
							Name:          "http",
						}},
					}},
				},
			},
		},
	}
	ctrl.SetControllerReference(v, dep, r.Scheme)
	return dep
}

// SetupWithManager sets up the controller with the Manager.
func (r *ItWorksReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1alpha1.ItWorks{}).
		Complete(r)
}
