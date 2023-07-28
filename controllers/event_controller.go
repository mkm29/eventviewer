/*
Copyright 2023.

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

	"github.com/ic2hrmk/promtail"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// EventReconciler reconciles a Event object
type EventReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	PromtailClient promtail.Client
	CommonLabels   map[string]string
}

//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=events/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Event object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *EventReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// get the event
	var event corev1.Event
	if err := r.Get(ctx, req.NamespacedName, &event); err != nil {
		logger.Error(err, "unable to fetch Event")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// construct map of labels for Loki
	extraLabels := map[string]string{
		"namespace": event.InvolvedObject.Namespace,
		"reason":    event.Reason,
		"type":      event.Type,
		"pod":       event.InvolvedObject.Name,
	}
	level := promtail.Info
	if event.Type != "Normal" {
		level = promtail.Warn
	}

	r.PromtailClient.LogfWithLabels(level, extraLabels, event.Note)

	logger.V(5).Info("processed event", "note", event.Note)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EventReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Event{}).
		Complete(r)
}
