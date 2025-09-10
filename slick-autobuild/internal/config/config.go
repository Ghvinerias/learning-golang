package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Root is the top-level configuration structure for the build tool.
type Root struct {
	Runtime RuntimeConfig   `yaml:"runtime"`
	Matrix  []MatrixEntry   `yaml:"matrix"`
	Defaults DefaultSection `yaml:"defaults"`
}

type RuntimeConfig struct {
	Dotnet VersionSet `yaml:"dotnet"`
	Node   VersionSet `yaml:"node"`
}

type VersionSet struct {
	Versions []string `yaml:"versions"`
}

type MatrixEntry struct {
	Path          string   `yaml:"path"`
	Type          string   `yaml:"type"`
	Frameworks    []string `yaml:"frameworks"` // dotnet specific (SDK versions override)
	NodeVersions  []string `yaml:"nodeVersions"`
	PackageManager string  `yaml:"packageManager"`
	BuildScripts  []string `yaml:"buildScripts"`
	Docker        *DockerConfig `yaml:"docker,omitempty"`
}

type DockerConfig struct {
	Enabled    bool     `yaml:"enabled"`
	Repository string   `yaml:"repository"`
	Tags       []string `yaml:"tags"`
	Push       bool     `yaml:"push"`
	Registries []string `yaml:"registries"`
	Dockerfile string   `yaml:"dockerfile"`
}

type DefaultSection struct {
	Concurrency int    `yaml:"concurrency"`
	ArtifactDir string `yaml:"artifactDir"`
}

// Load reads a YAML config file.
func Load(path string) (*Root, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var r Root
	if err := yaml.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	return &r, nil
}
