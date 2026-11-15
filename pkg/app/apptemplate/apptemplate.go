package apptemplate

import (
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
)

//go:embed templates/*
var embedFS embed.FS

type TemplateData struct {
	Version         string // CLI Version
	VersionFileHash string // CLI Version - SHA256 File Hash
}

// Render renders a template and returns its content
func Render(templateFile string, data TemplateData) ([]byte, error) {
	content, err := embedFS.ReadFile(path.Join("templates", templateFile))
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow template %s: %w", templateFile, err)
	}

	rendered, err := vcsapp.Render(string(content), data)
	if err != nil {
		return nil, fmt.Errorf("failed to render template %s: %w", templateFile, err)
	}

	return rendered, nil
}

// RenderFile renders a template file
func RenderFile(templateFile string, data TemplateData, outputFile string) error {
	template, err := Render(templateFile, data)
	if err != nil {
		return err
	}

	// create file
	err = os.MkdirAll(filepath.Dir(outputFile), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create workflow file parent directory: %w", err)
	}
	err = os.WriteFile(outputFile, template, 0644)
	if err != nil {
		return fmt.Errorf("failed to create workflow file: %w", err)
	}

	return nil
}

func NewTemplateData(conf appconfig.Config) TemplateData {
	return TemplateData{
		Version:         conf.Version,
		VersionFileHash: conf.VersionHash,
	}
}
