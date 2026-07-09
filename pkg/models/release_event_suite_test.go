package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReleaseEventLinksRoundTrip(t *testing.T) {
	links := ReleaseEventLinks{{
		Kind:   ReleaseEventLinkKindBuildLogs,
		Label:  "View build logs",
		Target: map[string]string{"build_id": "b1", "resource_name": "api"},
	}}
	v, err := links.Value()
	require.NoError(t, err)

	var out ReleaseEventLinks
	require.NoError(t, out.Scan(v))
	require.Equal(t, links, out)
}

func TestReleaseEventMetadataRoundTrip(t *testing.T) {
	m := ReleaseEventMetadata{ReleaseEventMetaReason: "image not found"}
	v, err := m.Value()
	require.NoError(t, err)

	var out ReleaseEventMetadata
	require.NoError(t, out.Scan(v))
	require.Equal(t, m, out)
}
