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
	"reflect"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

	logger.Info("Project operator Reconcile Network Policy method starts...")

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

	// Get array of networkpolicie names
	netpolnames := ProjectNetworkPolicy.Spec.NetworkPolicies
	logger.Info("List of NetworkPolicy names", "NetworkPolicy.Names", netpolnames)

	// Get namespace for of network policy
	netpolnamespace := ProjectNetworkPolicy.Spec.ProjectName
	logger.Info("Namespace for NetworkPolicy", "NetworkPolicy.Namespace", netpolnamespace)

	// TODO

	// Find if network policy exists
	networkPolicyFound := &networkingv1.NetworkPolicy{}
	// Find if projetc network policy template exists
	projectNetworkPolicyTemplateFound := &projectv1alpha1.ProjectNetworkPolicyTemplate{}

	// Iterate through policy names
	for _, netpolname := range netpolnames {

		logger.Info("Checking NetworkPolicy", "NetworkPolicy.Name", netpolname, "NetworkPolicy.Namespace", netpolnamespace)

		// get network policy template
		netpoltemp_err := r.Get(ctx, types.NamespacedName{
			Name:      netpolname,
			Namespace: ProjectNetworkPolicy.ObjectMeta.Namespace,
		}, projectNetworkPolicyTemplateFound)

		if netpoltemp_err == nil {

			logger.Info("Network policy spec:", "NetworkPolicy", projectNetworkPolicyTemplateFound.Spec.PolicySpec, "NetworkPolicy.Name", projectNetworkPolicyTemplateFound.Name)
		}
		// NetworkPolicyTemplate does'not exist
		if netpoltemp_err != nil && errors.IsNotFound(netpoltemp_err) {
			logger.Info("Not existing ProjectNetworkPolicyTemplate", "ProjectNetworkPolicyTemplate.Name", netpolname)
			logger.Info("Skipping creation of NetworkPolicy", "NetworkPolicy.Name", netpolname)
			continue
			// skip the loop for netpolname
		}

		err = r.Get(ctx, types.NamespacedName{
			Name: netpolname, Namespace: netpolnamespace}, networkPolicyFound)

		if err != nil && errors.IsNotFound(err) {

			// Define new networkpolicy
			netpol := r.networkpolicyForProjectApp(ProjectNetworkPolicy, netpolname, projectNetworkPolicyTemplateFound) // networkpolicyForProjectApp() returns a network policy
			logger.Info("Creating a new NetworkPolicy", "NetworkPolicy.Name", netpolname, "NetworkPolicy.Namespace", netpol.Namespace)
			err = r.Create(ctx, netpol)
			// in case of error
			if err != nil {
				logger.Error(err, "Failed to create new NetworkPolicy", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
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

		// TODO UPDATE LOGIC
		logger.Info("Update NetworkPolicy", "NetworkPolicy.Name", netpolname, "NetworkPolicy.Namespace", netpolnamespace)

		netpol := r.networkpolicyForProjectApp(ProjectNetworkPolicy, netpolname, projectNetworkPolicyTemplateFound) // networkpolicyForProjectApp() returns a NetworkPolicy
		labels := ProjectNetworkPolicy.GetLabels()
		annotations := ProjectNetworkPolicy.GetAnnotations()
		// TODO not working good
		netpolspec := projectNetworkPolicyTemplateFound.Spec.PolicySpec

		netpol_unchanged_labels := IsMapSubset(networkPolicyFound.ObjectMeta.Labels, labels)
		netpol_unchanged_annotations := IsMapSubset(networkPolicyFound.ObjectMeta.Annotations, annotations)
		netpol_unchanged_spec := reflect.DeepEqual(netpolspec, networkPolicyFound.Spec)

		if !(netpol_unchanged_labels && netpol_unchanged_annotations && netpol_unchanged_spec) {
			logger.Info("Update NetworkPolicy changed", "NetworkPolicy.Name", netpolname, "NetworkPolicy.Namespace", netpolnamespace)
			logger.Info("Update NetworkPolicy changed", "NetworkPolicy.Name", netpolname, "NetworkPolicy.netpol_unchanged_labels", netpol_unchanged_labels)
			logger.Info("Update NetworkPolicy changed", "NetworkPolicy.Name", netpolname, "NetworkPolicy.netpol_unchanged_annotations", netpol_unchanged_annotations)
			logger.Info("TODO Update NetworkPolicy changed", "NetworkPolicy.Name", netpolname, "NetworkPolicy.netpol_unchanged_spec", netpol_unchanged_spec)
			if !netpol_unchanged_labels {
				logger.Info("Desired labels ", "Labels:", labels)
				logger.Info("Actual labels ", "Labels:", networkPolicyFound.ObjectMeta.Labels)
			}

			if !netpol_unchanged_annotations {

				logger.Info("Desired annotations", "Annotations:", annotations)
				logger.Info("Actual annotations", "Annotations:", networkPolicyFound.ObjectMeta.Annotations)
			}
			if !netpol_unchanged_spec {

				logger.Info("TODO Desired spec", "Specification:", netpolspec)
				logger.Info("TODO Actual spec", "Specification:", networkPolicyFound.Spec)
			}

			netpol.ObjectMeta.Labels = labels
			netpol.ObjectMeta.Annotations = annotations
			netpol.Spec = netpolspec
			err = r.Update(ctx, netpol)
			if err != nil {
				logger.Error(err, "Failed to update NetworkPolicy", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
				return ctrl.Result{}, err
			}
		} else {
			logger.Info("TODO Update NetworkPolicy unchanged", "NetworkPolicy.Name", netpolname, "NetworkPolicy.Namespace", netpolnamespace)
			return ctrl.Result{Requeue: true}, nil

		}

		//return ctrl.Result{Requeue: true}, nil

	} //of or _, netpolname := range netpolnames

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
