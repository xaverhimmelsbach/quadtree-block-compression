package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type QuadtreeConfig struct {
	// Minimal similarity of base and upsampled image required to be a leaf
	SimilarityCutoff float64 `yaml:"SimilarityCutoff"`
	// Interpolation algorithm used to downsample base image
	DownsamplingInterpolator string `yaml:"DownsamplingInterpolator"`
	// Interpolation algorithm used to upsample downsampled image
	UpsamplingInterpolator string `yaml:"UpsamplingInterpolator"`
}

// Config holds parameters that influence the partitioning and encoding process of the quadtree
type Config struct {
	Quadtree QuadtreeConfig `yaml:"Quadtree"`
}

// NewConfigFromFile constructs a Config object from a YAML file
func NewConfigFromFile(path string) (*Config, error) {
	cfgBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewConfigFromBytes(cfgBytes)
}

// NewConfigFromBytes constructs a Config object from a YAML string
func NewConfigFromBytes(cfgBytes []byte) (*Config, error) {
	cfg := new(Config)
	err := yaml.Unmarshal(cfgBytes, cfg)
	return cfg, err
}
