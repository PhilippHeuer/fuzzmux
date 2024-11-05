package kubernetes

import (
	"context"
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
	"os"
	"path/filepath"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	projectsv1 "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Module struct {
	Config config.KubernetesModuleConfig
}

func (p Module) Name() string {
	return "kubernetes"
}

func (p Module) Options() ([]recon.Option, error) {
	var options []recon.Option

	for _, cluster := range p.Config.Clusters {
		if cluster.OpenShift {
			opts, err := processOpenShiftCluster(cluster)
			if err != nil {
				return nil, err
			}
			options = append(options, opts...)
			continue
		}

		opts, err := processKubernetesCluster(cluster, p.Config.StartDirectory)
		if err != nil {
			return nil, err
		}
		options = append(options, opts...)
	}

	return options, nil
}

func (p Module) OptionsOrCache(maxAge float64) ([]recon.Option, error) {
	options, err := recon.LoadOptions(p.Name(), 0)
	if err == nil {
		return options, nil
	}

	options, err = p.Options()
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}

	err = recon.SaveOptions(p.Name(), options)
	if err != nil {
		log.Warn().Err(err).Msg("failed to save options to cache")
	}

	return options, nil
}

func (p Module) SelectOption(option *recon.Option) error {
	err := option.CreateStartDirectoryIfMissing()
	if err != nil {
		return err
	}

	return nil
}

func (p Module) Columns() []recon.Column {
	return recon.DefaultColumns()
}

func NewModule(config config.KubernetesModuleConfig) Module {
	if config.StartDirectory == "" {
		config.StartDirectory = "~/k8s/{{clusterName}}/{{namespace}}"
	}

	return Module{
		Config: config,
	}
}

func processKubernetesCluster(cluster config.KubernetesCluster, startDirectory string) (options []recon.Option, err error) {
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
		options = append(options, recon.Option{
			ProviderName:   "kubernetes",
			Id:             item.GetName(),
			DisplayName:    displayName,
			Name:           item.GetName(),
			StartDirectory: startDirectory,
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

func processOpenShiftCluster(cluster config.KubernetesCluster) (options []recon.Option, err error) {
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
		options = append(options, recon.Option{
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
