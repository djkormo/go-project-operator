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
	//	"encoding/json"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	projectv1alpha1 "github.com/djkormo/go-project-operator/api/v1alpha1"
	helpers "github.com/djkormo/go-project-operator/helpers"
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
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;
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

	logger := log.Log.WithValues(req.Namespace, req.NamespacedName)

	logger.Info("Reconcile Project method starts...")

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

	// exit if pause reconciliation label is set to true
	if v, ok := Project.Labels[pauseReconciliationLabel]; ok && v == "true" {
		logger.Info("Not reconciling Project: label", pauseReconciliationLabel, "is true")

		return ctrl.Result{}, nil
	}

	// creating missing override configmap

	configMapFound := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: configMapName, Namespace: Project.Namespace}, configMapFound)

	if err != nil && errors.IsNotFound(err) {
		// define a new configmap
		cm := r.populateConfigOverrideConfigMap(Project) // returns a configmap
		logger.Info("Creating a new Configmap", "ConfigMap.Name", configMapName)
		err = r.Create(ctx, cm)
		if err != nil {
			logger.Error(err, "Failed to create new ConfigMap", "Configmap.Name", configMapName)
			return ctrl.Result{}, err
		}
		// configmap created, return and requeue
		logger.Info("ConfigMap created", "ConfiMap.Name", configMapName)
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get ConfigMap")
		// Reconcile failed due to error - requeue
		return ctrl.Result{}, err
	}

	// if configmap exists, use its data
	// TODO

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

	ns_unchanged_labels := helpers.IsMapSubset(namespaceFound.ObjectMeta.Labels, labels)
	ns_unchanged_annotations := helpers.IsMapSubset(namespaceFound.ObjectMeta.Annotations, annotations)
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
	rq_unchanged_labels := helpers.IsMapSubset(resourceQuotaFound.ObjectMeta.Labels, labels)
	rq_unchanged_annotations := helpers.IsMapSubset(resourceQuotaFound.ObjectMeta.Annotations, annotations)

	rq_unchanged_spec := helpers.IsMapSubset(resourceQuotaFound.Spec.Hard, resourcequotas)
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

	lr_unchanged_labels := helpers.IsMapSubset(limitRangeFound.ObjectMeta.Labels, labels)
	lr_unchanged_annotations := helpers.IsMapSubset(limitRangeFound.ObjectMeta.Annotations, annotations)
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

			logger.Info("Desired spec", "Specification:", limits)
			logger.Info("Actual spec", "Specification:", limitRangeFound.Spec.Limits)
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
		WithEventFilter(ignoreDeletionPredicate()).
		Complete(r)
}

type UpdateEvent struct {
	// ObjectOld is the object from the event.
	ObjectOld runtime.Object

	// ObjectNew is the object from the event.
	ObjectNew runtime.Object
}

// Predicate filters events before enqueuing the keys.
type Predicate interface {
	Create(event.CreateEvent) bool
	Delete(event.DeleteEvent) bool
	Update(event.UpdateEvent) bool
	Generic(event.GenericEvent) bool
}

// Funcs implements Predicate.
type Funcs struct {
	CreateFunc  func(event.CreateEvent) bool
	DeleteFunc  func(event.DeleteEvent) bool
	UpdateFunc  func(event.UpdateEvent) bool
	GenericFunc func(event.GenericEvent) bool
}

func ignoreDeletionPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
}

func (r *ProjectReconciler) populateConfigOverrideConfigMap(m *projectv1alpha1.Project) *corev1.ConfigMap {

	placeholderConfig := map[string]string{
		"pauseReconciliationLabel": pauseReconciliationLabel,
	}
	labels := map[string]string{
		pauseReconciliationLabel: "false",
	}
	namespace := m.Namespace
	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
			Labels:    labels,
		},
		Data: placeholderConfig,
	}

	return configmap
}
