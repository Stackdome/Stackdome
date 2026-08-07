package main

import (
	"reflect"
	"testing"
)

func TestKubernetesServiceQueryUsesDefaultNamespace(t *testing.T) {
	want := []string{
		"get", "svc", "kubernetes",
		"-n", "default",
		"-o", "jsonpath={.spec.clusterIP}",
	}

	if got := kubernetesServiceQueryArgs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("kubernetes service query args = %q, want %q", got, want)
	}
}
