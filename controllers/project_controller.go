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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	projectv1alpha1 "github.com/djkormo/go-project-operator/api/v1alpha1"
)

// ProjectReconciler reconciles a Project object
type ProjectReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=project.djkormo.github.io,resources=projects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=project.djkormo.github.io,resources=projects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=project.djkormo.github.io,resources=projects/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Project object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile

func (r *ProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	logger := log.Log.WithValues("Project", req.NamespacedName)

	logger.Info("Project operator Reconcile method...")

	// fetch the Project CR instance
	Project := &projectv1alpha1.Project{}
	err := r.Get(ctx, req.NamespacedName, Project)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("Project resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Project Operator instance")
		return ctrl.Result{}, err
	}

	// Find if namespace exists
	namespaceFound := &corev1.Namespace{}
	err = r.Get(ctx, types.NamespacedName{Name: Project.Name}, namespaceFound)

	if err != nil && errors.IsNotFound(err) {
		// define a new namespace
		ns := r.namespaceForProjectApp(Project) // namespaceForProjectApp() returns a namespace
		logger.Info("Creating a new Namespace", "Namespace.Name", ns.Name)
		err = r.Create(ctx, ns)
		if err != nil {
			logger.Error(err, "Failed to create new Namespace", "Namespace.Name", ns.Name)
			return ctrl.Result{}, err
		}
		// namespace created, return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get Namespace")
		// Reconcile failed due to error - requeue
		return ctrl.Result{}, err
	}

	// Find if resourcequota exists
	resourceQuotaFound := &corev1.ResourceQuota{}
	err = r.Get(ctx, types.NamespacedName{Name: Project.Name, Namespace: Project.Name}, resourceQuotaFound)

	if err != nil && errors.IsNotFound(err) {
		// define a new resourcequota
		rq := r.resourceQuotaForProject(Project) // resourceQuotaForProject() returns a resourcequota
		logger.Info("Creating a new ResourceQuota", "ResourceQuota.Namespace", rq.Namespace, "ResourceQuota.Name", rq.Name)
		err = r.Create(ctx, rq)
		if err != nil {
			logger.Error(err, "Failed to create new ResourceQuota", "ResourceQuota.Namespace", rq.Namespace, "ResourceQuota.Name", rq.Name)
			return ctrl.Result{}, err
		}
		// resourcequota created, return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get ResourceQuota")
		// Reconcile failed due to error - requeue
		return ctrl.Result{}, err
	}
	// Find if limitrange exists
	limitRangeFound := &corev1.LimitRange{}
	err = r.Get(ctx, types.NamespacedName{Name: Project.Name, Namespace: Project.Name}, limitRangeFound)
	if err != nil && errors.IsNotFound(err) {
		// define a new resourcequota
		lr := r.limitRangeForProjectApp(Project) // limitRangeForProjectApp() returns a limitrange
		logger.Info("Creating a new LimitRange", "LimitRange.Namespace", lr.Namespace, "LimitRange.Name", lr.Name)
		err = r.Create(ctx, lr)
		if err != nil {
			logger.Error(err, "Failed to create new LimitRange", "LimitRange.Namespace", lr.Namespace, "LimitRange.Name", lr.Name)
			return ctrl.Result{}, err
		}
		// limitrange created, return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get LimitRange")
		// Reconcile failed due to error - requeue
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ProjectReconciler) namespaceForProjectApp(m *projectv1alpha1.Project) *corev1.Namespace {
	labels := m.GetLabels()
	annotations := m.GetAnnotations()
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   m.Name,
			Name:        m.Name,
			Labels:      labels,
			Annotations: annotations,
		},
	}
	return namespace
}

func (r *ProjectReconciler) resourceQuotaForProject(m *projectv1alpha1.Project) *corev1.ResourceQuota {
	labels := m.GetLabels()
	annotations := m.GetAnnotations()
	//hard:=m.spec.hard()
	resourceQuota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   m.Name,
			Name:        m.Name,
			Labels:      labels,
			Annotations: annotations,
		},
		//		Spec: corev1.ResourceQuotaSpec {
		//			Hard: {
		//				requests.cpu:    "1",
		//				requests.memory: "2",
		//				limits.cpu:      "1",
		//				limits.memory:   "2",
		//			},
		//		},
	}
	return resourceQuota
}

func (r *ProjectReconciler) limitRangeForProjectApp(m *projectv1alpha1.Project) *corev1.LimitRange {
	labels := m.GetLabels()
	annotations := m.GetAnnotations()
	limitRange := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   m.Name,
			Name:        m.Name,
			Labels:      labels,
			Annotations: annotations,
		},
	}
	return limitRange
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&projectv1alpha1.Project{}).
		Complete(r)
}
