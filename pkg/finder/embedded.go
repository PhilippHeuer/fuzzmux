package finder

import (
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/ktr0731/go-fuzzyfinder"
)

func FuzzyFinderEmbedded(options []recon.Option, cfg config.FinderConfig) (recon.Option, error) {
	var fOptions = []fuzzyfinder.Option{
		fuzzyfinder.WithCursorPosition(fuzzyfinder.CursorPositionBottom),
	}

	if cfg.Preview {
		fOptions = append(fOptions, fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return options[i].RenderPreview()
		}))
	}

	idx, err := fuzzyfinder.Find(
		options,
		func(i int) string {
			return options[i].DisplayName
		},
		fOptions...,
	)
	if err != nil {
		return recon.Option{}, fmt.Errorf("failed to find option: %w", err)
	}

	return options[idx], nil
}
