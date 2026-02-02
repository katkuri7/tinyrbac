package tinyrbac

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

const yamlConfigFiletype = "yaml"

func readFromYaml(path string) (*config, error) {
	if path == "" {
		return nil, ErrConfigFileNotProvided
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, errConfigNotFound(yamlConfigFiletype, path, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, errConfigRead(yamlConfigFiletype, path, err)
	}

	c := config{}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, errConfigUnmarshal(yamlConfigFiletype, path, err)
	}

	return &c, nil
}
