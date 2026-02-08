package tinyrbac

import "math"

const (
	// Maximum sizes at compile time to overcome
	// slice header and pointer overhead.
	maxRoles      = 20
	maxActions    = 5 // HTTP
	maxResources  = 64
	unknownAction = -1

	allResources = "*"

	// allResourceAccess is strongly dependent on what the resourceSet type represents.
	allResourceAccess = math.MaxUint64
)
