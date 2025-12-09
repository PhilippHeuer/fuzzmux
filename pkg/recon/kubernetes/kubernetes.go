package kubernetes

import (
	"context"
	"fmt"
	"os"

	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
	projectsv1 "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const moduleType = "kubernetes"
const defaultStartDirectory = "~/k8s/{{clusterName}}/{{namespace}}"

type Module struct {
	Config ModuleConfig
}

type ModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

	// DisplayName is a template string to render a custom display name
	DisplayName string `yaml:"display-name"`

	// StartDirectory is a template string that defines the start directory
	StartDirectory string `yaml:"start-directory"`

	// Clusters is a list of kubernetes clusters that should be scanned
	Clusters []KubernetesCluster `yaml:"clusters"`
}

type KubernetesCluster struct {
	// Name of the cluster
	Name string `yaml:"name"`

	// Tags that apply to the cluster
	Tags []string `yaml:"tags"`

	// OpenShift indicates if this is an OpenShift cluster (default: false)
	OpenShift bool `yaml:"openshift"`

	// KubeConfig is the absolute path to the kubeconfig file
	KubeConfig string `yaml:"kubeconfig"`
}

func (p Module) Name() string {
	if p.Config.Name != "" {
		return p.Config.Name
	}
	return moduleType
}

func (p Module) Type() string {
	return moduleType
}

func (p Module) Options() ([]recon.Option, error) {
	var options []recon.Option

	for _, cluster := range p.Config.Clusters {
		if cluster.OpenShift {
			opts, err := processOpenShiftCluster(cluster, p.Name(), p.Config)
			if err != nil {
				return nil, err
			}
			options = append(options, opts...)
			continue
		}

		opts, err := processKubernetesCluster(cluster, p.Name(), p.Config)
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

func NewModule(config ModuleConfig) Module {
	return Module{
		Config: config,
	}
}

func processKubernetesCluster(cluster KubernetesCluster, moduleName string, moduleConf ModuleConfig) (result []recon.Option, err error) {
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
		opt := recon.Option{
			ProviderName:   moduleName,
			ProviderType:   moduleType,
			Id:             item.GetName(),
			DisplayName:    displayName,
			Name:           item.GetName(),
			StartDirectory: defaultStartDirectory,
			Tags:           cluster.Tags,
			Context: map[string]string{
				"clusterName": clusterName,
				"clusterHost": conf.Host,
				"clusterUser": conf.Username,
				"clusterType": "kubernetes",
				"kubeConfig":  configFile,
				"namespace":   item.GetName(),
			},
		}
		opt.ProcessUserTemplateStrings(moduleConf.DisplayName, moduleConf.StartDirectory)
		result = append(result, opt)
	}

	return result, nil
}

func processOpenShiftCluster(cluster KubernetesCluster, moduleName string, moduleConf ModuleConfig) (result []recon.Option, err error) {
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
		opt := recon.Option{
			ProviderName:   moduleName,
			ProviderType:   moduleType,
			Id:             item.GetName(),
			DisplayName:    displayName,
			Name:           item.GetName(),
			StartDirectory: defaultStartDirectory,
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
		}
		opt.ProcessUserTemplateStrings(moduleConf.DisplayName, moduleConf.StartDirectory)
		result = append(result, opt)
	}

	return result, nil
}
