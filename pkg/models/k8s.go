package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Condition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastTransitionTime time.Time `json:"last_transition_time"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
}

type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Annotation struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Labels []Label
type Annotations []Annotation

func (l Labels) Value() (driver.Value, error) {
	return json.Marshal(l)
}

func (l *Labels) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte for Labels failed")
	}
	return json.Unmarshal(v, &l)
}

func (a Annotations) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Annotations) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte for Annotations failed")
	}
	return json.Unmarshal(v, &a)
}

func ConvertConditions(k8sconditions []metav1.Condition) []Condition {
	conditions := make([]Condition, len(k8sconditions))
	for i := range conditions {
		conditions[i] = Condition{
			Type:               k8sconditions[i].Type,
			Status:             string(k8sconditions[i].Status),
			LastTransitionTime: k8sconditions[i].LastTransitionTime.Time,
			Reason:             k8sconditions[i].Reason,
			Message:            k8sconditions[i].Message,
		}
	}
	return conditions
}
