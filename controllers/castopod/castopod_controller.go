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

package castopod

import (
	"context"
	b64 "encoding/base64"
	"fmt"

	"github.com/kopilote/castopod-operator/apis/castopod/v1beta1"
	apisv1beta1 "github.com/kopilote/castopod-operator/pkg/apis/v1beta1"
	"github.com/kopilote/castopod-operator/pkg/controllerutils"
	. "github.com/kopilote/castopod-operator/pkg/typeutils"
	pkgError "github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CastopodMutator reconciles a Castopod object
type CastopodMutator struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=castopod.kopilote.io,resources=castopods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=castopod.kopilote.io,resources=castopods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=castopod.kopilote.io,resources=castopods/finalizers,verbs=update
// +kubebuilder:rbac:groups=castopod.kopilote.io,resources=configurations,verbs=get;list;watch
// +kubebuilder:rbac:groups=castopod.kopilote.io,resources=versions,verbs=get;list;watch

type config struct {
	App           *v1beta1.Castopod
	Configuration *v1beta1.Configuration
	Version       *v1beta1.Version
}

func (r *CastopodMutator) Mutate(ctx context.Context, app *v1beta1.Castopod) (*ctrl.Result, error) {
	apisv1beta1.SetProgressing(app)

	// Get Configuration Object
	configuration := &v1beta1.Configuration{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Name: app.Spec.ConfigurationSpec,
	}, configuration); err != nil {
		if errors.IsNotFound(err) {
			return nil, pkgError.New("Configuration object not found")
		}
		return controllerutils.Requeue(), fmt.Errorf("error retrieving Configuration object: %s", err)
	}

	// Get Version Object
	version := &v1beta1.Version{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Name: app.Spec.VersionSpec,
	}, version); err != nil {
		if errors.IsNotFound(err) {
			return nil, pkgError.New("Version object not found")
		}
		return controllerutils.Requeue(), fmt.Errorf("error retrieving Version object: %s", err)
	}

	// Create config Object
	appConfig := &config{
		App:           app,
		Configuration: configuration,
		Version:       version,
	}

	// Create Namespace
	if err := r.reconcileNamespace(ctx, appConfig); err != nil {
		return controllerutils.Requeue(), pkgError.Wrap(err, "Reconciling namespace")
	}

	ns := fmt.Sprintf("castopod-%s", appConfig.App.Name)
	app.SetNamespace(ns)

	// Reconcile for Castopod App
	castopodDeploy, err := r.reconcileDeploymentForApp(ctx, appConfig)
	if err != nil {
		return controllerutils.Requeue(), pkgError.Wrap(err, "Reconciling deployment for Castopod")
	}
	castopodService, err := r.reconcileServiceForApp(ctx, appConfig, castopodDeploy)
	if err != nil {
		return controllerutils.Requeue(), pkgError.Wrap(err, "Reconciling service for Castopod")
	}
	_, err = r.reconcileIngressForApp(ctx, appConfig, castopodService)
	if err != nil {
		return controllerutils.Requeue(), pkgError.Wrap(err, "Reconciling ingress for Castopod")
	}

	apisv1beta1.SetReady(app)
	return nil, nil

}

