package visitorssite

import (
	"context"

	visitorsv1alpha1 "github.com/jdob/visitors-operator/pkg/apis/visitors/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const backendPort = 8000
const backendServicePort = 30685
const backendImage = "jdob/visitors-service:latest"

func (r *ReconcileVisitorsSite) backendDeployment(v *visitorsv1alpha1.VisitorsSite) *appsv1.Deployment {
	labels := labels(v, "backend")
	size := v.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:		v.Name + "-backend",
			Namespace: 	v.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &size,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:	backendImage,
						ImagePullPolicy: corev1.PullAlways,
						Name:	"visitors-service",
						Ports:	[]corev1.ContainerPort{{
							ContainerPort: 	backendPort,
							Name:			"visitors",
						}},
						Env:	[]corev1.EnvVar{
							{
								Name:	"MYSQL_DATABASE",
								Value:	"visitors",
							},
						},
					}},
				},
			},
		},
	}

	controllerutil.SetControllerReference(v, dep, r.scheme)
	return dep
}

func (r *ReconcileVisitorsSite) backendService(v *visitorsv1alpha1.VisitorsSite) *corev1.Service {
	labels := labels(v, "backend")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:		v.Name + "-backend-service",
			Namespace: 	v.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol: corev1.ProtocolTCP,
				Port: backendPort,
				TargetPort: intstr.FromInt(backendPort),
				NodePort: 30685,
			}},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	controllerutil.SetControllerReference(v, s, r.scheme)
	return s
}

func (r *ReconcileVisitorsSite) updateBackendStatus(v *visitorsv1alpha1.VisitorsSite) (error) {
	v.Status.BackendImage = backendImage
	err := r.client.Status().Update(context.TODO(), v)
	return err
}