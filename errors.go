package tinyrbac

import (
	"errors"
	"fmt"
)

var (
	ErrConfigFileNotProvided = errors.New("config file path is empty")
)

func errConfigNotFound(filetype, path string, err error) error {
	return fmt.Errorf("open %s config %q: %w", filetype, path, err)
}

func errConfigRead(filetype, path string, err error) error {
	return fmt.Errorf("read %s config %q: %w", filetype, path, err)
}

func errConfigUnmarshal(filetype, path string, err error) error {
	return fmt.Errorf("unmarshal %s config %q: %w", filetype, path, err)
}