func generateEnv(config config) []corev1.EnvVar {
	env := []corev1.EnvVar{
		apisv1beta1.Env("CP_CACHE_HANDLER", "file"),
		apisv1beta1.Env("CP_APP_HOSTNAME", "localhost"),
		apisv1beta1.Env("CP_ANALYTICS_SALT", b64.StdEncoding.EncodeToString([]byte(config.App.Name))),
		apisv1beta1.Env("CP_DATABASE_NAME", fmt.Sprintf("castopod_%s", config.App.Name)),
		// URL
		apisv1beta1.Env("CP_BASEURL", fmt.Sprintf("https://%s", config.App.Spec.Config.URL.Base)),
		apisv1beta1.Env("CP_MEDIA_BASEURL", fmt.Sprintf("https://%s", config.App.Spec.Config.URL.Media)),
		apisv1beta1.Env("CP_LEGALNOTICE_BASEURL", fmt.Sprintf("https://%s", config.App.Spec.Config.URL.LegalNotice)),
		// Limit
		apisv1beta1.Env("CP_LIMIT_STORAGE", fmt.Sprintf("%d", config.App.Spec.Config.Limit.Storage)),
		apisv1beta1.Env("CP_LIMIT_BANDWIDTH", fmt.Sprintf("%d", config.App.Spec.Config.Limit.Bandwidth)),
		// Gateway
		apisv1beta1.Env("CP_ADMIN_GATEWAY", config.App.Spec.Config.Gateway.Admin),
		apisv1beta1.Env("CP_AUTH_GATEWAY", config.App.Spec.Config.Gateway.Auth),
		apisv1beta1.Env("CP_INSTALL_GATEWAY", config.App.Spec.Config.Gateway.Install),
		// SMTP TODO: Add SMTP env vars
	}
	env = append(env, config.Configuration.Spec.Mysql.Env("")...)
	return env
}

func (r *CastopodMutator) reconcileNamespace(ctx context.Context, config *config) error {
	log.FromContext(ctx).Info("Reconciling Namespace")

	_, operationResult, err := controllerutils.CreateOrUpdateWithController(ctx, r.Client, r.Scheme, types.NamespacedName{
		Name: fmt.Sprintf("castopod-%s", config.App.Name),
	}, config.App, func(ns *corev1.Namespace) error {
		// No additional mutate needed
		return nil
	})
	switch {
	case err != nil:
		apisv1beta1.SetNamespaceError(config.App, err.Error())
		return err
	case operationResult == controllerutil.OperationResultNone:
	default:
		apisv1beta1.SetNamespaceCreated(config.App)
	}

	log.FromContext(ctx).Info("Namespace ready")
	return nil
}

func (r *CastopodMutator) reconcileDeploymentForApp(ctx context.Context, config *config) (*appsv1.Deployment, error) {
	matchLabels := CreateMap("app.kubernetes.io/name", "castopod")

	ret, operationResult, err := controllerutils.CreateOrUpdateWithController(ctx, r.Client, r.Scheme, types.NamespacedName{
		Namespace: config.App.Namespace,
		Name:      config.App.Name,
	}, config.App, func(deployment *appsv1.Deployment) error {
		deployment.Spec = appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: matchLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "web",
							Image:           fmt.Sprintf("castopod/web-server:%s", config.Version.Spec.ImageTag),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env:             generateEnv(*config),
							Ports: []corev1.ContainerPort{{
								Name:          "castopod",
								ContainerPort: 80,
							}},
							LivenessProbe: controllerutils.DefaultLiveness(),
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    *resource.NewMilliQuantity(100, resource.DecimalSI),
									corev1.ResourceMemory: *resource.NewMilliQuantity(256, resource.DecimalSI),
								},
							},
						},
						{
							Name:            "app",
							Image:           fmt.Sprintf("castopod/app:%s", config.Version.Spec.ImageTag),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env:             generateEnv(*config),
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    *resource.NewMilliQuantity(100, resource.DecimalSI),
									corev1.ResourceMemory: *resource.NewMilliQuantity(256, resource.DecimalSI),
								},
							},
						},
						{
							Name:            "video-clipper",
							Image:           fmt.Sprintf("castopod/video-clipper:%s", config.Version.Spec.ImageTag),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env:             generateEnv(*config),
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    *resource.NewMilliQuantity(100, resource.DecimalSI),
									corev1.ResourceMemory: *resource.NewMilliQuantity(256, resource.DecimalSI),
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:            "init-create-db",
							Image:           "mysql:8",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"sh",
								"-c",
								`mysql -h ${CP_DATABASE_HOSTNAME} -P ${CP_DATABASE_PORT} -u ${CP_DATABASE_USERNAME} -p${CP_DATABASE_PASSWORD} -e "CREATE DATABASE IF NOT EXISTS ${CP_DATABASE_NAME};"`,
							},
							Env: generateEnv(*config),
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    *resource.NewMilliQuantity(100, resource.DecimalSI),
									corev1.ResourceMemory: *resource.NewMilliQuantity(256, resource.DecimalSI),
								},
							},
						},
					},
				},
			},
		}
		return nil
	})
	switch {
	case err != nil:
		apisv1beta1.SetDeploymentError(config.App, err.Error())
		return nil, err
	case operationResult == controllerutil.OperationResultNone:
	default:
		apisv1beta1.SetDeploymentReady(config.App)
	}

	return ret, err
}

