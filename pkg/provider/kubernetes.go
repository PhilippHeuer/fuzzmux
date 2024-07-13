package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/core/util"
	projectsv1 "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
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

	for _, cluster := range p.Clusters {
		if cluster.OpenShift {
			opts, err := processOpenShiftCluster(cluster)
			if err != nil {
				return nil, err
			}
			options = append(options, opts...)
			continue
		}

		opts, err := processKubernetesCluster(cluster)
		if err != nil {
			return nil, err
		}
		options = append(options, opts...)
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
	if option.StartDirectory == "" {
		return nil
	}

	// create startDirectory
	if _, err := os.Stat(option.StartDirectory); os.IsNotExist(err) {
		err = os.MkdirAll(option.StartDirectory, 0755)
		if err != nil {
			return errors.Join(ErrFailedToCreateStartDirectory, err)
		}
	}

	return nil
}

func processKubernetesCluster(cluster config.KubernetesCluster) (options []Option, err error) {
	clusterName := "default"
	if cluster.Name != "" {
		clusterName = cluster.Name
	}

	// file exists?
	configFile := util.ResolvePath(cluster.KubeConfig)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig file does not exist: %w", err)
	}

	// read config
	conf, err := clientcmd.BuildConfigFromFlags("", configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	// create client
	client, err := kubernetes.NewForConfig(conf)
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
			ProviderName:   "kubernetes",
			Id:             item.GetName(),
			DisplayName:    displayName,
			Name:           item.GetName(),
			StartDirectory: filepath.Join(util.GetHomeDir(), "fuzzmux", "k8s", clusterName, item.GetName()),
			Tags:           cluster.Tags,
			Context: map[string]string{
				"clusterName": clusterName,
				"clusterHost": conf.Host,
				"clusterUser": conf.Username,
				"clusterType": "kubernetes",
				"kubeConfig":  configFile,
				"namespace":   item.GetName(),
			},
		})
	}

	return options, nil
}

func processOpenShiftCluster(cluster config.KubernetesCluster) (options []Option, err error) {
	clusterName := "default"
	if cluster.Name != "" {
		clusterName = cluster.Name
	}

	// file exists?
	configFile := util.ResolvePath(cluster.KubeConfig)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig file does not exist: %w", err)
	}

	// read config
	conf, err := clientcmd.BuildConfigFromFlags("", configFile)
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
			ProviderName:   "kubernetes",
			Id:             item.GetName(),
			DisplayName:    displayName,
			Name:           item.GetName(),
			StartDirectory: filepath.Join(util.GetHomeDir(), "fuzzmux", "k8s", clusterName, item.GetName()),
			Tags:           cluster.Tags,
			Context: map[string]string{
				"clusterName": clusterName,
				"clusterHost": conf.Host,
				"clusterUser": conf.Username,
				"clusterType": "openshift",
				"kubeConfig":  configFile,
				"namespace":   item.GetName(),
				"description": description,
			},
		})
	}

	return options, nil
}
