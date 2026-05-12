package config

import (
	"os"
	"testing"
)

func TestStringVarLookup_Found(t *testing.T) {
	os.Setenv("TEST_STR_VAR", "hello")
	defer os.Unsetenv("TEST_STR_VAR")

	v := StringVar("TEST_STR_VAR", "test", nil, true)
	val, ok := v.Lookup()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if val != "hello" {
		t.Fatalf("expected 'hello', got %q", val)
	}
}

func TestStringVarLookup_NotFound_NoDefault(t *testing.T) {
	os.Unsetenv("TEST_STR_MISSING")

	v := StringVar("TEST_STR_MISSING", "test", nil, true)
	val, ok := v.Lookup()
	if ok {
		t.Fatal("expected ok=false")
	}
	if val != "" {
		t.Fatalf("expected empty string, got %q", val)
	}
}

func TestStringVarLookup_NotFound_WithDefault(t *testing.T) {
	os.Unsetenv("TEST_STR_DEFAULT")

	v := StringVar("TEST_STR_DEFAULT", "test", ptr("fallback"), false)
	val, ok := v.Lookup()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if val != "fallback" {
		t.Fatalf("expected 'fallback', got %q", val)
	}
}

func TestStringVarLookup_TrimsWhitespace(t *testing.T) {
	os.Setenv("TEST_STR_TRIM", "  spaced  ")
	defer os.Unsetenv("TEST_STR_TRIM")

	v := StringVar("TEST_STR_TRIM", "test", nil, false)
	val, ok := v.Lookup()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if val != "spaced" {
		t.Fatalf("expected 'spaced', got %q", val)
	}
}

func TestIntVarLookup_Valid(t *testing.T) {
	os.Setenv("TEST_INT_VAR", "42")
	defer os.Unsetenv("TEST_INT_VAR")

	v := IntVar("TEST_INT_VAR", "test", nil, true)
	val, ok := v.Lookup()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if val != 42 {
		t.Fatalf("expected 42, got %d", val)
	}
}

func TestIntVarLookup_Invalid_FallsBackToDefault(t *testing.T) {
	os.Setenv("TEST_INT_BAD", "notanumber")
	defer os.Unsetenv("TEST_INT_BAD")

	v := IntVar("TEST_INT_BAD", "test", ptr(99), false)
	val, ok := v.Lookup()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if val != 99 {
		t.Fatalf("expected 99, got %d", val)
	}
}

func TestIntVarLookup_Invalid_NoDefault(t *testing.T) {
	os.Setenv("TEST_INT_BAD2", "notanumber")
	defer os.Unsetenv("TEST_INT_BAD2")

	v := IntVar("TEST_INT_BAD2", "test", nil, false)
	val, ok := v.Lookup()
	if ok {
		t.Fatal("expected ok=false")
	}
	if val != 0 {
		t.Fatalf("expected 0, got %d", val)
	}
}

func TestBoolVarLookup_True(t *testing.T) {
	os.Setenv("TEST_BOOL_VAR", "true")
	defer os.Unsetenv("TEST_BOOL_VAR")

	v := BoolVar("TEST_BOOL_VAR", "test", nil, false)
	val, ok := v.Lookup()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if !val {
		t.Fatal("expected true")
	}
}

func TestBoolVarLookup_False(t *testing.T) {
	os.Setenv("TEST_BOOL_VAR2", "false")
	defer os.Unsetenv("TEST_BOOL_VAR2")

	v := BoolVar("TEST_BOOL_VAR2", "test", ptr(true), false)
	val, ok := v.Lookup()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if val {
		t.Fatal("expected false")
	}
}

func TestBoolVarLookup_NotFound_WithDefault(t *testing.T) {
	os.Unsetenv("TEST_BOOL_MISSING")

	v := BoolVar("TEST_BOOL_MISSING", "test", ptr(true), false)
	val, ok := v.Lookup()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if !val {
		t.Fatal("expected true")
	}
}
