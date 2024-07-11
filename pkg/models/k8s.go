package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Condition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	ObservedGeneration int32     `json:"observed_generation"`
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
