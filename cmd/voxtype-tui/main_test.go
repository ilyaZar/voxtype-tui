package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRejectsUnexpectedToggleArg(t *testing.T) {
	err := run([]string{"toggle", "extra"})
	if err == nil || !strings.Contains(err.Error(), "unexpected toggle argument") {
		t.Fatalf("err = %v", err)
	}
}

func TestRunRejectsUnexpectedVersionArg(t *testing.T) {
	err := run([]string{"version", "extra"})
	if err == nil || !strings.Contains(err.Error(), "unexpected version argument") {
		t.Fatalf("err = %v", err)
	}
}

func TestRunRecordAcceptsConfigFlagBeforeOrAfterAction(t *testing.T) {
	path := writeRecordConfig(t)

	if err := run([]string{"record", "--config-file", path, "toggle"}); err != nil {
		t.Fatalf("flag before action: %v", err)
	}
	if err := run([]string{"record", "toggle", "--config-file", path}); err != nil {
		t.Fatalf("flag after action: %v", err)
	}
}

func TestRunRecordRejectsUnexpectedArgAfterAction(t *testing.T) {
	err := run([]string{"record", "toggle", "extra"})
	if err == nil || !strings.Contains(err.Error(), "unexpected record argument") {
		t.Fatalf("err = %v", err)
	}
}

func writeRecordConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	data := []byte("[voxtype]\ncommand = \"true\"\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
