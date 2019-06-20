package visitorssite

import (
	"context"
	"time"

	visitorsv1alpha1 "github.com/jdob/visitors-operator/pkg/apis/visitors/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileVisitorsSite) mysqlDeployment(v *visitorsv1alpha1.VisitorsSite) *appsv1.Deployment {
	labels := labels(v, "mysql")
	size := int32(1)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:		"mysql",
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
						Image:	"mysql:5.7",
						Name:	"visitors-mysql",
						Ports:	[]corev1.ContainerPort{{
							ContainerPort: 	3306,
							Name:			"mysql",
						}},
						Env:	[]corev1.EnvVar{
							{
								Name:	"MYSQL_ROOT_PASSWORD",
								Value: 	"password",
							},
							{
								Name:	"MYSQL_DATABASE",
								Value:	"visitors",
							},
							{
								Name:	"MYSQL_USER",
								Value:	"visitors",
							},
							{
								Name:	"MYSQL_PASSWORD",
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

func (r *ReconcileVisitorsSite) mysqlService(v *visitorsv1alpha1.VisitorsSite) *corev1.Service {
	labels := labels(v, "mysql")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:		"mysql",
			Namespace: 	v.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: 	labels,
			Ports: 		[]corev1.ServicePort{{
				Port: 		3306,
			}},
			ClusterIP:	"None",
		},
	}

	controllerutil.SetControllerReference(v, s, r.scheme)
	return s
}

// Blocks until the MySQL deployment has finished
func (r *ReconcileVisitorsSite) waitForMysql(v *visitorsv1alpha1.VisitorsSite) (error) {
	deployment := &appsv1.Deployment{}
	err := wait.Poll(1*time.Second, 1*time.Minute,
		func() (done bool, err error) {
			err = r.client.Get(context.TODO(), types.NamespacedName{
				Name: "mysql",
				Namespace: v.Namespace,
				}, deployment)
			if err != nil {
				log.Error(err, "Deployment mysql not found")
				return false, nil
			}

			if deployment.Status.ReadyReplicas == 1 {
				log.Info("MySQL ready replica count met")
				return true, nil
			}

			log.Info("Waiting for MySQL to start")
			return false, nil
		},
	)
	return err
}