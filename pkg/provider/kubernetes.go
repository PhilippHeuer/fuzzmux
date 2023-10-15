package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PhilippHeuer/tmux-tms/pkg/config"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesProvider struct {
	Clusters []config.KubernetesCluster
}

func (p KubernetesProvider) Name() string {
	return "kubernetes"
}

func (p KubernetesProvider) Options() ([]Option, error) {
	var options []Option

	// get home directory
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", homeErr)
	}

	for _, cluster := range p.Clusters {
		// read config
		config, err := clientcmd.BuildConfigFromFlags("", cluster.KubeConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
		}

		// create client
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create client from config: %w", err)
		}

		// list namespaces
		list, err := client.CoreV1().Namespaces().List(context.Background(), v1.ListOptions{})
		if err != nil {
			return nil, err
		}

		// add namespaces
		for _, item := range list.Items {
			displayName := item.GetName()
			if cluster.Name != "" {
				displayName = fmt.Sprintf("%s @ %s", displayName, cluster.Name)
			}

			// add option
			options = append(options, Option{
				ProviderName:   p.Name(),
				Id:             item.GetName(),
				DisplayName:    displayName,
				Name:           item.GetName(),
				StartDirectory: filepath.Join(homeDir, "fuzzmux", "k8s", item.GetName()),
				Tags:           cluster.Tags,
				Context: map[string]string{
					"kubeconfig": cluster.KubeConfig,
					"namespace":  item.GetName(),
				},
			})
		}
	}

	return options, nil
}

func (p KubernetesProvider) OptionsOrCache(maxAge float64) ([]Option, error) {
	options, err := LoadOptions(p.Name(), 0)
	if err == nil {
		return options, nil
	}

	options, err = p.Options()
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}

	err = SaveOptions(p.Name(), options)
	if err != nil {
		log.Warn().Err(err).Msg("failed to save options to cache")
	}

	return options, nil
}

func (p KubernetesProvider) SelectOption(option *Option) error {
	// create startDirectory
	if _, err := os.Stat(option.StartDirectory); os.IsNotExist(err) {
		err = os.MkdirAll(option.StartDirectory, 0755)
		if err != nil {
			return fmt.Errorf("failed to create start directory: %w", err)
		}
	}

	return nil
}
