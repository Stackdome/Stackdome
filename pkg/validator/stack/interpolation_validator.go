package stack

import (
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	interpolation "stackdome.io/cluster-agent/pkg/interpolation"
)

type interpolationValidation struct{}

func NewInterpolationValidation() *interpolationValidation {
	return &interpolationValidation{}
}

func (i *interpolationValidation) ValidateStackInterpolations(in *models.Stack) error {
	interpolationCtx := newValidationInterpolationContext(in)
	interpolator := interpolation.NewInterpolator(&interpolationCtx)

	for _, sr := range in.StackResources {
		if sr.ExecutionConfig != nil && len(sr.ExecutionConfig.Env) > 0 {
			for _, env := range sr.ExecutionConfig.Env {
				if env.Value != "" {
					_, err := interpolator.InterpolateString(env.Value)
					if err != nil {
						return fmt.Errorf(
							"failed to interpolate env '%s' value %s for resource %s: %w", env.Name, env.Value, sr.Name, err,
						)
					}
				}
			}
		}
	}
	return nil
}

func newValidationInterpolationContext(stack *models.Stack) interpolation.InterpolationContext {
	res := interpolation.InterpolationContext{
		Resources: make(map[string]interpolation.ResourceContext),
	}
	for _, sr := range stack.StackResources {
		publicIngressCtx := make([]interpolation.IngressContext, 0)
		currResourceCtx := interpolation.ResourceContext{
			Name: sr.Name,
			Status: interpolation.ResourceStatus{
				InternalService: &sr.Name,
			},
		}
		if len(sr.Ports) > 0 {
			for _, port := range sr.Ports {
				if port.ExposedToPublic {
					publicIngressCtx = append(publicIngressCtx, interpolation.IngressContext{
						ExposedToPublic: port.ExposedToPublic,
						TargetPort:      port.Number,
						URL:             "validation_url",
					})
				}
			}
		}
		currResourceCtx.Status.PublicIngresses = publicIngressCtx
		res.Resources[sr.Name] = currResourceCtx
	}
	return res
}
