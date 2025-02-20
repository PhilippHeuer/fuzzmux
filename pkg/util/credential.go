package util

import (
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

func ResolvePasswordValue(value string) string {
	if strings.HasPrefix(value, "env:") {
		return os.Getenv(strings.TrimPrefix(value, "env:"))
	} else if strings.HasPrefix(value, "file:") {
		file := strings.TrimPrefix(value, "file:")
		file = strings.Replace(file, "~", os.Getenv("HOME"), 1)
		file = os.ExpandEnv(file)

		bytes, err := os.ReadFile(file)
		if err != nil {
			return ""
		}
		return string(bytes)
	} else if strings.HasPrefix(value, "pass:") {
		secretPath := os.ExpandEnv(strings.TrimPrefix(value, "pass:"))

		cmd := exec.Command("pass", "show", secretPath)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		out, err := cmd.Output()
		if err != nil {
			log.Fatal().Err(err).Str("secretPath", secretPath).Msg("failed to execute pass command")
		}

		return strings.TrimSpace(string(out))
	} else if strings.HasPrefix(value, "cmd:") && os.Getenv("FUZZMUX_ALLOW_COMMANDS") == "true" {
		cmd := exec.Command("sh", "-c", strings.TrimPrefix(value, "cmd:"))
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		out, err := cmd.Output()
		if err != nil {
			log.Fatal().Err(err).Str("command", value).Msg("failed to execute password command")
		}

		return strings.TrimSpace(string(out))
	}

	return value
}
