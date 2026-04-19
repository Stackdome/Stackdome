package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Add condition helper methods (FindCondition, IsConditionTrue, SetCondition)
// and typed condition constants (see docs/plans/status-conditions-enhancement.md).
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

// Labels to map
func (l Labels) ToMap() map[string]string {
	m := make(map[string]string)
	for _, label := range l {
		m[label.Key] = label.Value
	}
	return m
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

// Annotations to map
func (a Annotations) ToMap() map[string]string {
	m := make(map[string]string)
	for _, annotation := range a {
		m[annotation.Key] = annotation.Value
	}
	return m
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
