package helmcommon

import (
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

func PrepareKubeConfig(kubeConfigFile string, environmentName string, env map[string]string) error {
	var possibleKeys []string

	if environmentName != "" {
		possibleKeys = append(possibleKeys, fmt.Sprintf("KUBECONFIG_%s_BASE64", strings.ToUpper(environmentName)))
	}
	possibleKeys = append(possibleKeys, "KUBECONFIG_BASE64")

	for _, key := range possibleKeys {
		kubeConfigContent, err := GetDecodedEnvVar(env, key)
		if err == nil {
			err = os.MkdirAll(path.Dir(kubeConfigFile), 0755)
			if err != nil {
				return err
			}

			return os.WriteFile(kubeConfigFile, []byte(kubeConfigContent), 0644)
		}
	}

	return fmt.Errorf("no valid kubeconfig Base64 environment variable found - checked %v", possibleKeys)
}

type KubeConfig struct {
	CurrentContext string    `yaml:"current-context"`
	Clusters       []Cluster `yaml:"clusters"`
	Contexts       []Context `yaml:"contexts"`
}

type Cluster struct {
	Name    string      `yaml:"name"`
	Cluster ClusterData `yaml:"cluster"`
}

type ClusterData struct {
	Server string `yaml:"server"`
}

type Context struct {
	Name    string      `yaml:"name"`
	Context ContextData `yaml:"context"`
}

type ContextData struct {
	Cluster string `yaml:"cluster"`
}

func ParseKubeConfigCluster(kubeconfigPath string) (Cluster, error) {
	// read file
	data, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return Cluster{}, fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	// unmarshal
	var config KubeConfig
	if err = yaml.Unmarshal(data, &config); err != nil {
		return Cluster{}, fmt.Errorf("failed to parse kubeconfig YAML: %w", err)
	}

	// get cluster from current context
	var clusterName string
	for _, ctx := range config.Contexts {
		if ctx.Name == config.CurrentContext {
			clusterName = ctx.Context.Cluster
			break
		}
	}
	if clusterName == "" {
		return Cluster{}, fmt.Errorf("current context not found")
	}

	// get cluster data
	var cluster Cluster
	for _, c := range config.Clusters {
		if c.Name == clusterName {
			cluster = c
			break
		}
	}
	if cluster.Name == "" {
		return Cluster{}, fmt.Errorf("cluster not found")
	}

	return cluster, nil
}

func GetDecodedEnvVar(env map[string]string, varName string) (string, error) {
	value, exists := env[varName]
	if !exists || value == "" {
		return "", fmt.Errorf("environment variable %s is not set or empty", varName)
	}

	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", fmt.Errorf("failed to decode Base64 for %s: %w", varName, err)
	}

	return string(decoded), nil
}
