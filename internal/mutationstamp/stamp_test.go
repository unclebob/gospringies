package mutationstamp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStampWritesValidStampForUnstampedContent(t *testing.T) {
	path := writeTempFeature(t, "Feature: Smoke\n\nScenario: one\n  Given ready\n")

	if err := Stamp(path); err != nil {
		t.Fatal(err)
	}

	content := readTempFeature(t, path)
	if !strings.HasPrefix(content, Prefix) {
		t.Fatalf("content was not stamped:\n%s", content)
	}
	if !Valid(path) {
		t.Fatal("stamp should be valid")
	}
}

func TestStampReplacesExistingStamp(t *testing.T) {
	path := writeTempFeature(t, Prefix+"stale\nFeature: Smoke\n")

	if err := Stamp(path); err != nil {
		t.Fatal(err)
	}

	content := readTempFeature(t, path)
	if strings.Contains(content, "stale") {
		t.Fatalf("stale stamp was not replaced:\n%s", content)
	}
	if !Valid(path) {
		t.Fatal("replacement stamp should be valid")
	}
}

func TestRemoveDeletesOnlyFirstStampLine(t *testing.T) {
	content := "Feature: Smoke\n" + Prefix + "old\nScenario: one\n" + Prefix + "kept\n"
	path := writeTempFeature(t, content)

	if err := Remove(path); err != nil {
		t.Fatal(err)
	}

	want := "Feature: Smoke\nScenario: one\n" + Prefix + "kept\n"
	if got := readTempFeature(t, path); got != want {
		t.Fatalf("content = %q, want %q", got, want)
	}
}

func TestValidRejectsMissingAndStaleStamps(t *testing.T) {
	unstamped := writeTempFeature(t, "Feature: Smoke\n")
	if Valid(unstamped) {
		t.Fatal("unstamped file should be invalid")
	}

	stale := writeTempFeature(t, Prefix+"stale\nFeature: Smoke\n")
	if Valid(stale) {
		t.Fatal("stale stamp should be invalid")
	}
}

func TestSplitReturnsStampAndUnstampedContent(t *testing.T) {
	stamp, unstamped := Split("Feature: Smoke\r\n" + Prefix + "abc\r\nScenario: one\r\n")

	if stamp != "abc" {
		t.Fatalf("stamp = %q", stamp)
	}
	if unstamped != "Feature: Smoke\r\nScenario: one\r\n" {
		t.Fatalf("unstamped = %q", unstamped)
	}
}

func writeTempFeature(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "feature.feature")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func readTempFeature(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
