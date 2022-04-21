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
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	projectv1alpha1 "github.com/djkormo/go-project-operator/api/v1alpha1"
)

// ProjectNetworkPolicyReconciler reconciles a ProjectNetworkPolicy object
type ProjectNetworkPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=project.djkormo.github.io,resources=projectnetworkpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=project.djkormo.github.io,resources=projectnetworkpolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=project.djkormo.github.io,resources=projectnetworkpolicies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ProjectNetworkPolicy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ProjectNetworkPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	logger := log.Log.WithValues(req.Namespace, req.NamespacedName)

	logger.Info("Project operator Reconcile method starts for Network Policy...")

	// fetch the ProjectNetworkPolicy CR instance
	ProjectNetworkPolicy := &projectv1alpha1.ProjectNetworkPolicy{}
	err := r.Get(ctx, req.NamespacedName, ProjectNetworkPolicy)
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

	// fetch the ProjectNetworkPolicyTemplate CR instance
	ProjectNetworkPolicyTemplateFound := &projectv1alpha1.ProjectNetworkPolicyTemplateList{}
	// Listing all ProjectNetworkPolicyTemplates
	projNetpoltempErr := r.List(ctx, ProjectNetworkPolicyTemplateFound)
	if projNetpoltempErr != nil {
		if errors.IsNotFound(projNetpoltempErr) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("ProjectNetworkPolicyTemplate not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(projNetpoltempErr, "Failed to get Project Operator for NetworkPolicyTemplate")
		return ctrl.Result{}, projNetpoltempErr
	}

	// iterating all network policy instances
	for _, projectNetworkPolicytemplateItem := range ProjectNetworkPolicyTemplateFound.Items {

		// Fetch the ProjectNetworkPolicyTemplate instance
		projnetpoltempinstance := &projectv1alpha1.ProjectNetworkPolicyTemplate{}
		err = r.Get(ctx, types.NamespacedName{
			Name:      projectNetworkPolicytemplateItem.ObjectMeta.Name,
			Namespace: projectNetworkPolicytemplateItem.ObjectMeta.Namespace,
		}, projnetpoltempinstance)

		//logger.Info("Getting spec from ProjectNetworkPolicyTemplate", "projectNetworkPolicytemplateItem.Name", projectNetworkPolicytemplateItem.Name, "projectNetworkPolicytemplateItem.Namespace", projectNetworkPolicytemplateItem.Namespace)
		//logger.Info("Template policy spec:", "projectNetworkPolicytemplateItem.spec", projnetpoltempinstance.Spec.PolicySpec)

		// Get array of networkpolicie names
		netpolnames := ProjectNetworkPolicy.Spec.NetworkPolicies
		// Get namespace for of network policy
		netpolnamespace := ProjectNetworkPolicy.Spec.ProjectName
		// Find if network policy exists
		networkPolicyFound := &networkingv1.NetworkPolicy{}

		// Iterate through policy names
		for _, netpolname := range netpolnames {

			err = r.Get(ctx, types.NamespacedName{
				Name: netpolname, Namespace: netpolnamespace}, networkPolicyFound)
			if err != nil && errors.IsNotFound(err) {

				// Define new networkpolicy
				netpol := r.networkpolicyForProjectApp(ProjectNetworkPolicy, netpolname, projnetpoltempinstance) // networkpolicyForProjectApp() returns a network policy
				logger.Info("Creating a new NetworkPolicy", "NetworkPolicy.Name", netpolname, "NetworkPolicy.Namespace", netpol.Namespace)
				err = r.Create(ctx, netpol)
				//err = nil
				// in case of error
				if err != nil {
					logger.Error(err, "Failed to create new Network Policy", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
					return ctrl.Result{}, err
				}
				// in case of success
				logger.Info("NetworkPolicy created", "NetworkPolicy.Name", netpolname, "NetworkPolicy.Namespace", netpol.Namespace)
				return ctrl.Result{Requeue: true}, nil

			} else if err != nil {
				logger.Error(err, "Failed to get NetworkPolicy")
				// Reconcile failed due to error - requeue
				return ctrl.Result{}, err
			}

			netpol := r.networkpolicyForProjectApp(ProjectNetworkPolicy, netpolname, projnetpoltempinstance) // networkpolicyForProjectApp() returns a network policy
			labels := ProjectNetworkPolicy.GetLabels()
			annotations := ProjectNetworkPolicy.GetAnnotations()
			netpolspec := projnetpoltempinstance.Spec.PolicySpec

			netpol_unchanged_labels := IsMapSubset(ProjectNetworkPolicy.ObjectMeta.Labels, labels)
			netpol_unchanged_annotations := IsMapSubset(ProjectNetworkPolicy.ObjectMeta.Annotations, annotations)
			netpol_unchanged_spec := reflect.DeepEqual(networkPolicyFound.Spec, netpolspec)
			if !(netpol_unchanged_labels && netpol_unchanged_annotations && netpol_unchanged_spec && false) {

				logger.Info("Updating NetworkPolicy", "NetworkPolicy.Name", netpolname, "NetworkPolicy.Namespace", netpol.Namespace)

				if !netpol_unchanged_labels {
					logger.Info("Desired labels ", "Labels:", labels)
					logger.Info("Actual labels ", "Labels:", ProjectNetworkPolicy.ObjectMeta.Labels)
				}

				if !netpol_unchanged_annotations {

					logger.Info("Desired annotations", "Annotations:", annotations)
					logger.Info("Actual annotations", "Annotations:", ProjectNetworkPolicy.ObjectMeta.Annotations)
				}

				if !netpol_unchanged_spec {

					logger.Info("Desired spec", "Annotations:", netpolspec)
					logger.Info("Actual spec", "Annotations:", networkPolicyFound.Spec)
				}

				networkPolicyFound.ObjectMeta.Labels = labels
				networkPolicyFound.ObjectMeta.Annotations = annotations
				networkPolicyFound.Spec = netpolspec

				err = r.Update(ctx, netpol)
				// in case of error
				if err != nil {
					logger.Error(err, "Failed to Update NetworkPolicy", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
					return ctrl.Result{}, err
				}
				// in case of success
				logger.Info("NetworkPolicy updated", "NetworkPolicy.Name", netpolname, "NetworkPolicy.Namespace", netpol.Namespace)
				return ctrl.Result{Requeue: true}, nil

			} // of networkpolicy Update

		} // of for iterate through policy names

	} // of for  iterating all network policy instances

	return ctrl.Result{Requeue: true}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProjectNetworkPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&projectv1alpha1.ProjectNetworkPolicy{}).
		Complete(r)
}

func (r *ProjectNetworkPolicyReconciler) networkpolicyForProjectApp(m *projectv1alpha1.ProjectNetworkPolicy, name string, t *projectv1alpha1.ProjectNetworkPolicyTemplate) *networkingv1.NetworkPolicy {
	labels := m.GetLabels()
	annotations := m.GetAnnotations()
	namespace := m.Spec.ProjectName
	spec := t.Spec.PolicySpec

	networkpolicy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   namespace,
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: spec,
	}
	return networkpolicy
}