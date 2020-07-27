package util

import (
	"reflect"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ComposePredicates will compose a variable number of predicte and return a predicate that
// will allow events that are allowed by any of the given predicates.
func ComposePredicates(predicates ...predicate.Predicate) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			for _, p := range predicates {
				if p != nil && p.Create(e) {
					return true
				}
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			for _, p := range predicates {
				if p != nil && p.Delete(e) {
					return true
				}
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			for _, p := range predicates {
				if p != nil && p.Update(e) {
					return true
				}
			}
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			for _, p := range predicates {
				if p != nil && p.Generic(e) {
					return true
				}
			}
			return false
		},
	}
}

// LabelsChangedPredicate will only allow events that changed Metadata.Labels
type LabelsChangedPredicate struct {
	predicate.Funcs
}

// Update implements the update event trap for LabelsChangedPredicate
func (p LabelsChangedPredicate) Update(e event.UpdateEvent) bool {
	return e.MetaOld != nil &&
		e.MetaNew != nil &&
		!reflect.DeepEqual(e.MetaOld.GetLabels(), e.MetaNew.GetLabels())
}

// FinalizersChangedPredicate will only allow events that changed Metadata.Finalizers
type FinalizersChangedPredicate struct {
	predicate.Funcs
}

// Update implements the update event trap for FinalizersChangedPredicate
func (p FinalizersChangedPredicate) Update(e event.UpdateEvent) bool {
	return e.MetaOld != nil &&
		e.MetaNew != nil &&
		!reflect.DeepEqual(e.MetaOld.GetFinalizers(), e.MetaNew.GetFinalizers())
}

// LogEventsPredicate will passthrough events while loging a message for each
type LogEventsPredicate struct {
}

// Create implements the create event trap for LogEventsPredicate
func (p LogEventsPredicate) Create(e event.CreateEvent) bool {
	if e.Meta != nil {
		logrus.Infof("Create event detected for %s (%s), queuing Reconcile",
			e.Meta.GetName(), e.Meta.GetNamespace())
	}
	return true
}

// Delete implements the delete event trap for LogEventsPredicate
func (p LogEventsPredicate) Delete(e event.DeleteEvent) bool {
	if e.Meta != nil {
		logrus.Infof("Delete event detected for %s (%s), queuing Reconcile",
			e.Meta.GetName(), e.Meta.GetNamespace())
	}
	return true
}

// Update implements the update event trap for LogEventsPredicate
func (p LogEventsPredicate) Update(e event.UpdateEvent) bool {
	if e.MetaOld != nil {
		logrus.Infof("Update event detected for %s (%s), queuing Reconcile",
			e.MetaOld.GetName(), e.MetaOld.GetNamespace())
	}
	return true
}

// Generic implements the generic event trap for LogEventsPredicate
func (p LogEventsPredicate) Generic(e event.GenericEvent) bool {
	if e.Meta != nil {
		logrus.Infof("Generic event detected for %s (%s), queuing Reconcile",
			e.Meta.GetName(), e.Meta.GetNamespace())
	}
	return true
}
