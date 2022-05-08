package helm

import (
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	"mime/multipart"
	"os"
	"io"
	"bytes"
	"strings"
	"time"
	"strconv"
)

var netClient = &http.Client{
	Timeout: time.Second * 10,
}

type ChartConfig struct {
	ApiVersion  string `yaml:"apiVersion"`
	AppVersion  string `yaml:"appVersion"`
	KubeVersion string `yaml:"kubeVersion"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Name        string `yaml:"name"`
	Deprecated  bool   `yaml:"deprecated"`
	Keywords    []string `yaml:"keywords"`
	Home string `yaml:"home"`
	Icon string `yaml:"icon"`
	Annotations map[string]string `yaml:"annotations"`
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

func UploadChart(url string, username string, password string, file string) (int, []byte) {
	// prepare
	contentType, body, err := createForm(map[string]string{"file": "@"+file})
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

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Err(err).Msg("failed to parse response")
	}

	// resp code
	responseCodeInt, _ := strconv.Atoi(resp.Status)

	return responseCodeInt, content
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

func createForm(form map[string]string) (string, io.Reader, error) {
	body := new(bytes.Buffer)
	mp := multipart.NewWriter(body)
	defer mp.Close()
	for key, val := range form {
	   if strings.HasPrefix(val, "@") {
		  val = val[1:]
		  file, err := os.Open(val)
		  if err != nil { return "", nil, err }
		  defer file.Close()
		  part, err := mp.CreateFormFile(key, val)
		  if err != nil { return "", nil, err }
		  io.Copy(part, file)
	   } else {
		  mp.WriteField(key, val)
	   }
	}
	return mp.FormDataContentType(), body, nil
}
