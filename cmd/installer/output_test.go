package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestJSONArgumentFailureSeparatesOutput(t *testing.T) {
	previous := installerOutput
	t.Cleanup(func() { installerOutput = previous })

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	const secretFixture = "fixture-super-secret"
	code := runMainWithIO(
		[]string{"install", "--output", "json", "--", "--github-client-secret", secretFixture},
		&stdout,
		&stderr,
		true,
	)
	if code == 0 {
		t.Fatal("runMainWithIO returned success for invalid arguments")
	}
	if strings.Count(stdout.String(), "\n") != 1 || !strings.HasSuffix(stdout.String(), "\n") {
		t.Fatalf("JSON stdout = %q, want exactly one newline-terminated object", stdout.String())
	}
	var result errorResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("JSON stdout is not parseable: %v", err)
	}
	if result.Status != "error" || result.Phase != "arguments" {
		t.Fatalf("JSON error = %#v, want arguments error", result)
	}
	if strings.Contains(stdout.String(), "\x1b[") {
		t.Fatalf("JSON stdout contains ANSI escapes: %q", stdout.String())
	}
	if stderr.Len() == 0 {
		t.Fatal("diagnostic stderr is empty")
	}
	if strings.Contains(stdout.String(), secretFixture) || strings.Contains(stderr.String(), secretFixture) {
		t.Fatal("secret-like positional argument leaked to installer output")
	}
}

func TestWriteCredentialsEnforcesFilesystemBoundary(t *testing.T) {
	credentials := bootstrapCredentials{
		URL:           "https://stackdome.example.com",
		AdminEmail:    "installer@example.com",
		AdminPassword: "fixture-password",
	}
	privateDir := filepath.Join(t.TempDir(), "stackdome")
	path := filepath.Join(privateDir, "bootstrap.json")
	if err := writeCredentials(path, credentials); err != nil {
		t.Fatalf("writeCredentials() error = %v", err)
	}

	dirInfo, err := os.Stat(privateDir)
	if err != nil {
		t.Fatal(err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Fatalf("credentials directory mode = %o, want 700", got)
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fileInfo.Mode().Perm(); got != 0o600 {
		t.Fatalf("credentials file mode = %o, want 600", got)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got bootstrapCredentials
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, credentials) {
		t.Fatalf("credentials = %#v, want %#v", got, credentials)
	}

	replacement := credentials
	replacement.AdminPassword = "replacement-password"
	if err := writeCredentials(path, replacement); err != nil {
		t.Fatalf("replacing regular credentials file: %v", err)
	}
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), credentials.AdminPassword) || !strings.Contains(string(data), replacement.AdminPassword) {
		t.Fatalf("credential replacement did not atomically install new content: %s", data)
	}

	target := filepath.Join(privateDir, "target.json")
	if err := os.WriteFile(target, []byte("target"), 0o600); err != nil {
		t.Fatal(err)
	}
	symlink := filepath.Join(privateDir, "symlink.json")
	if err := os.Symlink(target, symlink); err != nil {
		t.Fatal(err)
	}
	if err := writeCredentials(symlink, credentials); err == nil {
		t.Fatal("writeCredentials() accepted a symlink destination")
	}

	directoryDestination := filepath.Join(privateDir, "directory.json")
	if err := os.Mkdir(directoryDestination, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := writeCredentials(directoryDestination, credentials); err == nil {
		t.Fatal("writeCredentials() accepted a directory destination")
	}
}

func TestCommandLabelDoesNotIncludeTrailingArguments(t *testing.T) {
	label := commandLabel("helm", []string{"upgrade", "--set", "password=fixture-secret"})
	if label != "helm upgrade" {
		t.Fatalf("commandLabel() = %q, want %q", label, "helm upgrade")
	}
	if strings.Contains(label, "fixture-secret") {
		t.Fatalf("commandLabel() leaked trailing argument: %q", label)
	}
}
