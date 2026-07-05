package validation

import (
	"testing"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

// Regression: ValidateVolume previously reflected on fields that no longer
// exist on the generated openapi.Volume (OrganisationId, WorkspaceName,
// Namespace). reflect.FieldByName on a missing field yields an invalid Value
// whose String() is "<invalid Value>", so validateEmpty failed for EVERY
// payload with "organisation_id must be empty".
func TestValidateVolume(t *testing.T) {
	minimal := func() openapi.Volume {
		return openapi.Volume{
			Name: "tj-data",
			Spec: openapi.VolumeSpec{
				Size:       "1Gi",
				AccessMode: "ReadWriteOnce",
			},
		}
	}

	t.Run("accepts a minimal client payload", func(t *testing.T) {
		v := minimal()
		assert.Nil(t, ValidateVolume(&v)())
	})

	t.Run("rejects a client-supplied id", func(t *testing.T) {
		v := minimal()
		v.Id = ptr.To("some-id")
		err := ValidateVolume(&v)()
		assert.NotNil(t, err)
		assert.Contains(t, err.Reason, "id must be empty")
	})

	t.Run("rejects a missing name", func(t *testing.T) {
		v := minimal()
		v.Name = ""
		assert.NotNil(t, ValidateVolume(&v)())
	})

	t.Run("rejects an invalid size quantity", func(t *testing.T) {
		v := minimal()
		v.Spec.Size = "not-a-size"
		err := ValidateVolume(&v)()
		assert.NotNil(t, err)
		assert.Contains(t, err.Reason, "spec.size is not a valid quantity")
	})
}
