package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

type Condition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastTransitionTime time.Time `json:"last_transition_time"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
}

const LabelValueTrue = "true"

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

func FindCondition(conditions []Condition, condType string) *Condition {
	for i := range conditions {
		if conditions[i].Type == condType {
			return &conditions[i]
		}
	}
	return nil
}

func IsConditionTrue(conditions []Condition, condType string) bool {
	c := FindCondition(conditions, condType)
	return c != nil && c.Status == string(metav1.ConditionTrue)
}

func SetCondition(conditions *[]Condition, cond Condition) {
	for i := range *conditions {
		if (*conditions)[i].Type == cond.Type {
			(*conditions)[i] = cond
			return
		}
	}
	*conditions = append(*conditions, cond)
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
