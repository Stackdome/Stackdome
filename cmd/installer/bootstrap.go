package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ashishmax31/stackdome-api-server/install"
)

var apiBaseURL string

type bootstrapResult struct {
	Token     string
	OrgID     string
	ClusterID string
}

func runAPIBootstrap(vals install.TemplateValues, secrets *BootstrapSecrets) (*bootstrapResult, error) {
	phaseLog(5, "Bootstrapping platform via API...")

	apiBaseURL = apiServerBaseURL()
	stepLog(fmt.Sprintf("API server URL: %s", apiBaseURL))

	token, orgID, err := authenticate(vals.AdminEmail, secrets.AdminPassword)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	stepLog(fmt.Sprintf("Authenticated -- org ID: %s", orgID))

	if err := configureDomain(token, orgID, vals.Domain); err != nil {
		return nil, fmt.Errorf("domain configuration failed: %w", err)
	}
	stepLog(fmt.Sprintf("Domain configured: %s", vals.Domain))

	stepLog("Deploying RBAC resources...")
	rbacManifest, err := install.ReadManifest("rbac.yaml")
	if err != nil {
		return nil, fmt.Errorf("reading RBAC manifest: %w", err)
	}
	if err := kubectlApply(rbacManifest); err != nil {
		return nil, fmt.Errorf("applying RBAC: %w", err)
	}

	time.Sleep(5 * time.Second)

	clusterURL, caData, saToken, err := extractClusterCredentials()
	if err != nil {
		return nil, fmt.Errorf("extracting cluster credentials: %w", err)
	}
	stepLog("Cluster credentials extracted")

	clusterID, err := registerCluster(token, orgID, clusterURL, caData, saToken)
	if err != nil {
		return nil, fmt.Errorf("cluster registration failed: %w", err)
	}
	stepLog(fmt.Sprintf("Cluster registered -- ID: %s", clusterID))

	successLog("API bootstrap complete")
	return &bootstrapResult{Token: token, OrgID: orgID, ClusterID: clusterID}, nil
}

func authenticate(email, password string) (token, orgID string, err error) {
	token, orgID, err = tryLogin(email, password)
	if err == nil {
		stepLog("Logged in as existing user")
		return token, orgID, nil
	}

	stepLog(fmt.Sprintf("Signing up %s...", email))
	token, orgID, err = signup(email, password)
	if err == nil {
		return token, orgID, nil
	}

	// Signup may fail with 409 if user exists but login failed due to timing.
	// Retry login.
	stepLog("Signup failed, retrying login...")
	time.Sleep(2 * time.Second)
	return tryLogin(email, password)
}

