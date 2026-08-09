package install

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBootstrapRootFailureOutputModes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bootstrap.sh requires POSIX sh")
	}
	fakeBin := t.TempDir()
	fakeID := filepath.Join(fakeBin, "id")
	if err := os.WriteFile(fakeID, []byte("#!/bin/sh\nprintf '501\\n'\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	run := func(args ...string) (string, string, int) {
		t.Helper()
		cmd := exec.Command("sh", append([]string{"bootstrap.sh"}, args...)...)
		cmd.Env = append(os.Environ(), "PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err == nil {
			return stdout.String(), stderr.String(), 0
		}
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("bootstrap execution error = %v", err)
		}
		return stdout.String(), stderr.String(), exitErr.ExitCode()
	}

	jsonStdout, jsonStderr, jsonCode := run("--output", "json")
	if jsonCode == 0 {
		t.Fatal("JSON root failure exited zero")
	}
	wantJSON := "{\"status\":\"error\",\"phase\":\"transport\",\"message\":\"root access required\"}\n"
	if jsonStdout != wantJSON {
		t.Fatalf("JSON stdout = %q, want %q", jsonStdout, wantJSON)
	}
	if !strings.Contains(jsonStderr, "root access required") {
		t.Fatalf("JSON stderr = %q, want root diagnostic", jsonStderr)
	}

	humanStdout, humanStderr, humanCode := run()
	if humanCode == 0 {
		t.Fatal("human root failure exited zero")
	}
	if humanStdout != "" {
		t.Fatalf("human stdout = %q, want empty", humanStdout)
	}
	if !strings.Contains(humanStderr, "curl -fsSL https://get.stackdome.com/install.sh | sudo sh") {
		t.Fatalf("human stderr = %q, want canonical command", humanStderr)
	}
}
