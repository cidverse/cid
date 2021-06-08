package changelog

import "embed"

//go:embed templates/*
var TemplateFS embed.FS

type Config struct {
	Templates    []string          `yaml:"templates"`
	TitleMaps    map[string]string `yaml:"title_maps"`
	NoteKeywords []NoteKeyword     `yaml:"note_keywords"`
	IssuePrefix  string            `yaml:"issue_prefix"`
}

type NoteKeyword struct {
	Keyword string
	Title   string
}
