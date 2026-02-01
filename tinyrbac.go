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

// NewFromJsonConfig creates an RBAC instance from a JSON config
// file at the given path. An error is returned when the config
// file cannot be proccessed.
func NewFromJsonConfig(path string) (*Rbac, error) {
	conf, err := readFromJson(path)
	if err != nil {
		return nil, fmt.Errorf("read config error: %w", err)
	}

	return buildFromConfig(conf)
}

// NewFromJsonConfig creates an RBAC instance from a YAML config
// file at the given path. An error is returned when the config
// file cannot be proccessed.
func NewFromYamlConfig(path string) (*Rbac, error) {
	conf, err := readFromYaml(path)
	if err != nil {
		return nil, fmt.Errorf("read config error: %w", err)
	}

	return buildFromConfig(conf)
}

// buildFromConfig builds the actual access map from config.
func buildFromConfig(conf *config) (*Rbac, error) {
	r := &Rbac{}
	buildRoleAndResourceNames(conf, r)

	for _, role := range conf.Roles {
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

// Check validates if 'role' has access to perform 'action' on 'resource'.
// TODO: Justify linearly searching instead of using a hash map.
func (r *Rbac) Check(role, resource, action string) bool {
	roleIdx, resourceIdx := -1, -1
	for idx, roleName := range r.roleIdxMap {
		if roleName == role {
			roleIdx = idx
			break
		}
	}
	if roleIdx == -1 {
		return false
	}

	for idx, resourceName := range r.resourceIdxMap {
		if resourceName == resource {
			resourceIdx = idx
			break
		}
	}
	if resourceIdx == -1 {
		return false
	}

	accessIdx := roleIdx*maxActions + getHTTPActionOffset(action)
	return r.accessMap[accessIdx]&resourceSet(1<<resourceIdx) != 0
}
