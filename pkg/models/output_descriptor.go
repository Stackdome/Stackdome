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

const (
	OutputNameHost          = "host"
	OutputNamePort          = "port"
	OutputNameDatabase      = "database"
	OutputNameUsername      = "username"
	OutputNamePassword      = "password"
	OutputNameSSLMode       = "sslmode"
	OutputNameCACertificate = "ca_certificate"
	OutputNameURL           = "url"
	OutputNamePublicHost    = "public_host"
	OutputNamePublicURL     = "public_url"
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

// stackResourceOutputKey builds a per-port output key. A single-port resource
// drops the suffix entirely; a multi-port resource disambiguates by port name.
func stackResourceOutputKey(base, portName string, multiPort bool) string {
	if !multiPort {
		return base
	}
	return base + "." + portName
}

func (r *StackResource) ToOutputMap() map[string]string {
	outputs := make(map[string]string)

	host := r.InternalServiceHost()
	outputs[OutputNameHost] = host

	multiPort := len(r.Ports) > 1
	for _, port := range r.Ports {
		outputs[stackResourceOutputKey(OutputNamePort, port.Name, multiPort)] = strconv.Itoa(port.Number)
		if port.Protocol == PortProtocolHTTP {
			outputs[stackResourceOutputKey(OutputNameURL, port.Name, multiPort)] = fmt.Sprintf("http://%s:%d", host, port.Number)
		} else {
			outputs[stackResourceOutputKey(OutputNameURL, port.Name, multiPort)] = fmt.Sprintf("%s:%d", host, port.Number)
		}

		if !port.ExposedToPublic || port.ExposedFqdn == "" {
			continue
		}
		outputs[stackResourceOutputKey(OutputNamePublicHost, port.Name, multiPort)] = port.ExposedFqdn
		outputs[stackResourceOutputKey(OutputNamePublicURL, port.Name, multiPort)] = "http://" + port.ExposedFqdn
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
		{Name: OutputNameHost, Type: OutputValueTypeString, Sensitive: false},
	}

	multiPort := len(resource.Ports) > 1
	for _, port := range resource.Ports {
		outputs = append(outputs,
			OutputDescriptor{Name: stackResourceOutputKey(OutputNamePort, port.Name, multiPort), Type: OutputValueTypeInteger, Sensitive: false},
			OutputDescriptor{Name: stackResourceOutputKey(OutputNameURL, port.Name, multiPort), Type: OutputValueTypeString, Sensitive: false},
		)
		if port.ExposedToPublic {
			outputs = append(outputs,
				OutputDescriptor{Name: stackResourceOutputKey(OutputNamePublicHost, port.Name, multiPort), Type: OutputValueTypeString, Sensitive: false},
				OutputDescriptor{Name: stackResourceOutputKey(OutputNamePublicURL, port.Name, multiPort), Type: OutputValueTypeString, Sensitive: false},
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
		{Name: OutputNameHost, Type: OutputValueTypeString, Sensitive: false},
		{Name: OutputNamePort, Type: OutputValueTypeInteger, Sensitive: false},
		{Name: OutputNameDatabase, Type: OutputValueTypeString, Sensitive: false},
		{Name: OutputNameUsername, Type: OutputValueTypeString, Sensitive: true},
		{Name: OutputNamePassword, Type: OutputValueTypeString, Sensitive: true},
		{Name: OutputNameSSLMode, Type: OutputValueTypeString, Sensitive: false},
		{Name: OutputNameCACertificate, Type: OutputValueTypeString, Sensitive: true},
		{Name: OutputNameURL, Type: OutputValueTypeString, Sensitive: true},
	}
}
