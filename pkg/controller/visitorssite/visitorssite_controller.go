package visitorssite

import (
	"context"

	visitorsv1alpha1 "github.com/jdob/visitors-operator/pkg/apis/visitors/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_visitorssite")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new VisitorsSite Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileVisitorsSite{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("visitorssite-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource VisitorsSite
	err = c.Watch(&source.Kind{Type: &visitorsv1alpha1.VisitorsSite{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &visitorsv1alpha1.VisitorsSite{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileVisitorsSite implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileVisitorsSite{}

// ReconcileVisitorsSite reconciles a VisitorsSite object
type ReconcileVisitorsSite struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a VisitorsSite object and makes changes based on the state read
// and what is in the VisitorsSite.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileVisitorsSite) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling VisitorsSite")

	// Fetch the VisitorsSite instance
	instance := &visitorsv1alpha1.VisitorsSite{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// == Visitors Service ==
	var result *reconcile.Result

	result, err = r.ensureBackendDeployment(request, instance)
	if result != nil {
		return *result, err
	}

	result, err = r.ensureBackendService(request, instance)
	if result != nil {
		return *result, err
	}

	// Everything went fine, don't requeue
	return reconcile.Result{}, nil
}

func (r *ReconcileVisitorsSite) ensureBackendDeployment(request reconcile.Request, instance *visitorsv1alpha1.VisitorsSite) (*reconcile.Result, error) {

	// See if deployment already exists and create if it doesn't
	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-backend", Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the deployment config
		dep := r.backendDeployment(instance)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)

		// Create the deployment
		err = r.client.Create(context.TODO(), dep)

		if err != nil {
			// Deployment failed
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return &reconcile.Result{}, err
		} else {
			// Deployment was successful
			return &reconcile.Result{Requeue: true}, nil
		}
	} else if err != nil {
		// Error that isn't due to the deployment not existing
		log.Error(err, "Failed to get Deployment")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *ReconcileVisitorsSite) ensureBackendService(request reconcile.Request, instance *visitorsv1alpha1.VisitorsSite) (*reconcile.Result, error) {
	found := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-backend", Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the service spec
		dep := r.backendService(instance)
		log.Info("Creating a new Service", "Service.Namespace", dep.Namespace, "Service.Name", dep.Name)

		// Create the service
		err = r.client.Create(context.TODO(), dep)

		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new Service", "Service.Namespace", dep.Namespace, "Service.Name", dep.Name)
			return &reconcile.Result{}, err
		} else {
			// Creation was successful
			return &reconcile.Result{Requeue: true}, nil
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		log.Error(err, "Failed to get Service")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *ReconcileVisitorsSite) backendDeployment(v *visitorsv1alpha1.VisitorsSite) *appsv1.Deployment {
	labels := map[string]string{"app": "visitors", "visitorssite_cr": v.Name + "-backend"}
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
						Image:	"jdob/visitors-service:latest",
						Name:	"visitors-service",
						Ports:	[]corev1.ContainerPort{{
							ContainerPort: 	8000,
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

func (r *ReconcileVisitorsSite) backendService(v *visitorsv1alpha1.VisitorsSite) *corev1.Service {
	labels := map[string]string{"app": "visitors", "visitorssite_cr": v.Name + "-backend"}

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:		v.Name + "-backend",
			Namespace: 	v.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol: corev1.ProtocolTCP,
				Port: 8000,
				TargetPort: intstr.FromInt(8000),
				NodePort: 30685,
			}},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	log.Info("Service Spec", "Service.Name", s.ObjectMeta.Name)

	controllerutil.SetControllerReference(v, s, r.scheme)
	return s
}

func (r *ReconcileVisitorsSite) webuiDeployment(v *visitorsv1alpha1.VisitorsSite) *appsv1.Deployment {
	labels := map[string]string{"app": "visitors", "visitorssite_cr": v.Name}
	size := int32(1)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:		v.Name,
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
							ContainerPort: 	3000,
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
