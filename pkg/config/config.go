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

type SkipOutOfBoundsBlocksConfig struct {
	// Should blocks that are not visible be skipped during encoding?
	Enable bool `yaml:"Enable"`
}

type DeduplicateBlocksConfig struct {
	// Should similar blocks be deduplicated during encoding?
	Enable bool `yaml:"Enable"`
	// How similar do blocks have to be to be deduplicated
	MinimalSimilarity float64 `yaml:"MinimalSimilarity"`
}

type EncodingConfig struct {
	//Underlying archive format of the encoded file
	// TODO: Document accepted values
	ArchiveFormat         string                      `yaml:"ArchiveFormat"`
	SkipOutOfBoundsBlocks SkipOutOfBoundsBlocksConfig `yaml:"SkipOutOfBoundsBlocks"`
	DeduplicateBlocks     DeduplicateBlocksConfig     `yaml:"DeduplicateBlocks"`
}

type VisualizationConfig struct {
	// Should the visualizations be created?
	Enable bool `yaml:"Enable"`
}

// Config holds parameters that influence the partitioning and encoding process of the quadtree
type Config struct {
	Quadtree            QuadtreeConfig      `yaml:"Quadtree"`
	Encoding            EncodingConfig      `yaml:"Encoding"`
	VisualizationConfig VisualizationConfig `yaml:"Visualization"`
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
