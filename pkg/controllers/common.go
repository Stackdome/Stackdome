package controllers

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func WorkspaceStorageIdLabelPresentPredicate[T client.Object]() predicate.TypedFuncs[T] {
	return predicate.NewTypedPredicateFuncs(func(object T) bool {
		objectLabels := object.GetLabels()
		_, ok := objectLabels[models.WorkspaceStorageIDLabel]
		if !ok {
			return false
		}
		return true
	})
}
