package workspacestorage

// const (
// 	controllerName = "stack-storage-controller"
// )

// type stackStorageReconciler struct {
// 	Client         client.Client
// 	Log            logger.Logger
// 	StorageService services.StackStorageService
// 	Env            string
// }

// type StackStorageReconcilerSpec struct {
// 	Log            logger.Logger
// 	StorageService services.StackStorageService
// 	Env            string
// }

// func NewStackStorageReconciler(spec StackStorageReconcilerSpec) *stackStorageReconciler {
// 	return &stackStorageReconciler{
// 		Log:            spec.Log,
// 		StorageService: spec.StorageService,
// 		Env:            spec.Env,
// 	}
// }

// func (w *stackStorageReconciler) AddToManager(manager manager.Manager) error {
// 	w.Client = manager.GetClient()
// 	controller, err := controller.New(controllerName, manager, controller.Options{
// 		Reconciler: w,
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	src := source.Kind(
// 		manager.GetCache(),
// 		&storagev1alpha1.Storage{},
// 		&handler.TypedEnqueueRequestForObject[*storagev1alpha1.Storage]{},
// 		controllers.StackStorageIdLabelPresentPredicate[*storagev1alpha1.Storage](),
// 	)

// 	return controller.Watch(src)
// }

// func (w *stackStorageReconciler) Name() string {
// 	return controllerName
// }

// func (r *stackStorageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
// 	r.Log.Infof("reconciling stack storage: %s in namespace %s", req.Name, req.Namespace)

// 	storageCr := &storagev1alpha1.Storage{}
// 	err := r.Client.Get(ctx, req.NamespacedName, storageCr)
// 	if err != nil {
// 		if errors.IsNotFound(err) {
// 			return ctrl.Result{}, nil
// 		}
// 		return ctrl.Result{}, err
// 	}
// 	stackStorageID, ok := storageCr.Labels[models.StackStorageIDLabel]
// 	if !ok {
// 		r.Log.Errorf("stack storage %s in namespace %s does not have a stack storage id label", storageCr.Name, storageCr.Namespace)
// 		return ctrl.Result{}, nil
// 	}
// 	dbStackStorage, serr := r.StorageService.InternalGet(ctx, stackStorageID)
// 	if serr != nil {
// 		if serr.Code == apperrors.ErrorNotFound {
// 			r.Log.Infof("stack storage %s in namespace %s not found in DB", storageCr.Name, storageCr.Namespace)
// 			return ctrl.Result{Requeue: true}, nil
// 		}
// 		return ctrl.Result{}, fmt.Errorf("failed to get stack storage %s in namespace %s: %w from DB", storageCr.Name, storageCr.Namespace, serr)
// 	}

// 	objectHashChanged := dbStackStorage.Status == nil || (storageCr.Status.StatusHash != dbStackStorage.Status.LastObservedStatusHash)
// 	if objectHashChanged {
// 		dbStackStorage.Status = mapClusterObjStatusToDBObjStatus(storageCr)
// 		serr = r.StorageService.UpdateStatus(ctx, stackStorageID, dbStackStorage.Status)
// 		if serr != nil {
// 			return ctrl.Result{}, fmt.Errorf("failed to update stack storage %s in namespace %s: %w from DB", storageCr.Name, storageCr.Namespace, serr)
// 		}
// 		return ctrl.Result{}, nil
// 	}
// 	return ctrl.Result{}, nil
// }

// func mapClusterObjStatusToDBObjStatus(clusterInstance *storagev1alpha1.Storage) *models.StackStorageStatus {
// 	clusterObjectStatus := clusterInstance.Status
// 	return &models.StackStorageStatus{
// 		ObservedVersion: clusterObjectStatus.ObservedStackdomeServerObjectGeneration,
// 		// Use phase as state, why do we need to have phase separately in the DB?.
// 		State:                    dbObjStateFromClusterObj(clusterInstance),
// 		Phase:                    string(clusterObjectStatus.Phase),
// 		Conditions:               models.ConvertConditions(clusterObjectStatus.Conditions),
// 		StorageServerServiceName: clusterObjectStatus.ServiceName,
// 		LastObservedStatusHash:   clusterObjectStatus.StatusHash,
// 	}
// }

// func dbObjStateFromClusterObj(clusterInstance *storagev1alpha1.Storage) models.StackStorageState {
// 	availableCondition := meta.FindStatusCondition(clusterInstance.Status.Conditions, string(storagev1alpha1.StorageAvailable))
// 	failedCondition := meta.FindStatusCondition(clusterInstance.Status.Conditions, string(storagev1alpha1.StorageFailed))
// 	switch {
// 	case availableCondition == nil:
// 		return models.StackStorageStatePending
// 	case availableCondition.Status == metav1.ConditionTrue:
// 		return models.StackStorageStateReady
// 	case availableCondition.Status == metav1.ConditionFalse:
// 		return models.StackStorageStateCreating
// 	case failedCondition != nil && failedCondition.Status == metav1.ConditionTrue:
// 		return models.StackStorageStateFailed
// 	default:
// 		return models.StackStorageStatePending
// 	}
// }
