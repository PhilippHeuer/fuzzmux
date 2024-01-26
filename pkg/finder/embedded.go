package finder

import (
	"fmt"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/provider"
	"github.com/ktr0731/go-fuzzyfinder"
)

func FuzzyFinderEmbedded(options []provider.Option) (*provider.Option, error) {
	idx, err := fuzzyfinder.Find(
		options,
		func(i int) string {
			return options[i].DisplayName
		},
		fuzzyfinder.WithCursorPosition(fuzzyfinder.CursorPositionBottom),
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}

			var builder strings.Builder
			builder.WriteString(options[i].DisplayName + "\n\n")
			builder.WriteString("Provider: " + options[i].ProviderName + "\n")
			if options[i].StartDirectory != "" {
				builder.WriteString("Directory: " + options[i].StartDirectory + "\n")
			}
			if len(options[i].Tags) > 0 {
				builder.WriteString("Tags: " + strings.Join(options[i].Tags, ", ") + "\n")
			}

			// k8s, openshift
			if options[i].Context["clusterName"] != "" {
				builder.WriteString("K8S Cluster Name: " + options[i].Context["clusterName"] + "\n")
			}
			if options[i].Context["clusterHost"] != "" {
				builder.WriteString("K8S Cluster API: " + options[i].Context["clusterHost"] + "\n")
			}
			if options[i].Context["clusterUser"] != "" {
				builder.WriteString("K8S Cluster User: " + options[i].Context["clusterUser"] + "\n")
			}
			if options[i].Context["clusterType"] != "" {
				builder.WriteString("K8S Cluster Type: " + options[i].Context["clusterType"] + "\n")
			}

			// free-text description
			if options[i].Context["description"] != "" {
				builder.WriteString("\n\n" + options[i].Context["description"] + "\n")
			}

			return builder.String()
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find option: %w", err)
	}

	return &options[idx], nil
}
