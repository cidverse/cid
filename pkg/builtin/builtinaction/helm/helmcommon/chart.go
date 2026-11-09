package helmcommon

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.yaml.in/yaml/v4"
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
func UploadChart(url string, username string, password string, file string) (responseCode int, responseContent []byte, err error) {
	// prepare
	contentType, body, err := createForm(map[string]string{"file": "@" + file})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to prepare request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	// auth
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", contentType)

	// execute
	resp, err := netClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to upload chart: %w", err)
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to read response: %w", err)
	}

	return resp.StatusCode, content, nil
}

// ChartArtifactQuery returns the artifact query for the packaged charts of a module
func ChartArtifactQuery(moduleSlug string) string {
	return fmt.Sprintf(`artifact_type == "helm-chart" && format == "tgz" && module == %q`, moduleSlug)
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
				return "", nil, partErr
			}
			_, _ = io.Copy(part, file)
		} else {
			_ = mp.WriteField(key, val)
		}
	}
	return mp.FormDataContentType(), body, nil
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
