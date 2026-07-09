package models

import (
	"testing"
)

func TestBuildConfigValidateAllowsEmptyBranchAndTag(t *testing.T) {
	spec := BuildConfigSpec{
		SourceContext:        BuildContextSource{Git: &GitBuildSource{RepoURL: "https://github.com/org/repo"}},
		SourceRevision:       BuildSourceRevision{Git: &GitRevision{}},
		BuildImageRepository: BuildImageRepository{UseInClusterRegistry: true},
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}