func (r *CastopodMutator) reconcileServiceForApp(ctx context.Context, config *config, deployment *appsv1.Deployment) (*corev1.Service, error) {
	ret, operationResult, err := controllerutils.CreateOrUpdateWithController(ctx, r.Client, r.Scheme, types.NamespacedName{
		Namespace: config.App.Namespace,
		Name:      deployment.Name,
	}, config.App, func(service *corev1.Service) error {
		service.ObjectMeta.Annotations = config.Configuration.Spec.Ingress.Annotations
		service.Spec = corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:        "http",
				Port:        deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort,
				Protocol:    "TCP",
				AppProtocol: pointer.String("http"),
				TargetPort:  intstr.FromString(deployment.Spec.Template.Spec.Containers[0].Ports[0].Name),
			}},
			Selector: deployment.Spec.Template.Labels,
		}
		return nil
	})
	switch {
	case err != nil:
		apisv1beta1.SetServiceError(config.App, err.Error())
		return nil, err
	case operationResult == controllerutil.OperationResultNone:
	default:
		apisv1beta1.SetServiceReady(config.App)
	}
	return ret, err
}

func (r *CastopodMutator) reconcileIngressForApp(ctx context.Context, config *config, service *corev1.Service) (*networkingv1.Ingress, error) {
	annotations := config.Configuration.Spec.Ingress.Annotations
	var serviceDest string
	if config.App.Spec.Activated {
		serviceDest = service.Name
	} else {
		serviceDest = "castopod-nginx"
	}
	ret, operationResult, err := controllerutils.CreateOrUpdateWithController(ctx, r.Client, r.Scheme, types.NamespacedName{
		Namespace: config.App.Namespace,
		Name:      service.Name,
	}, config.App, func(ingress *networkingv1.Ingress) error {
		pathType := networkingv1.PathTypePrefix
		ingress.ObjectMeta.Annotations = annotations
		ingress.Spec = networkingv1.IngressSpec{
			TLS: config.Configuration.Spec.Ingress.TLS.AsK8SIngressTLSSlice(),
			Rules: []networkingv1.IngressRule{
				{
					Host: config.App.Spec.Config.URL.Base,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: serviceDest,
											Port: networkingv1.ServiceBackendPort{
												Name: service.Spec.Ports[0].Name,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		return nil
	})
	switch {
	case err != nil:
		apisv1beta1.SetIngressError(config.App, err.Error())
		return nil, err
	case operationResult == controllerutil.OperationResultNone:
	default:
		apisv1beta1.SetIngressReady(config.App)
	}
	return ret, nil
}

// SetupWithBuilder SetupWithManager sets up the controller with the Manager.
func (r *CastopodMutator) SetupWithBuilder(mgr ctrl.Manager, builder *ctrl.Builder) error {
	builder.
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{})
	return nil
}

func NewCastopodMutator(client client.Client, scheme *runtime.Scheme) controllerutils.Mutator[*v1beta1.Castopod] {
	return &CastopodMutator{
		Client: client,
		Scheme: scheme,
	}
}
