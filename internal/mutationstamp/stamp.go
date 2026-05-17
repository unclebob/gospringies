package mutationstamp

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
)

const Prefix = "# mutation-stamp: "

func Valid(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	stamp, unstamped := Split(string(content))
	return stamp != "" && stamp == Hash(unstamped)
}

func Stamp(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, unstamped := Split(string(content))
	return os.WriteFile(path, []byte(Prefix+Hash(unstamped)+"\n"+unstamped), 0o644)
}

func Remove(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, unstamped := Split(string(content))
	return os.WriteFile(path, []byte(unstamped), 0o644)
}

func Split(content string) (string, string) {
	lines := strings.SplitAfter(content, "\n")
	var unstamped strings.Builder
	removed := false
	stamp := ""
	for _, line := range lines {
		if !removed {
			if value, ok := strings.CutPrefix(strings.TrimRight(line, "\r\n"), Prefix); ok {
				stamp = value
				removed = true
				continue
			}
		}
		unstamped.WriteString(line)
	}
	return stamp, unstamped.String()
}

func Hash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}
