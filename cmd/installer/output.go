package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/term"
)

type outputMode string

const (
	outputHuman outputMode = "human"
	outputJSON  outputMode = "json"

	defaultCredentialsFile = "/root/.config/stackdome/bootstrap.json"
)

type outputRuntime struct {
	mode     outputMode
	stdout   io.Writer
	stderr   io.Writer
	color    bool
	terminal bool
}

var installerOutput = outputRuntime{
	mode:     outputHuman,
	stdout:   os.Stdout,
	stderr:   os.Stderr,
	color:    term.IsTerminal(int(os.Stderr.Fd())),
	terminal: term.IsTerminal(int(os.Stderr.Fd())),
}

func stderrIsTerminal() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}

func configureOutput(mode string, noColor, terminal bool, stdout, stderr io.Writer) error {
	parsed := outputMode(mode)
	if parsed != outputHuman && parsed != outputJSON {
		return fmt.Errorf("--output must be human or json")
	}
	installerOutput = outputRuntime{
		mode:     parsed,
		stdout:   stdout,
		stderr:   stderr,
		color:    parsed == outputHuman && terminal && !noColor,
		terminal: terminal,
	}
	return nil
}

func (o outputRuntime) isJSON() bool {
	return o.mode == outputJSON
}

func (o outputRuntime) diagnosticf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(o.stderr, format, args...)
}

func (o outputRuntime) finalf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(o.stdout, format, args...)
}

type outputFlags struct {
	mode    *string
	noColor *bool
}

func registerOutputFlags(fs *flag.FlagSet) *outputFlags {
	return &outputFlags{
		mode:    fs.String("output", string(outputHuman), "Output format: human or json"),
		noColor: fs.Bool("no-color", false, "Disable ANSI color output"),
	}
}

func (f *outputFlags) apply() error {
	return configureOutput(*f.mode, *f.noColor, installerOutput.terminal, installerOutput.stdout, installerOutput.stderr)
}

func argumentsRequestJSON(args []string) bool {
	outputValue := ""
	wantValue := false
	for _, arg := range args {
		if wantValue {
			outputValue = arg
			wantValue = false
			continue
		}
		switch arg {
		case "--output":
			wantValue = true
		case "--output=json":
			outputValue = string(outputJSON)
		case "--output=human":
			outputValue = string(outputHuman)
		}
	}
	return outputValue == string(outputJSON)
}

func argumentsRequestNoColor(args []string) bool {
	for _, arg := range args {
		if arg == "--no-color" {
			return true
		}
	}
	return false
}

type installerError struct {
	Phase  string
	Public string
	Err    error
}

func (e *installerError) Error() string {
	if e.Err == nil || e.Public == e.Err.Error() {
		return e.Public
	}
	return fmt.Sprintf("%s: %v", e.Public, e.Err)
}

func (e *installerError) Unwrap() error {
	return e.Err
}

func installationError(phase, public string, err error) error {
	return &installerError{Phase: phase, Public: public, Err: err}
}

type namedResult struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type installSuccessResult struct {
	Status          string      `json:"status"`
	URL             string      `json:"url"`
	AdminEmail      string      `json:"admin_email"`
	CredentialsFile string      `json:"credentials_file"`
	Organization    namedResult `json:"organization"`
	Cluster         namedResult `json:"cluster"`
}

type upgradeSuccessResult struct {
	Status            string `json:"status"`
	URL               string `json:"url"`
	APIServerImage    string `json:"api_server_image"`
	AgentChartVersion string `json:"agent_chart_version"`
}

type errorResult struct {
	Status  string `json:"status"`
	Phase   string `json:"phase"`
	Message string `json:"message"`
}

type bootstrapCredentials struct {
	URL           string `json:"url"`
	AdminEmail    string `json:"admin_email"`
	AdminPassword string `json:"admin_password"`
}

func emitJSON(value interface{}) error {
	return json.NewEncoder(installerOutput.stdout).Encode(value)
}

func emitFailure(err error) {
	phase := "unknown"
	public := "installer failed"
	var installErr *installerError
	if errors.As(err, &installErr) {
		phase = installErr.Phase
		public = installErr.Public
	}
	errLog(err.Error())
	if installerOutput.isJSON() {
		if encodeErr := emitJSON(errorResult{Status: "error", Phase: phase, Message: public}); encodeErr != nil {
			installerOutput.diagnosticf("failed to encode JSON error: %v\n", encodeErr)
		}
	}
}

func credentialsPath(requested string) string {
	if requested == "" && installerOutput.isJSON() {
		return defaultCredentialsFile
	}
	return requested
}

func writeCredentials(path string, credentials bootstrapCredentials) error {
	if path == "" {
		return fmt.Errorf("credentials path is empty")
	}
	parent := filepath.Dir(path)
	parentInfo, err := os.Lstat(parent)
	switch {
	case err == nil:
		if parentInfo.Mode()&os.ModeSymlink != 0 || !parentInfo.IsDir() {
			return fmt.Errorf("credentials parent %s is not a directory", parent)
		}
		if parentInfo.Mode().Perm()&0o077 != 0 {
			return fmt.Errorf("credentials parent %s must not be accessible by group or other users", parent)
		}
	case os.IsNotExist(err):
		if err := os.MkdirAll(parent, 0o700); err != nil {
			return fmt.Errorf("creating credentials directory: %w", err)
		}
		if err := os.Chmod(parent, 0o700); err != nil {
			return fmt.Errorf("securing credentials directory: %w", err)
		}
	case err != nil:
		return fmt.Errorf("inspecting credentials directory: %w", err)
	}

	if existing, err := os.Lstat(path); err == nil {
		if existing.Mode()&os.ModeSymlink != 0 || !existing.Mode().IsRegular() {
			return fmt.Errorf("credentials destination must be a regular file")
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspecting credentials destination: %w", err)
	}

	tmp, err := os.CreateTemp(parent, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("creating temporary credentials file: %w", err)
	}
	tmpPath := tmp.Name()
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tmpPath)
		}
	}()

	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("securing temporary credentials file: %w", err)
	}
	if err := json.NewEncoder(tmp).Encode(credentials); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("encoding credentials: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("syncing credentials: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing credentials: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("installing credentials file: %w", err)
	}
	removeTemp = false
	return nil
}
