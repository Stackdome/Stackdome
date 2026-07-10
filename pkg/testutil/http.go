package testutil

import (
	"bytes"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func LoadKubernetesObjectsFromBytes(fileYaml []byte) ([]unstructured.Unstructured, error) {
	// Trim empty starting and ending objects
	fileYaml = bytes.Trim(fileYaml, "-\n")

	var objects []unstructured.Unstructured
	// Split for every included yaml document.
	for i, yamlDocument := range bytes.Split(fileYaml, []byte("---\n")) {
		obj := unstructured.Unstructured{}
		if err := yaml.Unmarshal(yamlDocument, &obj); err != nil {
			return nil, fmt.Errorf(
				"unmarshalling yaml document at index %d: %w", i, err)
		}

		objects = append(objects, obj)
	}

	return objects, nil
}
