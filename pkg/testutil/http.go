package testutil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/mt-sre/devkube/dev"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

type httpObjectApplierImpl struct {
	URLs []string
}

func httpObjectApplier(urls ...string) *httpObjectApplierImpl {
	return &httpObjectApplierImpl{URLs: urls}
}

func (a *httpObjectApplierImpl) Init(ctx context.Context, cluster *dev.Cluster) error {
	if len(a.URLs) == 0 {
		return nil
	}
	var objects []unstructured.Unstructured
	for _, url := range a.URLs {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to fetch URL %s: %w", url, err)
		}
		defer resp.Body.Close()

		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response from %s: %w", url, err)
		}

		objs, err := LoadKubernetesObjectsFromBytes(content)
		if err != nil {
			return fmt.Errorf("loading objects from %s: %w", url, err)
		}
		objects = append(objects, objs...)
	}

	for _, obj := range objects {
		if err := cluster.CtrlClient.Create(ctx, &obj); err != nil {
			if !k8serrors.IsAlreadyExists(err) {
				return fmt.Errorf("creating object from %s: %w", obj.GetName(), err)
			}
		}
	}
	return nil
}

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
