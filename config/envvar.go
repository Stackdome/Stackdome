package config

import (
	"os"
	"strconv"
	"strings"
)

type EnvVar[T any] struct {
	Name         string
	Description  string
	DefaultValue *T
	Required     bool
	parse        func(string) (T, error)
}

func (e EnvVar[T]) Lookup() (T, bool) {
	raw, found := os.LookupEnv(e.Name)
	if !found {
		if e.DefaultValue != nil {
			return *e.DefaultValue, true
		}
		var zero T
		return zero, false
	}
	val, err := e.parse(strings.TrimSpace(raw))
	if err != nil {
		if e.DefaultValue != nil {
			return *e.DefaultValue, true
		}
		var zero T
		return zero, false
	}
	return val, true
}

func ptr[T any](v T) *T {
	return &v
}

func StringVar(name, description string, defaultVal *string, required bool) EnvVar[string] {
	return EnvVar[string]{
		Name:         name,
		Description:  description,
		DefaultValue: defaultVal,
		Required:     required,
		parse:        func(s string) (string, error) { return s, nil },
	}
}

func IntVar(name, description string, defaultVal *int, required bool) EnvVar[int] {
	return EnvVar[int]{
		Name:         name,
		Description:  description,
		DefaultValue: defaultVal,
		Required:     required,
		parse:        strconv.Atoi,
	}
}

func BoolVar(name, description string, defaultVal *bool, required bool) EnvVar[bool] {
	return EnvVar[bool]{
		Name:         name,
		Description:  description,
		DefaultValue: defaultVal,
		Required:     required,
		parse:        strconv.ParseBool,
	}
}
