package tinyrbac

import (
	"encoding/json"
	"io"
	"os"
)

const jsonConfigFiletype = "json"

func readFromJson(path string) (*config, error) {
	if path == "" {
		return nil, ErrConfigFileNotProvided
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, errConfigNotFound(jsonConfigFiletype, path, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, errConfigRead(jsonConfigFiletype, path, err)
	}

	var c config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, errConfigUnmarshal(jsonConfigFiletype, path, err)
	}

	return &c, nil
}