func tryLogin(email, password string) (string, string, error) {
	payload, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	resp, err := http.Post(apiBaseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("login returned %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	token := parseJSONField(body, "token")
	orgID := parseNestedJSONField(body, "user", "organisation_id")
	if token == "" || orgID == "" {
		return "", "", fmt.Errorf("missing token or org_id in login response")
	}
	return token, orgID, nil
}

func signup(email, password string) (string, string, error) {
	payload, _ := json.Marshal(map[string]interface{}{
		"name":     "Platform Admin",
		"email":    email,
		"password": password,
		"organisation": map[string]string{
			"name": "Default",
		},
	})

	resp, err := http.Post(apiBaseURL+"/api/v1/user-signup", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", "", fmt.Errorf("signup request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("signup returned %d: %s", resp.StatusCode, string(body))
	}

	token := parseJSONField(body, "jwt_token")
	orgID := parseNestedJSONField(body, "user", "organisation_id")
	if token == "" || orgID == "" {
		return "", "", fmt.Errorf("missing jwt_token or org_id in signup response")
	}
	return token, orgID, nil
}

func configureDomain(token, orgID, domain string) error {
	if domainExists(token, orgID, domain) {
		stepLog("Domain already configured")
		return nil
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"domains": []map[string]string{
			{"fqdn": domain},
		},
	})

	req, _ := http.NewRequest("PUT",
		fmt.Sprintf("%s/api/v1/organizations/%s", apiBaseURL, orgID),
		bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("domain config returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

type orgResponse struct {
	Domains []struct {
		FQDN string `json:"fqdn"`
	} `json:"domains"`
}

func domainExists(token, orgID, domain string) bool {
	req, _ := http.NewRequest("GET",
		fmt.Sprintf("%s/api/v1/organizations/%s", apiBaseURL, orgID),
		nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	body, _ := io.ReadAll(resp.Body)

	var org orgResponse
	if json.Unmarshal(body, &org) != nil {
		return false
	}

	for _, d := range org.Domains {
		if d.FQDN == domain {
			return true
		}
	}
	return false
}

func extractClusterCredentials() (clusterURL, caData, saToken string, err error) {
	// Use the kubernetes service ClusterIP instead of the kubeconfig URL (127.0.0.1:6443),
	// because the API server runs inside the cluster and needs an in-cluster address.
	k8sSvcIP, err := output("kubectl", "get", "svc", "kubernetes",
		"-o", "jsonpath={.spec.clusterIP}")
	if err != nil {
		return "", "", "", fmt.Errorf("getting kubernetes service IP: %w", err)
	}
	clusterURL = fmt.Sprintf("https://%s:443", k8sSvcIP)

	caData, err = output("kubectl", "get", "secret",
		"stackdome-api-server-account-secret",
		"-n", chartNamespace,
		"-o", "jsonpath={.data.ca\\.crt}")
	if err != nil {
		return "", "", "", fmt.Errorf("getting CA data: %w", err)
	}

	saTokenB64, err := output("kubectl", "get", "secret",
		"stackdome-api-server-account-secret",
		"-n", chartNamespace,
		"-o", "jsonpath={.data.token}")
	if err != nil {
		return "", "", "", fmt.Errorf("getting SA token: %w", err)
	}

	decoded, err := output("sh", "-c", fmt.Sprintf("echo '%s' | base64 -d", saTokenB64))
	if err != nil {
		return "", "", "", fmt.Errorf("decoding SA token: %w", err)
	}

	return clusterURL, caData, decoded, nil
}

func registerCluster(token, orgID, clusterURL, caData, saToken string) (string, error) {
	req, _ := http.NewRequest("GET",
		fmt.Sprintf("%s/api/v1/organizations/%s/clusters", apiBaseURL, orgID),
		nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var listResp struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
		}
		if json.Unmarshal(body, &listResp) == nil && len(listResp.Items) > 0 {
			stepLog("Cluster already registered -- reusing")
			return listResp.Items[0].ID, nil
		}
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"name":             "local",
		"cluster_url":      clusterURL,
		"cluster_ca_data":  caData,
		"cluster_sa_token": saToken,
		"cluster_image_registry": map[string]interface{}{
			"name": "default-registry",
			"spec": map[string]interface{}{
				"backend_storage_size": "50Gi",
				"max_repositories":     10,
				"tags_per_repository":  5,
				"delete_untagged":      true,
			},
		},
	})

	req, _ = http.NewRequest("POST",
		fmt.Sprintf("%s/api/v1/organizations/%s/clusters", apiBaseURL, orgID),
		bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("cluster registration returned %d: %s", resp.StatusCode, string(body))
	}

	clusterID := parseJSONField(body, "id")
	if clusterID == "" {
		return "", fmt.Errorf("missing id in cluster registration response")
	}
	return clusterID, nil
}

func parseJSONField(body []byte, field string) string {
	var m map[string]interface{}
	if json.Unmarshal(body, &m) != nil {
		return ""
	}
	s, _ := m[field].(string)
	return s
}

func parseNestedJSONField(body []byte, path ...string) string {
	var current interface{}
	if json.Unmarshal(body, &current) != nil {
		return ""
	}
	for _, key := range path {
		m, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current = m[key]
	}
	s, _ := current.(string)
	return s
}
