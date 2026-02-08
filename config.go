package tinyrbac

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	jsonConfigFiletype = "json"
	yamlConfigFiletype = "yaml"
)

type config struct {
	Description string
	Roles       []role
	Resources   []string
}

type role struct {
	Name        string
	Description string
	Resources   []resource
}

type resource struct {
	Name    string
	Actions []string
}

func newConfigFromJson(path string) (*config, error) {
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

func newConfigFromYaml(path string) (*config, error) {
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

// validate checks if the config fields are valid and consistent.
//
// Validations are done in the below order. An error is returned for the following:
// No resources.
// Resources greater than max resources.
// No roles.
// No role name.
// No resources for a role.
// Undefined resource provided for a role.
// Roles greater than max roles.
// No resource name.
// TODO: Action validation
func (c *config) validate() error {
	if len(c.Resources) == 0 {
		return ErrNoResources
	}

	resources := make(map[string]bool)
	for _, r := range c.Resources {
		if r == "" {
			continue
		}
		resources[r] = true
	}

	if len(resources) > maxResources {
		return fmt.Errorf("resources exceeded: maximum %d but config has %d", maxResources, len(c.Resources))
	}

	if len(c.Roles) == 0 {
		return ErrNoRoles
	}

	// Roles are unique because json unmarshaling
	// will overwrite duplicate entries. Hence, a map
	// filtering is not required unlike resources.
	roleCount := 0
	for i, role := range c.Roles {
		if role.Name == "" {
			return fmt.Errorf("empty role: name not defined at index %d", i)
		}

		if len(role.Resources) == 0 {
			return fmt.Errorf("empty resources: not defined for role %s", role.Name)
		}

		for _, re := range role.Resources {
			if ok := resources[re.Name]; re.Name != allResources && !ok {
				return fmt.Errorf("undefined resource: %s for role %s: %s not defined in resources", re.Name, role.Name, re.Name)
			}
		}

		roleCount++
	}

	if roleCount > maxRoles {
		return fmt.Errorf("roles exceeded: maximum %d but config has %d", maxRoles, len(c.Roles))

	}

	return nil
}
