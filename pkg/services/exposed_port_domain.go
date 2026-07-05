package services

import (
	"crypto/md5"
	"encoding/base32"
	"fmt"
	"strconv"
	"strings"

	"github.com/Stackdome/stackdome/pkg/models"
)

// AssignExposedPortFQDNs mutates ports in place, setting generated prefixes and FQDNs
// for every port with ExposedToPublic=true. Uses stackID+resourceName for stable prefixes.
func AssignExposedPortFQDNs(stackID, resourceName, orgDomain string, ports models.Ports) {
	for i := range ports {
		if !ports[i].ExposedToPublic {
			continue
		}
		if ports[i].SubdomainPrefix != "" {
			ports[i].ExposedFqdn = fmt.Sprintf("%s.%s", ports[i].SubdomainPrefix, orgDomain)
			continue
		}
		prefix := EncodeStackResourceSubdomainPrefix(stackID, resourceName, ports[i].Number)
		ports[i].GeneratedSubdomainPrefix = prefix
		ports[i].ExposedFqdn = fmt.Sprintf("%s.%s.%s", prefix, resourceName, orgDomain)
	}
}

// EncodeStackResourceSubdomainPrefix returns a short stable subdomain token from
// stack ID, resource name, and port number.
func EncodeStackResourceSubdomainPrefix(stackID, resourceName string, port int) string {
	input := stackID + ":" + resourceName + ":" + strconv.Itoa(port)

	hasher := md5.New()
	hasher.Write([]byte(input))
	hash := hasher.Sum(nil)

	base32Encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hash)
	if len(base32Encoded) > 16 {
		base32Encoded = base32Encoded[:16]
	}
	return strings.ToLower(base32Encoded)
}

func stackResourceHasExposedPorts(ports models.Ports) bool {
	for _, port := range ports {
		if port.ExposedToPublic {
			return true
		}
	}
	return false
}

func exposedPortNumberSet(ports models.Ports) map[int]struct{} {
	set := make(map[int]struct{})
	for _, port := range ports {
		if port.ExposedToPublic {
			set[port.Number] = struct{}{}
		}
	}
	return set
}
