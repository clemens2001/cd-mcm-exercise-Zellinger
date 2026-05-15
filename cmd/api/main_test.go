package main

import "testing"

func TestGetEnvReturnsValue(t *testing.T) {
	t.Setenv("TEST_GET_ENV", "configured")

	if got := getEnv("TEST_GET_ENV", "fallback"); got != "configured" {
		t.Fatalf("expected configured value, got %q", got)
	}
}

func TestGetEnvReturnsFallback(t *testing.T) {
	if got := getEnv("TEST_GET_ENV_MISSING", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback value, got %q", got)
	}
}
