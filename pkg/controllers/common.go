package controllers

import (
	"github.com/Stackdome/stackdome/pkg/models"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
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
		_, ok := objectLabels[corev1alpha1.LabelStackID]
		return ok
	})
}

func ClusterImageRegistryIDLabelPresentPredicate[T client.Object]() predicate.TypedFuncs[T] {
	return predicate.NewTypedPredicateFuncs(func(object T) bool {
		objectLabels := object.GetLabels()
		_, ok := objectLabels[models.ImageRegistryIDLabel]
		return ok
	})
}
