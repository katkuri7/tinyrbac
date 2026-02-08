package tinyrbac

import (
	"fmt"
	"slices"
)

// The access information is stored as follows:
// roles	: r1, r2, r3
// resources: R1, R2, R3
// actions  : A1, A2, A3
//
// ┌────────────────────────────────────────────────────┐
// │ resource  R1R2R3  R1R2R3  R1R2R3  R1R2R3  R1R2R3   │
// │ access   [0 1 1] [0 0 1] [0 0 0] [0 0 0] [1 1 1]   │
// │ action    --A1--  --A2--  --A3--  --A4--  --A5--   │
// │ role      ----------------r1--------------------   │
// └────────────────────────────────────────────────────┘
//
// Interpreted as:
// Role r1 has A1 access for resources R2 and R3.
// Role r1 has A2 access for resource R3 only.
// Role r1 does not have A3 and A4 accesses on any of the resources.
// Role r1 has A5 access for all the resources.

type resourceSet uint64

type Rbac struct {
	accessMap      [maxActions * maxRoles]resourceSet
	roleIdxMap     [maxRoles]string
	resourceIdxMap [maxResources]string
}

// NewFromJsonConfig creates an RBAC instance from a JSON config
// file at the given path. An error is returned when the config
// file cannot be proccessed.
func NewFromJsonConfig(path string) (*Rbac, error) {
	c, err := newConfigFromJson(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return buildFromConfig(c)
}

// NewFromJsonConfig creates an RBAC instance from a YAML config
// file at the given path. An error is returned when the config
// file cannot be proccessed.
func NewFromYamlConfig(path string) (*Rbac, error) {
	c, err := newConfigFromYaml(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return buildFromConfig(c)
}

// buildRoleAndResourceMapping extracts roles and resources from config.
// The extracted information is stored in a sorted manner which allows for
// the core idea of using the role-index and resource-index mapping to perform rbac operations.
func buildRoleAndResourceMapping(c *config, r *Rbac) {
	resources := make(map[string]bool)
	for _, r := range c.Resources {
		resources[r] = true
	}

	i := 0
	for resource := range resources {
		r.resourceIdxMap[i] = resource
		i++
	}
	// Sorting because Go maps do not store/return data in an ordered fashion.
	slices.Sort(r.resourceIdxMap[:i])

	// Config validation makes sure roles are unique. So a map
	// filtering is not needed.
	i = 0
	for _, role := range c.Roles {
		r.roleIdxMap[i] = role.Name
		i++
	}

	// The [:] syntax returns a slice header that points to actual array data.
	// So a sort on this slice ultimately sorts our fixed size array.
	// We are concerned with only the first 'i' elements because performing a sort
	// on the entire array may result in the untouched elements (empty strings) accumulating in the beginning.
	slices.Sort(r.roleIdxMap[:i])
}

// buildFromConfig builds the actual access map from config.
func buildFromConfig(c *config) (*Rbac, error) {
	r := &Rbac{}
	buildRoleAndResourceMapping(c, r)

	for _, role := range c.Roles {
		accessIdx := slices.Index(r.roleIdxMap[:], role.Name) * maxActions
		for _, resource := range role.Resources {
			// If no actions are provided for a resource it can be ignored.
			// TODO: Should this be moved to config validation?
			actions := slices.DeleteFunc(resource.Actions, func(a string) bool {
				return a == ""
			})
			if len(actions) == 0 {
				continue
			}

			if resource.Name == allResources {
				for _, action := range actions {
					r.accessMap[accessIdx+getHTTPActionOffset(action)] = allResourceAccess
				}
			} else {
				resourceIdx := slices.Index(r.resourceIdxMap[:], resource.Name)
				for _, action := range actions {
					r.accessMap[accessIdx+getHTTPActionOffset(action)] |= 1 << resourceIdx
				}
			}
		}
	}

	return r, nil
}

// TODO: Justify linearly searching instead of using a hash map.
func (r *Rbac) check(role, resource, action string) (bool, error) {
	roleIdx, resourceIdx := -1, -1
	for idx, roleName := range r.roleIdxMap {
		if roleName == role {
			roleIdx = idx
			break
		}
	}
	if roleIdx == -1 {
		return false, fmt.Errorf("unknown role: %s", role)
	}

	for idx, resourceName := range r.resourceIdxMap {
		if resourceName == resource {
			resourceIdx = idx
			break
		}
	}
	if resourceIdx == -1 {
		return false, fmt.Errorf("unknown resource: %s", resource)
	}

	accessIdx := roleIdx*maxActions + getHTTPActionOffset(action)
	return r.accessMap[accessIdx]&resourceSet(1<<resourceIdx) != 0, nil
}

// Check returns (true, nil) if 'role' has access to perform 'action' on 'resource'
// and (false, nil) otheriwse. In case of an error false is returned along with the error.
func (r *Rbac) Check(role, resource, action string) (bool, error) {
	return r.check(role, resource, action)
}
