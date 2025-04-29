package helmcommon

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var netClient = &http.Client{
	Timeout: time.Second * 10,
}

type ChartConfig struct {
	APIVersion  string            `yaml:"apiVersion"`
	AppVersion  string            `yaml:"appVersion"`
	KubeVersion string            `yaml:"kubeVersion"`
	Version     string            `yaml:"version"`
	Description string            `yaml:"description"`
	Name        string            `yaml:"name"`
	Deprecated  bool              `yaml:"deprecated"`
	Keywords    []string          `yaml:"keywords"`
	Home        string            `yaml:"home"`
	Icon        string            `yaml:"icon"`
	Annotations map[string]string `yaml:"annotations"`
	Maintainers []struct {
		Name  string `yaml:"name"`
		EMail string `yaml:"email"`
		URL   string `yaml:"url"`
	} `yaml:"maintainers"`
	Dependencies []struct {
		Name       string `yaml:"name"`
		Version    string `yaml:"version"`
		Repository string `yaml:"repository"`
	} `yaml:"dependencies"`
}

// UploadChart will upload the chart to a nexus repository
func UploadChart(url string, username string, password string, file string) (responseCode int, responseContent []byte) {
	// prepare
	contentType, body, err := createForm(map[string]string{"file": "@" + file})
	if err != nil {
		//log.Err(err).Msg("failed to prepare request body")
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		//log.Err(err).Msg("failed to prepare request")
	}

	// auth
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", contentType)

	// execute
	resp, err := netClient.Do(req)
	if err != nil {
		//log.Err(err).Msg("failed to upload chart")
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		//log.Err(err).Msg("failed to parse response")
	}

	// resp code
	responseCodeInt, _ := strconv.Atoi(extractNumbers(resp.Status))

	return responseCodeInt, content
}

func ParseChart(content []byte) (ChartConfig, error) {
	var chart ChartConfig
	err := yaml.Unmarshal(content, &chart)
	if err != nil {
		return ChartConfig{}, fmt.Errorf("failed to parse chart file: %s", err.Error())
	}

	return chart, nil
}

func createForm(form map[string]string) (string, io.Reader, error) {
	body := new(bytes.Buffer)
	mp := multipart.NewWriter(body)
	defer mp.Close()
	for key, val := range form {
		if strings.HasPrefix(val, "@") {
			val = val[1:]
			file, err := os.Open(val)
			if err != nil {
				return "", nil, err
			}
			defer file.Close() //nolint
			part, partErr := mp.CreateFormFile(key, val)
			if partErr != nil {
				return "", nil, err
			}
			_, _ = io.Copy(part, file)
		} else {
			_ = mp.WriteField(key, val)
		}
	}
	return mp.FormDataContentType(), body, nil
}

func extractNumbers(input string) string {
	reg := regexp.MustCompile(`\D+`)
	return reg.ReplaceAllString(input, "")
}

type ChartSource string

const (
	ChartSourceLocal      ChartSource = "local"
	ChartSourceRepository ChartSource = "repo"
	ChartSourceOCI        ChartSource = "oci"
	ChartSourceUnknown    ChartSource = "unknown"
)

func GetChartSource(input string) ChartSource {
	if strings.HasPrefix(input, "oci://") {
		return ChartSourceOCI
	} else if strings.Count(input, "/") == 1 {
		return ChartSourceRepository
	} else if isHelmChartDir(input) {
		return ChartSourceLocal
	}

	return ChartSourceUnknown
}

func isHelmChartDir(path string) bool {
	path, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	stat, err := os.Stat(path)
	if err != nil || !stat.IsDir() {
		return false
	}
	chartPath := filepath.Join(path, "Chart.yaml")
	_, err = os.Stat(chartPath)
	return err == nil
}
