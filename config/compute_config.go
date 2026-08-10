package config

import "fmt"

type ComputeMode string

const (
	ComputeModeBYOC   ComputeMode = "bring_your_own"
	ComputeModeShared ComputeMode = "shared"
)

func (c *ApplicationConfig) validateComputeMode() error {
	if c.ComputeMode != ComputeModeBYOC && c.ComputeMode != ComputeModeShared {
		return fmt.Errorf("compute mode must be %q or %q", ComputeModeBYOC, ComputeModeShared)
	}
	if c.IsStackdomeCloud() && !c.UsesSharedCompute() {
		return fmt.Errorf("compute mode must be %q in %q runtime mode", ComputeModeShared, RuntimeModeStackdomeCloud)
	}
	return nil
}

func (c *ApplicationConfig) UsesSharedCompute() bool {
	return c.ComputeMode == ComputeModeShared
}
