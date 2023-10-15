package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PhilippHeuer/tmux-tms/pkg/config"
	projectsv1 "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type OpenShiftProvider struct {
	Clusters []config.KubernetesCluster
}

func (p OpenShiftProvider) Name() string {
	return "openshift"
}

func (p OpenShiftProvider) Options() ([]Option, error) {
	var options []Option

	// get home directory
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", homeErr)
	}

	for _, cluster := range p.Clusters {
		clusterName := "default"
		if cluster.Name != "" {
			clusterName = cluster.Name
		}

		// read config
		conf, err := clientcmd.BuildConfigFromFlags("", cluster.KubeConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
		}

		// create client
		client, err := projectsv1.NewForConfig(conf)
		if err != nil {
			return nil, fmt.Errorf("failed to create client from config: %w", err)
		}

		// list namespaces
		list, err := client.Projects().List(context.Background(), v1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, item := range list.Items {
			displayName := item.GetName()
			if item.Annotations["openshift.io/display-name"] != "" {
				displayName = fmt.Sprintf("[%s] %s", item.GetName(), item.Annotations["openshift.io/display-name"])
			}
			if cluster.Name != "" {
				displayName = fmt.Sprintf("%s @ %s", displayName, cluster.Name)
			}

			description := ""
			if item.Annotations["openshift.io/description"] != "" {
				description = item.Annotations["openshift.io/description"]
			}

			// add option
			options = append(options, Option{
				ProviderName:   p.Name(),
				Id:             item.GetName(),
				DisplayName:    displayName,
				Name:           item.GetName(),
				StartDirectory: filepath.Join(homeDir, "fuzzmux", "k8s", clusterName, item.GetName()),
				Tags:           cluster.Tags,
				Context: map[string]string{
					"clusterName": clusterName,
					"clusterHost": conf.Host,
					"clusterUser": conf.Username,
					"kubeConfig":  cluster.KubeConfig,
					"namespace":   item.GetName(),
					"description": description,
				},
			})
		}
	}

	return options, nil
}

func (p OpenShiftProvider) OptionsOrCache(maxAge float64) ([]Option, error) {
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

func (p OpenShiftProvider) SelectOption(option *Option) error {
	// create startDirectory
	if _, err := os.Stat(option.StartDirectory); os.IsNotExist(err) {
		err = os.MkdirAll(option.StartDirectory, 0755)
		if err != nil {
			return fmt.Errorf("failed to create start directory: %w", err)
		}
	}

	return nil
}
