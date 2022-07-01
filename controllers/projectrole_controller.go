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

	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	projectv1alpha1 "github.com/djkormo/go-project-operator/api/v1alpha1"
	helpers "github.com/djkormo/go-project-operator/helpers"
)

// ProjectRoleReconciler reconciles a ProjectRole object
type ProjectRoleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=project.djkormo.github.io,resources=projectroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=project.djkormo.github.io,resources=projectroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=project.djkormo.github.io,resources=projectroles/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ProjectRole object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ProjectRoleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	logger := log.Log.WithValues(req.Namespace, req.NamespacedName)
	logger.Info("Reconcile Role method starts...")

	// fetch the ProjectRole CR instance
	ProjectRole := &projectv1alpha1.ProjectRole{}
	err := r.Get(ctx, req.NamespacedName, ProjectRole)

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("ProjectRole resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Project Operator instance for ProjectRole")
		return ctrl.Result{}, err
	}
	// exit if pause reconciliation label is set to true
	if v, ok := ProjectRole.Labels[pauseReconciliationLabel]; ok && v == "true" {
		logger.Info("Not reconciling ProjectRole: label", pauseReconciliationLabel, "is true")
		return ctrl.Result{}, nil
	}

	// Get array of role names
	rolenames := ProjectRole.Spec.Roles
	logger.Info("List of Role names", "Role.Name", rolenames)

	// Get namespace for Role
	rolenamespace := ProjectRole.Spec.ProjectName
	logger.Info("Namespace for Role", "Role.Namespace", rolenamespace)

	// Find if Role exists
	roleFound := &v1.Role{}
	// Find if project role template exists
	projectRoleTemplateFound := &projectv1alpha1.ProjectRoleTemplate{}

	// Iterate through role names in ProjectRole
	for _, rolename := range rolenames {

		logger.Info("Checking state of Role", "Role.Name", rolename, "Role.Namespace", rolenamespace)

		// get role template
		roletemp_err := r.Get(ctx, types.NamespacedName{
			Name:      rolename,
			Namespace: ProjectRole.ObjectMeta.Namespace,
		}, projectRoleTemplateFound)

		if roletemp_err == nil {

			logger.Info("Role spec:", "Role.Spec", projectRoleTemplateFound.Spec, "Role.Name", projectRoleTemplateFound.Name)
			logger.Info("Role rules :", "Excluded namespaces", projectRoleTemplateFound.Spec.ExcludeNamespaces, "Role.Name", projectRoleTemplateFound.Name)

		}
		// RoleTemplate does not exist
		if roletemp_err != nil && errors.IsNotFound(roletemp_err) {
			logger.Info("Not existing ProjectRoleTemplate", "ProjectRoleTemplate.Name", rolename)
			logger.Info("Skipping creation of Role", "Role.Name", rolename)
			continue
			// skip the loop for rolename
		}

		err = r.Get(ctx, types.NamespacedName{
			Name: rolename, Namespace: rolenamespace}, roleFound)

		// CREATE ROLE LOGIC START
		if err != nil && errors.IsNotFound(err) {

			//Checking if namespace is exluded

			excluded_namespace := rolenamespace

			isExclude := false
			for _, e_namespace := range projectRoleTemplateFound.Spec.ExcludeNamespaces {
				if e_namespace == excluded_namespace {
					isExclude = true
				}
			}
			if isExclude {
				logger.Info("Skipping creating Role", "Excluded namespace", rolenamespace)
				continue
			}

			// checking if namespace is exluded

			// Define new Role
			role := r.roleForProjectApp(ProjectRole, rolename, projectRoleTemplateFound) // roleForProjectApp() returns a Role
			logger.Info("Creating a new Role", "Role.Name", rolename, "Role.Namespace", role.Namespace)
			err = r.Create(ctx, role)
			// in case of error
			if err != nil {
				logger.Error(err, "Failed to create new Role", "Role.Name", role.Name, "Role.Namespace", role.Namespace)
				return ctrl.Result{}, err
			}
			// in case of success
			logger.Info("Role created", "Role.Name", rolename, "Role.Namespace", role.Namespace)

		} else if err != nil {
			logger.Error(err, "Failed to get Role")
			// Reconcile failed due to error - requeue
			return ctrl.Result{}, err
		}

		// CREATE ROLE LOGIC END

		// UPDATE ROLE LOGIC START
		logger.Info("Update Role", "Role.Name", rolename, "Role.Namespace", rolenamespace)

		role := r.roleForProjectApp(ProjectRole, rolename, projectRoleTemplateFound) // roleForProjectApp() returns a Role
		labels := ProjectRole.GetLabels()
		annotations := ProjectRole.GetAnnotations()

		rolespec := projectRoleTemplateFound.Spec.RoleRules

		role_unchanged_labels := helpers.IsMapSubset(roleFound.ObjectMeta.Labels, labels)
		role_unchanged_annotations := helpers.IsMapSubset(roleFound.ObjectMeta.Annotations, annotations)
		role_unchanged_spec := false
		// https://github.com/kubernetes-sigs/kubebuilder/issues/592
		if equality.Semantic.DeepDerivative(rolespec, roleFound.Rules) {
			role_unchanged_spec = true
		}

		if !(role_unchanged_labels && role_unchanged_annotations && role_unchanged_spec) {
			logger.Info("Update Role changed", "Role.Name", rolename, "Role.Namespace", rolenamespace)
			logger.Info("Update Role changed", "Role.Name", rolename, "Role.role_unchanged_labels", role_unchanged_labels)
			logger.Info("Update Role changed", "Role.Name", rolename, "Role.role_unchanged_annotations", role_unchanged_annotations)
			logger.Info("Update Role changed", "Role.Name", rolename, "Role.role_unchanged_spec", role_unchanged_spec)
			if !role_unchanged_labels {
				logger.Info("Desired labels ", "Labels:", labels)
				logger.Info("Actual labels ", "Labels:", roleFound.ObjectMeta.Labels)
			}

			if !role_unchanged_annotations {

				logger.Info("Desired annotations", "Annotations:", annotations)
				logger.Info("Actual annotations", "Annotations:", roleFound.ObjectMeta.Annotations)
			}
			if !role_unchanged_spec {

				logger.Info("Desired spec", "Specification:", rolespec)
				logger.Info("Actual spec", "Specification:", roleFound.Rules)
			}

			role.ObjectMeta.Labels = labels
			role.ObjectMeta.Annotations = annotations
			role.Rules = rolespec
			err = r.Update(ctx, role)
			if err != nil {
				logger.Error(err, "Failed to update Role", "Role.Name", role.Name, "Role.Namespace", role.Namespace)
				return ctrl.Result{}, err
			}
		} else {
			logger.Info("Update Role unchanged", "Role.Name", rolename, "Role.Namespace", rolenamespace)

		}
		// UPDATE ROLE LOGIC END

	} //of for _, rolename := range rolenames

	return ctrl.Result{Requeue: true}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProjectRoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&projectv1alpha1.ProjectRole{}).
		Complete(r)
}

func (r *ProjectRoleReconciler) roleForProjectApp(m *projectv1alpha1.ProjectRole, name string, t *projectv1alpha1.ProjectRoleTemplate) *v1.Role {
	labels := m.GetLabels()
	annotations := m.GetAnnotations()
	namespace := m.Spec.ProjectName
	spec := t.Spec.RoleRules

	role := &v1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   namespace,
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
		},
		//Rule: spec,
		Rules: spec,
	}
	return role
}
