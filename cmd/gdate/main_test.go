package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var gdateBin string

func TestMain(m *testing.M) {
	// Build the gdate binary into a temp directory.
	tmp, err := os.MkdirTemp("", "gdate-test-*")
	if err != nil {
		panic(err)
	}

	gdateBin = filepath.Join(tmp, "gdate")
	cmd := exec.Command("go", "build", "-o", gdateBin, ".") //nolint:gosec // test-only: building our own binary
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("failed to build gdate: " + err.Error())
	}

	result := m.Run()
	_ = os.RemoveAll(tmp)
	os.Exit(result)
}

func runGdate(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(gdateBin, args...) //nolint:gosec // test-only: running our own built binary
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected error running gdate: %v", err)
		}
	}
	return strings.TrimSpace(outBuf.String()), strings.TrimSpace(errBuf.String()), exitCode
}

func TestDateFlag(t *testing.T) {
	stdout, _, code := runGdate(t, "-d", "2024-01-15")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "2024") || !strings.Contains(stdout, "Jan") {
		t.Errorf("unexpected output for -d 2024-01-15: %q", stdout)
	}
}

func TestOffsetFlag(t *testing.T) {
	stdout, _, code := runGdate(t, "-o", "3 days")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if stdout != "259200" {
		t.Errorf("expected 259200, got %q", stdout)
	}
}

func TestMutualExclusivity(t *testing.T) {
	_, stderr, code := runGdate(t, "-d", "2024-01-15", "-o", "3 days")
	if code == 0 {
		t.Fatal("expected non-zero exit for --date + --offset")
	}
	if !strings.Contains(stderr, "mutually exclusive") {
		t.Errorf("expected mutual exclusivity error, got stderr: %q", stderr)
	}
}

func TestNoArgs(t *testing.T) {
	stdout, _, code := runGdate(t)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if stdout == "" {
		t.Error("expected non-empty output for no args")
	}
}

func TestHelpFlag(t *testing.T) {
	stdout, _, code := runGdate(t, "-h")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "Usage") {
		t.Errorf("expected usage text, got: %q", stdout)
	}
}

func TestInvalidDate(t *testing.T) {
	_, stderr, code := runGdate(t, "-d", "not-a-date-at-all!!!")
	if code == 0 {
		t.Fatal("expected non-zero exit for invalid date")
	}
	if stderr == "" {
		t.Error("expected error on stderr for invalid date")
	}
}

func TestOffsetBareUnit(t *testing.T) {
	// "3 days" = 259200 seconds; in days that's exactly 3.
	stdout, _, code := runGdate(t, "-o", "3 days", "+days")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if stdout != "3" {
		t.Errorf("expected 3, got %q", stdout)
	}
}

func TestOffsetBareUnitDecimal(t *testing.T) {
	// "3 days and 4 hours" in days = 3.1666...
	stdout, _, code := runGdate(t, "-o", "3 days and 4 hours", "+days")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.HasPrefix(stdout, "3.166") {
		t.Errorf("expected ~3.1667, got %q", stdout)
	}
}

func TestOffsetBareUnitPlural(t *testing.T) {
	// "heleks" is not in the unit table but "helek" is; plural fallback should work.
	stdout, _, code := runGdate(t, "-o", "3 days and 4 hours", "+heleks")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.HasPrefix(stdout, "82080") {
		t.Errorf("expected ~82080, got %q", stdout)
	}
}

func TestOffsetCompositeFormat(t *testing.T) {
	// "3 days and 4 hours" with composite format: largest-to-smallest reduction.
	stdout, _, code := runGdate(t, "-o", "3 days and 4 hours", "+%{days} days %{hours} hours")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if stdout != "3 days 4 hours" {
		t.Errorf("expected '3 days 4 hours', got %q", stdout)
	}
}

func TestOffsetShortTokens(t *testing.T) {
	// Short tokens read raw field values.
	stdout, _, code := runGdate(t, "-o", "3 days and 4 hours", "+%D days %h hours")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if stdout != "3 days 4 hours" {
		t.Errorf("expected '3 days 4 hours', got %q", stdout)
	}
}

func TestOffsetFortnights(t *testing.T) {
	// "3 days and 4 hours" in fortnights = 76h / (14*24h) = 0.2261904...
	stdout, _, code := runGdate(t, "-o", "3 days and 4 hours", "+fortnights")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.HasPrefix(stdout, "0.226") {
		t.Errorf("expected ~0.2262, got %q", stdout)
	}
}
