package helm

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
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
		log.Err(err).Msg("failed to prepare request body")
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Err(err).Msg("failed to prepare request")
	}

	// auth
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", contentType)

	// execute
	resp, err := netClient.Do(req)
	if err != nil {
		log.Err(err).Msg("failed to upload chart")
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Err(err).Msg("failed to parse response")
	}

	// resp code
	responseCodeInt, _ := strconv.Atoi(extractNumbers(resp.Status))

	return responseCodeInt, content
}

func ParseChart(file string) *ChartConfig {
	content, contentErr := os.ReadFile(file)
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
