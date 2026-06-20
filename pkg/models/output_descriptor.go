package models

import (
	"fmt"
	"strconv"
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

	host := r.InternalServiceHost()
	outputs["host"] = host

	for _, port := range r.Ports {
		outputs["port."+port.Name] = strconv.Itoa(port.Number)
		if port.Protocol == "http" {
			outputs["url."+port.Name] = fmt.Sprintf("http://%s:%d", host, port.Number)
		} else {
			outputs["url."+port.Name] = fmt.Sprintf("%s:%d", host, port.Number)
		}

		if !port.ExposedToPublic {
			continue
		}

		if port.ExposedFqdn == "" {
			continue
		}
		outputs["public."+port.Name+".host"] = port.ExposedFqdn
		outputs["public."+port.Name+".url"] = "http://" + port.ExposedFqdn
	}

	return outputs
}

func (r *StackResource) InternalServiceHost() string {
	if r.Namespace == "" {
		return r.Name
	}
	return fmt.Sprintf("%s.%s.svc", r.Name, r.Namespace)
}

func (s *Secret) ToOutputMap() map[string]string {
	outputs := make(map[string]string, len(s.Data))
	for key, value := range s.Data {
		outputs[key] = value
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
			Name:      key,
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
