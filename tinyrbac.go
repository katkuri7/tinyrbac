package tinyrbac

import (
	"fmt"
	"math"
	"slices"
)

const (
	// Maximum sizes at compile time to overcome
	// slice header and pointer overhead.
	maxRoles     = 20
	maxActions   = 5 // HTTP
	maxResources = 64

	// roles     = 20
	// actions   = 5 // HTTP
	// resources = 64

	allResources = "*"

	// allResourceAccess is strongly dependent on what the resourceSet type represents.
	allResourceAccess = math.MaxUint64
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

// buildFromConfig builds the actual access map from config.
func buildFromConfig(c *config) (*Rbac, error) {
	r := &Rbac{}
	buildRoleAndResourceNames(c, r)

	for _, role := range c.Roles {
		accessIdx := slices.Index(r.roleIdxMap[:], role.Name) * maxActions
		for _, resource := range role.Resources {
			if resource.Name == allResources {
				for _, action := range resource.Actions {
					r.accessMap[accessIdx+getHTTPActionOffset(action)] = allResourceAccess
				}
			} else {
				resourceIdx := slices.Index(r.resourceIdxMap[:], resource.Name)
				for _, action := range resource.Actions {
					r.accessMap[accessIdx+getHTTPActionOffset(action)] |= 1 << resourceIdx
				}
			}
		}
	}

	return r, nil
}

// NewFromJsonConfig creates an RBAC instance from a JSON config
// file at the given path. An error is returned when the config
// file cannot be proccessed.
func NewFromJsonConfig(path string) (*Rbac, error) {
	c, err := readFromJson(path)
	if err != nil {
		return nil, fmt.Errorf("read config error: %w", err)
	}

	return buildFromConfig(c)
}

// NewFromJsonConfig creates an RBAC instance from a YAML config
// file at the given path. An error is returned when the config
// file cannot be proccessed.
func NewFromYamlConfig(path string) (*Rbac, error) {
	c, err := readFromYaml(path)
	if err != nil {
		return nil, fmt.Errorf("read config error: %w", err)
	}

	return buildFromConfig(c)
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

// Check validates if 'role' has access to perform 'action' on 'resource'.
// It returns false if there is a validation error before access check.
// To get a detailed error use CheckWithError.
func (r *Rbac) Check(role, resource, action string) bool {
	hasAccess, _ := r.check(role, resource, action)
	return hasAccess
}

// CheckWithError validates if 'role' has access to perform 'action' on 'resource'.
// If there is an error, false is returned along with the error.
func (r *Rbac) CheckWithError(role, resource, action string) (bool, error) {
	return r.check(role, resource, action)
}
