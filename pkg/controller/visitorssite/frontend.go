package visitorssite

import (
	visitorsv1alpha1 "github.com/jdob/visitors-operator/pkg/apis/visitors/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const frontendPort = 3000
const frontendServicePort = 30686

func (r *ReconcileVisitorsSite) frontendDeployment(v *visitorsv1alpha1.VisitorsSite) *appsv1.Deployment {
	labels := frontendLabels(v)
	size := int32(1)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:		v.Name + "-frontend",
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
						Image:	"jdob/visitors-webui:latest",
						Name:	"visitors-webui",
						Ports:	[]corev1.ContainerPort{{
							ContainerPort: 	frontendPort,
							Name:			"visitors",
						}},
					}},
				},
			},
		},
	}

	controllerutil.SetControllerReference(v, dep, r.scheme)
	return dep
}

func (r *ReconcileVisitorsSite) frontendService(v *visitorsv1alpha1.VisitorsSite) *corev1.Service {
	labels := frontendLabels(v)

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:		v.Name + "-frontend-service",
			Namespace: 	v.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol: corev1.ProtocolTCP,
				Port: frontendPort,
				TargetPort: intstr.FromInt(frontendPort),
				NodePort: frontendServicePort,
			}},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	log.Info("Service Spec", "Service.Name", s.ObjectMeta.Name)

	controllerutil.SetControllerReference(v, s, r.scheme)
	return s
}

func frontendLabels(v *visitorsv1alpha1.VisitorsSite) map[string]string {
	return map[string]string{
		"app": "visitors",
		"visitorssite_cr": v.Name + "-frontend",
	}
}
