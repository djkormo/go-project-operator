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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	projectv1alpha1 "github.com/djkormo/go-project-operator/api/v1alpha1"
)

// ProjectAccessReconciler reconciles a ProjectAccess object
type ProjectAccessReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=project.djkormo.github.io,resources=projectaccesses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=project.djkormo.github.io,resources=projectaccesses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=project.djkormo.github.io,resources=projectaccesses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ProjectAccess object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ProjectAccessReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	logger := log.Log.WithValues(req.Namespace, req.NamespacedName)

	logger.Info("Reconcile Access method starts...")

	// fetch the ProjectAccess CR instance
	ProjectAccess := &projectv1alpha1.ProjectAccess{}
	err := r.Get(ctx, req.NamespacedName, ProjectAccess)

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("ProjectNetworkPolicy resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Project Operator instance for NetworkPolicy")
		return ctrl.Result{}, err
	}
	// exit if pause reconciliation label is set to true
	if v, ok := ProjectAccess.Labels[pauseReconciliationLabel]; ok && v == "true" {
		logger.Info("Not reconciling ProjectAccess: label", pauseReconciliationLabel, "is true")
		return ctrl.Result{}, nil
	}

	// fetch Project CR instance

	projectFound := &projectv1alpha1.Project{}
	projectName := ProjectAccess.Spec.ProjectName
	projectNamespace := ProjectAccess.ObjectMeta.Namespace
	// get  project template
	project_err := r.Get(ctx, types.NamespacedName{
		Name:      projectName,
		Namespace: projectNamespace,
	}, projectFound)

	// Checking if pause reconciliation label is set to true on Project LEVEL
	if project_err == nil {
		logger.Info("Found project:", projectName, ". That's ok")
		logger.Info("Project labels:", projectFound.Labels[pauseReconciliationLabel], " That's ok")
		if v, ok := projectFound.Labels[pauseReconciliationLabel]; ok && v == "true" {
			logger.Info("Not reconciling ProjectNetworkPolicy: label on Project level", pauseReconciliationLabel, "is true")
			return ctrl.Result{}, nil
		}
	}
	// Checking if Project object exists
	if project_err != nil {
		if errors.IsNotFound(project_err) {

			logger.Info("Project resource not found. Ignoring since primary object must be deleted:", projectName, projectNamespace)
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProjectAccessReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&projectv1alpha1.ProjectAccess{}).
		Complete(r)
}
