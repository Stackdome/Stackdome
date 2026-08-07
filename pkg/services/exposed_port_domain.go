package services

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/slug"
)

const platformDomainLabelMaxLength = 63

func platformDomainID(resourceID string, port int) string {
	sum := sha256.Sum256([]byte(resourceID + ":" + strconv.Itoa(port)))
	return hex.EncodeToString(sum[:])[:8]
}

func FQDNForPortWithPlatformDomain(resourceID, resourceName, baseDomain string, port models.Port) (string, string, error) {
	readableHead := resourceName
	if port.SubdomainPrefix != "" {
		readableHead = port.SubdomainPrefix
	}

	head := slug.Make(readableHead)
	if head == "" {
		return "", "", errors.New("platform domain slug is empty")
	}

	id := platformDomainID(resourceID, port.Number)
	maxHeadLength := platformDomainLabelMaxLength - len(id) - 1
	if len(head) > maxHeadLength {
		head = strings.TrimRight(head[:maxHeadLength], "-")
	}

	return fmt.Sprintf("%s-%s.%s", head, id, baseDomain), id, nil
}

func FQDNForPortWithCustomDomain(stackID, resourceName, customDomain string, port models.Port) (fqdn, generatedPrefix string) {
	if port.SubdomainPrefix != "" {
		return fmt.Sprintf("%s.%s", port.SubdomainPrefix, customDomain), ""
	}

	generatedPrefix = EncodeStackResourceSubdomainPrefix(stackID, resourceName, port.Number)
	return fmt.Sprintf("%s.%s.%s", generatedPrefix, resourceName, customDomain), generatedPrefix
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
