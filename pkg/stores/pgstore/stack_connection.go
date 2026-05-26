package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
)

type StackConnectionStoreSpec struct {
	SessionFactory db.SessionFactory
}

type stackConnectionStore struct {
	sessionFactory db.SessionFactory
}

func NewStackConnectionStore(spec StackConnectionStoreSpec) stores.StackConnectionStore {
	return &stackConnectionStore{sessionFactory: spec.SessionFactory}
}

func (s *stackConnectionStore) ListByStackID(ctx context.Context, stackID string) (models.StackConnections, *errors.ServiceError) {
	var records []models.StackConnectionRecord
	if err := s.db(ctx).
		Where("stack_id = ?", stackID).
		Order("created_at ASC").
		Find(&records).Error; err != nil {
		return nil, errors.GeneralError("failed to list stack connections: %s", err.Error())
	}

	connections := make(models.StackConnections, 0, len(records))
	for _, record := range records {
		connections = append(connections, record.ToStackConnection())
	}
	return connections, nil
}

func (s *stackConnectionStore) CreateWithTx(ctx context.Context, stackID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}

	record := models.NewStackConnectionRecord(stackID, *connection)
	if err := tx.Create(&record).Error; err != nil {
		return nil, errors.GeneralError("failed to create stack connection: %s", err.Error())
	}

	created := record.ToStackConnection()
	return &created, nil
}

func (s *stackConnectionStore) UpdateWithTx(ctx context.Context, stackID, connectionID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}

	record := models.NewStackConnectionRecord(stackID, *connection)
	record.ID = connectionID
	result := tx.Model(&models.StackConnectionRecord{}).
		Where("stack_id = ? AND id = ?", stackID, connectionID).
		Updates(map[string]interface{}{
			"kind":     record.Kind,
			"from_ref": record.FromRef,
			"to_ref":   record.ToRef,
			"mappings": record.Mappings,
			"config":   record.Config,
		})
	if result.Error != nil {
		return nil, errors.GeneralError("failed to update stack connection: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, errors.NotFound("stack connection '%s' not found", connectionID)
	}

	updated := record.ToStackConnection()
	return &updated, nil
}

func (s *stackConnectionStore) DeleteWithTx(ctx context.Context, stackID, connectionID string) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}

	result := tx.Where("stack_id = ? AND id = ?", stackID, connectionID).Delete(&models.StackConnectionRecord{})
	if result.Error != nil {
		return errors.GeneralError("failed to delete stack connection: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("stack connection '%s' not found", connectionID)
	}
	return nil
}

func (s *stackConnectionStore) ReplaceByStackIDWithTx(ctx context.Context, stackID string, connections models.StackConnections) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}

	if err := tx.Where("stack_id = ?", stackID).Delete(&models.StackConnectionRecord{}).Error; err != nil {
		return errors.GeneralError("failed to delete existing stack connections: %s", err.Error())
	}
	if len(connections) == 0 {
		return nil
	}

	records := make([]models.StackConnectionRecord, 0, len(connections))
	for i := range connections {
		records = append(records, models.NewStackConnectionRecord(stackID, connections[i]))
	}

	if err := tx.Create(&records).Error; err != nil {
		return errors.GeneralError("failed to create stack connections: %s", err.Error())
	}
	return nil
}

func (s *stackConnectionStore) IsNodeReferenced(ctx context.Context, stackID string, ref models.TopologyNodeRef) (bool, error) {
	return s.isNodeReferenced(ctx, stackID, ref, "from_ref", "to_ref")
}

func (s *stackConnectionStore) IsNodeReferencedAsSource(ctx context.Context, stackID string, ref models.TopologyNodeRef) (bool, error) {
	return s.isNodeReferenced(ctx, stackID, ref, "from_ref")
}

func (s *stackConnectionStore) IsNodeReferencedAsTarget(ctx context.Context, stackID string, ref models.TopologyNodeRef) (bool, error) {
	return s.isNodeReferenced(ctx, stackID, ref, "to_ref")
}

func (s *stackConnectionStore) isNodeReferenced(ctx context.Context, stackID string, ref models.TopologyNodeRef, refColumns ...string) (bool, error) {
	if ref.Type == "" || (ref.Id == "" && ref.Name == "") || len(refColumns) == 0 {
		return false, nil
	}

	db := s.sessionFactory.New(ctx).Model(&models.StackConnectionRecord{})
	query := db.Where(s.nodeReferenceCondition(stackID, ref, refColumns[0]), s.nodeReferenceArgs(stackID, ref)...)
	for _, refColumn := range refColumns[1:] {
		query = query.Or(s.nodeReferenceCondition(stackID, ref, refColumn), s.nodeReferenceArgs(stackID, ref)...)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *stackConnectionStore) nodeReferenceCondition(stackID string, ref models.TopologyNodeRef, refColumn string) string {
	if ref.Id != "" {
		return refColumn + "->>'type' = ? AND " + refColumn + "->>'id' = ?"
	}
	if stackID != "" {
		return "stack_id = ? AND " + refColumn + "->>'type' = ? AND " + refColumn + "->>'name' = ?"
	}
	return refColumn + "->>'type' = ? AND " + refColumn + "->>'name' = ?"
}

func (s *stackConnectionStore) nodeReferenceArgs(stackID string, ref models.TopologyNodeRef) []interface{} {
	if ref.Id != "" {
		return []interface{}{ref.Type, ref.Id}
	}
	if stackID != "" {
		return []interface{}{stackID, ref.Type, ref.Name}
	}
	return []interface{}{ref.Type, ref.Name}
}

func (s *stackConnectionStore) db(ctx context.Context) *gorm.DB {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	return s.sessionFactory.New(ctx)
}
