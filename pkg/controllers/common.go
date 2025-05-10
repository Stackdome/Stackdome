package controllers

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func VolumeIdLabelPresentPredicate[T client.Object]() predicate.TypedFuncs[T] {
	return predicate.NewTypedPredicateFuncs(func(object T) bool {
		objectLabels := object.GetLabels()
		_, ok := objectLabels[models.VolumeIDLabel]
		return ok
	})
}

func DBObjectIDPresentPredicate[T client.Object](label string) predicate.TypedFuncs[T] {
	return predicate.NewTypedPredicateFuncs(func(object T) bool {
		objectLabels := object.GetLabels()
		_, ok := objectLabels[label]
		return ok
	})
}

func StackIDLabelPresentPredicate[T client.Object]() predicate.TypedFuncs[T] {
	return predicate.NewTypedPredicateFuncs(func(object T) bool {
		objectLabels := object.GetLabels()
		_, ok := objectLabels[models.StackIDLabel]
		return ok
	})
}
