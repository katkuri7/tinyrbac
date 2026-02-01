package tinyrbac

import (
	"slices"
)

func getHTTPActionOffset(action string) int {
	switch action {
	case "GET":
		return 0
	case "POST":
		return 1
	case "PUT":
		return 2
	case "PATCH":
		return 3
	case "DELETE":
		return 4
	default:
		return 0
	}
}

// buildRoleAndResourceNames extracts roles and resources from config.
// The extracted information is stored in a sorted manner which allows for
// the core idea of using the role-index and resource-index mapping to perform rbac operations.
func buildRoleAndResourceNames(c *config, r *Rbac) {
	resourcesMap := make(map[string]bool)
	i := 0

	// According to the config definition roles are always expected to be
	// unique and not repetitive.
	// TODO: Perhaps a validation of config is ideal before
	// further processing.
	for _, role := range c.Roles {
		r.roleIdxMap[i] = role.Name
		i++
		for _, resource := range role.Resources {
			// Special case where a role has access to all resources
			// for the provided actions.
			if resource.Name == allResources {
				continue
			}
			resourcesMap[resource.Name] = true
		}
	}

	// The [:] syntax returns a slice header that points to actual array data.
	// So a sort on this slice ultimately sorts our fixed size array.
	// We are concerned with only the first 'i' elements because performing a sort
	// on the entire array may result in the untouched elements (empty strings) accumulating in the beginning.
	slices.Sort(r.roleIdxMap[:i])

	i = 0
	for resource := range resourcesMap {
		r.resourceIdxMap[i] = resource
		i++
	}

	// Sorting because Go maps do not store/return data in an ordered fashion.
	slices.Sort(r.resourceIdxMap[:i])
}
