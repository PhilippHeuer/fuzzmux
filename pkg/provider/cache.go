package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/PhilippHeuer/tmux-tms/pkg/core/util"
)

var configDir = filepath.Join(util.GetAppDataDir(), "fuzzmux")

type OptionsCache struct {
	ProviderName string
	Options      []Option
	CreatedAt    time.Time
}

func SaveOptions(providerName string, options []Option) error {
	jsonData, err := json.MarshalIndent(OptionsCache{
		ProviderName: providerName,
		Options:      options,
		CreatedAt:    time.Now(),
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal options: %w", err)
	}

	err = os.WriteFile(filepath.Join(configDir, fmt.Sprintf("provider-%s.json", providerName)), jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write options: %w", err)
	}

	return nil
}

func LoadOptions(providerName string, maxAge float64) ([]Option, error) {
	var optionsCache OptionsCache

	jsonData, err := os.ReadFile(filepath.Join(configDir, fmt.Sprintf("provider-%s.json", providerName)))
	if err != nil {
		return nil, fmt.Errorf("failed to read options: %w", err)
	}

	err = json.Unmarshal(jsonData, &optionsCache)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal options: %w", err)
	}

	if time.Since(optionsCache.CreatedAt).Seconds() > maxAge {
		return nil, fmt.Errorf("cache is too old")
	}

	return optionsCache.Options, nil
}
