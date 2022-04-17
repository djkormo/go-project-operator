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
	"encoding/json"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fmt"

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

//+kubebuilder:rbac:groups=project.djkormo.github.io,namespace=project-operator,resources=projects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=project.djkormo.github.io,namespace=project-operator,resources=projects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=project.djkormo.github.io,namespace=project-operator,resources=projects/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update
//+kubebuilder:rbac:groups="",resources=resourcequotas,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=limitranges,verbs=get;list;watch;create;update;delete

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

	logger := log.Log.WithValues("Project", req.NamespacedName)

	logger.Info("Project operator Reconcile method starts...")

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
		logger.Info("Namespace created", "Namespace.Name", ns.Name)
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get Namespace")
		// Reconcile failed due to error - requeue
		return ctrl.Result{}, err
	}

	// This point, we have the namespace object created
	// Ensure the namespace labels and annotations are the same
	labels := Project.GetLabels()
	annotations := Project.GetAnnotations()
	resourcequotas := Project.Spec.ResourceQuota.Hard
	limits := Project.Spec.LimitRange.Limits

	ns_unchanged_labels := IsMapSubset(namespaceFound.ObjectMeta.Labels, labels)
	ns_unchanged_annotations := IsMapSubset(namespaceFound.ObjectMeta.Annotations, annotations)
	if !(ns_unchanged_labels && ns_unchanged_annotations) {
		logger.Info("Updating labels and annotation in namespace", "Name:", namespaceFound.Name)

		if !ns_unchanged_labels {
			logger.Info("Desired labels ", "Labels:", labels)
			logger.Info("Actual labels ", "Labels:", namespaceFound.ObjectMeta.Labels)
		}

		if !ns_unchanged_annotations {

			logger.Info("Desired annotations", "Annotations:", annotations)
			logger.Info("Actual annotations", "Annotations:", namespaceFound.ObjectMeta.Annotations)
		}
		namespaceFound.ObjectMeta.Labels = labels
		namespaceFound.ObjectMeta.Annotations = annotations
		err = r.Update(ctx, namespaceFound)
		if err != nil {
			logger.Error(err, "Failed to update Namespace", "Namespace.Namespace", namespaceFound.Namespace, "Namespace.Name", namespaceFound.Name)
			return ctrl.Result{}, err
		}
		// Spec updated return and requeue
		// Requeue for any reason other than an error
		return ctrl.Result{Requeue: true}, nil

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

	// This point, we have the resource quota object created
	// Ensure the resource quota specification is the same as in Project object
	// Ensure the project labels and annotations are the same as in Project object
	rq_unchanged_labels := IsMapSubset(resourceQuotaFound.ObjectMeta.Labels, labels)
	rq_unchanged_annotations := IsMapSubset(resourceQuotaFound.ObjectMeta.Annotations, annotations)

	rq_unchanged_spec := IsMapSubset(resourceQuotaFound.Spec.Hard, resourcequotas)
	if !(rq_unchanged_labels && rq_unchanged_annotations && rq_unchanged_spec) {
		logger.Info("Updating resourceQuota", "Name:", resourceQuotaFound.Name)

		if !rq_unchanged_labels {
			logger.Info("Desired labels ", "Labels:", labels)
			logger.Info("Actual labels ", "Labels:", resourceQuotaFound.ObjectMeta.Labels)
		}

		if !rq_unchanged_annotations {

			logger.Info("Desired annotations", "Annotations:", annotations)
			logger.Info("Actual annotations", "Annotations:", resourceQuotaFound.ObjectMeta.Annotations)
		}
		if !rq_unchanged_spec {

			logger.Info("Desired spec", "Annotations:", resourcequotas)
			logger.Info("Actual spec", "Annotations:", resourceQuotaFound.Spec.Hard)
		}
		resourceQuotaFound.ObjectMeta.Labels = labels
		resourceQuotaFound.ObjectMeta.Annotations = annotations
		resourceQuotaFound.Spec.Hard = resourcequotas
		err = r.Update(ctx, resourceQuotaFound)
		if err != nil {
			logger.Error(err, "Failed to update ResourceQuota", "ResourceQuota.Namespace", resourceQuotaFound.Namespace)
			return ctrl.Result{}, err
		}
		// Spec updated return and requeue
		// Requeue for any reason other than an error
		return ctrl.Result{Requeue: true}, nil

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

	// This point, we have the limit range object created
	// Ensure the limit range specification is the same as in Project object
	// Ensure the project labels and annotations are the same as in Project object

	lr_unchanged_labels := IsMapSubset(limitRangeFound.ObjectMeta.Labels, labels)
	lr_unchanged_annotations := IsMapSubset(limitRangeFound.ObjectMeta.Annotations, annotations)
	lr_unchanged_spec := reflect.DeepEqual(limitRangeFound.Spec.Limits, limits)

	if !(lr_unchanged_labels && lr_unchanged_annotations && lr_unchanged_spec) {
		logger.Info("Updating limitRange", "Name:", limitRangeFound.Name)

		if !lr_unchanged_labels {
			logger.Info("Desired labels ", "Labels:", labels)
			logger.Info("Actual labels ", "Labels:", limitRangeFound.ObjectMeta.Labels)
		}

		if !lr_unchanged_annotations {

			logger.Info("Desired annotations", "Annotations:", annotations)
			logger.Info("Actual annotations", "Annotations:", limitRangeFound.ObjectMeta.Annotations)
		}
		if !lr_unchanged_spec {

			logger.Info("Desired spec", "Annotations:", limits)
			logger.Info("Actual spec", "Annotations:", limitRangeFound.Spec.Limits)
		}
		limitRangeFound.ObjectMeta.Labels = labels
		limitRangeFound.ObjectMeta.Annotations = annotations
		limitRangeFound.Spec.Limits = limits
		err = r.Update(ctx, limitRangeFound)
		if err != nil {
			logger.Error(err, "Failed to update LimitRange", "LimitRange.Namespace", limitRangeFound.Namespace)
			return ctrl.Result{}, err
		}
		// Spec updated return and requeue
		// Requeue for any reason other than an error
		return ctrl.Result{Requeue: true}, nil

	}

	return ctrl.Result{Requeue: true}, nil
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
	resourceQuota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   m.Name,
			Name:        m.Name,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: m.Spec.ResourceQuota.Hard,
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
		Spec: corev1.LimitRangeSpec{
			Limits: m.Spec.LimitRange.Limits,
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

// https://stackoverflow.com/questions/67900919/check-if-a-map-is-subset-of-another-map
func IsMapSubset(mapSet interface{}, mapSubset interface{}) bool {

	mapSetValue := reflect.ValueOf(mapSet)
	mapSubsetValue := reflect.ValueOf(mapSubset)

	if fmt.Sprintf("%T", mapSet) != fmt.Sprintf("%T", mapSubset) {
		return false
	}

	if len(mapSetValue.MapKeys()) < len(mapSubsetValue.MapKeys()) {
		return false
	}

	if len(mapSubsetValue.MapKeys()) == 0 {
		return true
	}

	iterMapSubset := mapSubsetValue.MapRange()

	for iterMapSubset.Next() {
		k := iterMapSubset.Key()
		v := iterMapSubset.Value()

		value := mapSetValue.MapIndex(k)

		if !value.IsValid() || v.Interface() != value.Interface() {
			return false
		}
	}

	return true
}

//https://github.com/Myafq/limit-operator/blob/master/pkg/controller/clusterlimit/clusterlimit_controller.go

func areTheSame(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, ae := range a {
		if !includes(ae, b) {
			return false
		}
	}
	for _, be := range b {
		if !includes(be, a) {
			return false
		}
	}
	return true
}

func includes(a string, b []string) bool {
	for _, be := range b {
		if a == be {
			return true
		}
	}
	return false
}

//https://gosamples.dev/compare-slices/
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func AreEqualJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("error mashalling string 2 :: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}
