package helm

import (
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var netClient = &http.Client{
	Timeout: time.Second * 10,
}

type ChartConfig struct {
	ApiVersion  string `yaml:"apiVersion"`
	AppVersion  string `yaml:"appVersion"`
	Description string `yaml:"description"`
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Deprecated  bool   `yaml:"deprecated"`
	Maintainers []struct {
		Name  string `yaml:"name"`
		EMail string `yaml:"email"`
		Url   string `yaml:"url"`
	} `yaml:"maintainers"`
	Dependencies []struct {
		Name       string `yaml:"name"`
		Version    string `yaml:"version"`
		Repository string `yaml:"repository"`
	} `yaml:"dependencies"`
}

func UploadChart(url string, username string, password string, file string) (string, []byte) {
	data, err := os.Open(file)
	if err != nil {
		log.Err(err).Str("file", file).Msg("failed to get file content")
	}
	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		log.Err(err).Msg("failed to prepare request")
	}
	req.SetBasicAuth(username, password)
	resp, err := netClient.Do(req)
	if err != nil {
		log.Err(err).Msg("failed to upload chart")
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Err(err).Msg("failed to parse response")
	}
	return resp.Status, content
}

func ParseChart(file string) *ChartConfig {
	content, contentErr := ioutil.ReadFile(file)
	if contentErr != nil {
		log.Err(contentErr).Str("file", file).Msg("failed to get file content")
		return nil
	}

	var chart ChartConfig
	parseErr := yaml.Unmarshal(content, &chart)
	if parseErr != nil {
		log.Err(parseErr).Str("file", file).Msg("failed to parse Chart.yaml")
		return nil
	}

	return &chart
}
