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
	// Find resource quota
	resourceQuotaFound := &corev1.ResourceQuota{}
	err = r.Get(ctx, types.NamespacedName{Name: Project.Name, Namespace: Project.Name}, resourceQuotaFound)
	logger.Info("Looking for ResourceQuota")
	if err != nil && errors.IsNotFound(err) {
		logger.Error(err, "ResourceQuota not found. Trying to create")
		return ctrl.Result{}, nil
	}

	limitRangeFound := &corev1.LimitRange{}
	err = r.Get(ctx, types.NamespacedName{Name: Project.Name, Namespace: Project.Name}, limitRangeFound)
	logger.Info("Looking for LimitRange")
	if err != nil && errors.IsNotFound(err) {
		logger.Error(err, "LimitRange not found. Trying to create")
		return ctrl.Result{}, nil

	}

	return ctrl.Result{}, nil
}

func (r *ProjectReconciler) resourceQuotaForProject(m *projectv1alpha1.Project) *corev1.ResourceQuota {
	labels := m.GetLabels()
	annotations := m.GetAnnotations()
	resourceQuota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   m.Name,
			Name:        m.Name,
			Labels:      labels,
			Annotations: annotations,
		},
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
