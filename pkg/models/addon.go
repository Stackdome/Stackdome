package models

type Addon interface {
	Type() string
	AddonName() string
}
