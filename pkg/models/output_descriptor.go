package models

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type OutputValueType string

const (
	OutputValueTypeString  OutputValueType = "string"
	OutputValueTypeInteger OutputValueType = "integer"
	OutputValueTypeBoolean OutputValueType = "boolean"
)

type OutputDescriptor struct {
	Name      string          `json:"name"`
	Type      OutputValueType `json:"type"`
	Sensitive bool            `json:"sensitive"`
}

var simpleSecretOutputKeyPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

func (r *StackResource) EnsureDeclaredOutputs() []OutputDescriptor {
	if len(r.Outputs) > 0 {
		return r.Outputs
	}
	r.Outputs = StackResourceOutputDescriptors(r)
	return r.Outputs
}

func (s *Secret) EnsureDeclaredOutputs() []OutputDescriptor {
	if len(s.Outputs) > 0 {
		return s.Outputs
	}
	s.Outputs = SecretOutputDescriptors(s)
	return s.Outputs
}

func (p *PostgresAddon) EnsureDeclaredOutputs() []OutputDescriptor {
	if len(p.Outputs) > 0 {
		return p.Outputs
	}
	p.Outputs = PostgresAddonOutputDescriptors(p)
	return p.Outputs
}

func (r *StackResource) ToOutputMap() map[string]string {
	outputs := make(map[string]string)

	host := r.Name
	if r.Status != nil && r.Status.InternalServiceName != nil && *r.Status.InternalServiceName != "" {
		host = *r.Status.InternalServiceName
	}
	outputs["host"] = host

	publicURLsByPort := make(map[int]string)
	publicHostsByPort := make(map[int]string)
	if r.Status != nil {
		for _, ingress := range r.Status.PublicIngresses {
			if ingress.URL == "" {
				continue
			}
			publicURLsByPort[ingress.TargetPort] = ingress.URL
			if parsed, err := url.Parse(ingress.URL); err == nil && parsed.Host != "" {
				publicHostsByPort[ingress.TargetPort] = parsed.Host
			}
		}
	}

	for _, port := range r.Ports {
		outputs["port."+port.Name] = strconv.Itoa(port.Number)
		outputs["url."+port.Name] = fmt.Sprintf("%s://%s:%d", port.Protocol, host, port.Number)

		if !port.ExposedToPublic {
			continue
		}

		publicHost := publicHostsByPort[port.Number]
		if publicHost == "" {
			publicHost = port.ExposedFqdn
		}
		if publicHost != "" {
			outputs["public."+port.Name+".host"] = publicHost
		}

		publicURL := publicURLsByPort[port.Number]
		if publicURL == "" && port.ExposedFqdn != "" {
			publicURL = "http://" + port.ExposedFqdn
		}
		if publicURL != "" {
			outputs["public."+port.Name+".url"] = publicURL
		}
	}

	return outputs
}

func (s *Secret) ToOutputMap() map[string]string {
	outputs := make(map[string]string, len(s.Data))
	for key, value := range s.Data {
		outputs[secretOutputAccessor(key)] = value
	}
	return outputs
}

func StackResourceOutputDescriptors(resource *StackResource) []OutputDescriptor {
	outputs := []OutputDescriptor{
		{
			Name:      "host",
			Type:      OutputValueTypeString,
			Sensitive: false,
		},
	}

	for _, port := range resource.Ports {
		outputs = append(outputs,
			OutputDescriptor{
				Name:      "port." + port.Name,
				Type:      OutputValueTypeInteger,
				Sensitive: false,
			},
			OutputDescriptor{
				Name:      "url." + port.Name,
				Type:      OutputValueTypeString,
				Sensitive: false,
			},
		)

		if port.ExposedToPublic {
			outputs = append(outputs,
				OutputDescriptor{
					Name:      "public." + port.Name + ".host",
					Type:      OutputValueTypeString,
					Sensitive: false,
				},
				OutputDescriptor{
					Name:      "public." + port.Name + ".url",
					Type:      OutputValueTypeString,
					Sensitive: false,
				},
			)
		}
	}

	return outputs
}

func SecretOutputDescriptors(secret *Secret) []OutputDescriptor {
	outputs := make([]OutputDescriptor, 0, len(secret.Keys))
	for _, key := range secret.Keys {
		outputs = append(outputs, OutputDescriptor{
			Name:      secretOutputAccessor(key),
			Type:      OutputValueTypeString,
			Sensitive: true,
		})
	}
	return outputs
}

func PostgresAddonOutputDescriptors(_ *PostgresAddon) []OutputDescriptor {
	return []OutputDescriptor{
		{Name: "host", Type: OutputValueTypeString, Sensitive: false},
		{Name: "port", Type: OutputValueTypeInteger, Sensitive: false},
		{Name: "database", Type: OutputValueTypeString, Sensitive: false},
		{Name: "username", Type: OutputValueTypeString, Sensitive: true},
		{Name: "password", Type: OutputValueTypeString, Sensitive: true},
		{Name: "sslmode", Type: OutputValueTypeString, Sensitive: false},
		{Name: "ca_certificate", Type: OutputValueTypeString, Sensitive: true},
		{Name: "url", Type: OutputValueTypeString, Sensitive: true},
	}
}

func secretOutputAccessor(key string) string {
	if isSimpleSecretOutputKey(key) {
		return "key." + key
	}
	escaped := strings.ReplaceAll(key, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "'", "\\'")
	return "key['" + escaped + "']"
}

func isSimpleSecretOutputKey(key string) bool {
	// Simple identifiers use dot syntax; everything else falls back to quoted
	// bracket syntax so keys like tls.crt remain addressable without ambiguity.
	return simpleSecretOutputKeyPattern.MatchString(key)
}
