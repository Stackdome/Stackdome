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
	var connections models.StackConnections
	if err := s.db(ctx).
		Where("stack_id = ?", stackID).
		Order("created_at ASC").
		Find(&connections).Error; err != nil {
		return nil, errors.GeneralError("failed to list stack connections: %s", err.Error())
	}
	return connections, nil
}

func (s *stackConnectionStore) CreateWithTx(ctx context.Context, stackID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}

	connection.StackID = stackID
	if err := tx.Create(connection).Error; err != nil {
		return nil, errors.GeneralError("failed to create stack connection: %s", err.Error())
	}
	return connection, nil
}

func (s *stackConnectionStore) UpdateWithTx(ctx context.Context, stackID, connectionID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}

	connection.StackID = stackID
	connection.ID = connectionID
	result := tx.Model(&models.StackConnection{}).
		Where("stack_id = ? AND id = ?", stackID, connectionID).
		Updates(map[string]interface{}{
			"kind":     connection.Kind,
			"from_ref": connection.From,
			"to_ref":   connection.To,
			"mappings": connection.Mappings,
			"config":   connection.Config,
		})
	if result.Error != nil {
		return nil, errors.GeneralError("failed to update stack connection: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, errors.NotFound("stack connection '%s' not found", connectionID)
	}
	return connection, nil
}

func (s *stackConnectionStore) DeleteWithTx(ctx context.Context, stackID, connectionID string) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}

	result := tx.Where("stack_id = ? AND id = ?", stackID, connectionID).Delete(&models.StackConnection{})
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

	existing, err := s.ListByStackID(ctx, stackID)
	if err != nil {
		return err
	}

	desiredByID := make(map[string]models.StackConnection, len(connections))
	for i := range connections {
		connections[i].StackID = stackID
		if connections[i].ID != "" {
			desiredByID[connections[i].ID] = connections[i]
		}
	}

	existingByID := make(map[string]struct{}, len(existing))
	for _, e := range existing {
		existingByID[e.ID] = struct{}{}
		if _, keep := desiredByID[e.ID]; !keep {
			if delErr := tx.Where("stack_id = ? AND id = ?", stackID, e.ID).Delete(&models.StackConnection{}).Error; delErr != nil {
				return errors.GeneralError("failed to delete stack connection '%s': %s", e.ID, delErr.Error())
			}
		}
	}

	for i := range connections {
		c := &connections[i]
		if c.ID != "" {
			if _, exists := existingByID[c.ID]; exists {
				if updErr := tx.Model(&models.StackConnection{}).
					Where("stack_id = ? AND id = ?", stackID, c.ID).
					Updates(map[string]interface{}{
						"kind":     c.Kind,
						"from_ref": c.From,
						"to_ref":   c.To,
						"mappings": c.Mappings,
						"config":   c.Config,
					}).Error; updErr != nil {
					return errors.GeneralError("failed to update stack connection '%s': %s", c.ID, updErr.Error())
				}
				continue
			}
		}
		if createErr := tx.Create(c).Error; createErr != nil {
			return errors.GeneralError("failed to create stack connection: %s", createErr.Error())
		}
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

	db := s.sessionFactory.New(ctx).Model(&models.StackConnection{})
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
