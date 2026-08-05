package stackfile

import (
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	addonTypeKey = "type"
	yamlTagName  = "yaml"
	yamlTagSkip  = "-"
	topLevelPath = "top level"
)

// collectUnknownKeys walks the YAML against the Stackfile struct tags and reports
// keys the decoder silently drops: typos ("imagee") and removed fields ("stateful").
// Not checked: free-form maps (env, secrets) and the user-chosen names under
// resources:/volumes:/addons:.
func collectUnknownKeys(content []byte) []string {
	var root yaml.Node
	if err := yaml.Unmarshal(content, &root); err != nil {
		return nil
	}
	node := &root
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}

	var warnings []string
	walkUnknownKeys(node, reflect.TypeOf(Stackfile{}), "", &warnings)
	return warnings
}

var addonConfigType = reflect.TypeOf(AddonConnectionConfig{})

func walkUnknownKeys(node *yaml.Node, t reflect.Type, path string, out *[]string) {
	if node == nil {
		return
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		if node.Kind != yaml.MappingNode {
			return
		}
		if t == addonConfigType {
			walkAddon(node, path, out)
			return
		}
		fields := yamlFields(t)
		eachMapEntry(node, func(key string, val *yaml.Node) {
			ft, ok := fields[key]
			if !ok {
				*out = append(*out, unknownKeyWarning(key, path))
				return
			}
			walkUnknownKeys(val, ft, childPath(path, key), out)
		})

	case reflect.Map:
		// Keys are user-chosen names, so only the values are checked. Maps of
		// scalars (env, secrets) have nothing to check.
		if node.Kind != yaml.MappingNode || !isCheckable(t.Elem()) {
			return
		}
		eachMapEntry(node, func(key string, val *yaml.Node) {
			walkUnknownKeys(val, t.Elem(), childPath(path, key), out)
		})

	case reflect.Slice:
		if node.Kind != yaml.SequenceNode || !isCheckable(t.Elem()) {
			return
		}
		for i, item := range node.Content {
			walkUnknownKeys(item, t.Elem(), fmt.Sprintf("%s[%d]", path, i), out)
		}
	}
}

// Addon blocks decode through a custom unmarshaller: a common base plus
// type-specific fields. Extra keys on an unrecognised type are left alone.
func walkAddon(node *yaml.Node, path string, out *[]string) {
	allowed := yamlFields(addonConfigType)
	switch addonTypeOf(node) {
	case PostgresAddonType:
		for k, v := range yamlFields(reflect.TypeOf(PostgresAddonConfig{})) {
			allowed[k] = v
		}
	default:
		return
	}

	eachMapEntry(node, func(key string, _ *yaml.Node) {
		if _, ok := allowed[key]; !ok {
			*out = append(*out, unknownKeyWarning(key, path))
		}
	})
}

func addonTypeOf(node *yaml.Node) string {
	var addonType string
	eachMapEntry(node, func(key string, val *yaml.Node) {
		if key == addonTypeKey {
			addonType = val.Value
		}
	})
	return addonType
}

func eachMapEntry(node *yaml.Node, fn func(key string, val *yaml.Node)) {
	for i := 0; i+1 < len(node.Content); i += 2 {
		fn(node.Content[i].Value, node.Content[i+1])
	}
}

// yamlFields maps a struct's yaml key names to their field types.
func yamlFields(t reflect.Type) map[string]reflect.Type {
	fields := make(map[string]reflect.Type, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get(yamlTagName), ",")
		if name == yamlTagSkip {
			continue
		}
		if name == "" {
			name = strings.ToLower(f.Name)
		}
		fields[name] = f.Type
	}
	return fields
}

// Only containers of structs carry keys worth checking.
func isCheckable(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		return true
	case reflect.Map, reflect.Slice:
		return isCheckable(t.Elem())
	default:
		return false
	}
}

func childPath(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}

func unknownKeyWarning(key, path string) string {
	if path == "" {
		path = topLevelPath
	}
	return fmt.Sprintf("unknown key %q in %s", key, path)
}
