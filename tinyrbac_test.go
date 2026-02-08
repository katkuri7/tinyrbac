package tinyrbac

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const roles = 3
const rolesJson = `{
  "resources": ["instances", "applications", "audit-logs"],
  "roles": [
    {
      "name": "Admin",
      "resources": [
        {
          "name": "*",
          "actions": ["GET", "POST", "PUT", "PATCH", "DELETE"]
        }
      ]
    },
    {
      "name": "Instance Manager",
      "resources": [
        {
          "name": "instances",
          "actions": ["GET", "POST", "PUT", "PATCH", "DELETE"]
        },
		{
          "name": "audit-logs",
          "actions": [""]
        }
      ]
    },
    {
      "name": "Auditor",
      "resources": [
        {
          "name": "applications",
          "actions": ["GET"]
        },
        {
          "name": "audit-logs",
          "actions": ["GET"]
        }
      ]
    }
  ]
}`

const rolesYaml = `
resources:
- "instances"
- "applications"
- "audit-logs"
roles:
  - name: Admin
    resources:
      - name: "*"
        actions:
          - GET
          - POST
          - PUT
          - PATCH
          - DELETE
  
  - name: Instance Manager
    resources:
      - name: instances
        actions:
          - GET
          - POST
          - PUT
          - PATCH
          - DELETE
  
  - name: Auditor
    resources:
      - name: applications
        actions:
          - GET
      - name: audit-logs
        actions:
          - GET
`

func Test_NewFromJsonConfig(t *testing.T) {
	tests := []struct {
		name                    string
		jsonContent             string
		expectedRoleIdxMap      []string
		expectedResourcesIdxMap []string
		expectedAccessMap       []resourceSet
		wantErr                 bool
		expectedErr             string
	}{
		{
			name:                    "create rbac from json config",
			jsonContent:             rolesJson,
			expectedRoleIdxMap:      []string{"Admin", "Auditor", "Instance Manager"},
			expectedResourcesIdxMap: []string{"applications", "audit-logs", "instances"},
			expectedAccessMap: []resourceSet{
				allResourceAccess, allResourceAccess,
				allResourceAccess, allResourceAccess,
				allResourceAccess,
				3, 0, 0, 0, 0, 4, 4, 4, 4, 4,
			},
			wantErr:     false,
			expectedErr: "",
		},
		{
			name:                    "invalid json config",
			jsonContent:             " invalid json ",
			expectedRoleIdxMap:      nil,
			expectedResourcesIdxMap: nil,
			expectedAccessMap:       nil,
			wantErr:                 true,
			expectedErr:             "read config",
		},
		{
			name:                    "validation error",
			jsonContent:             `{"resources": []}`,
			expectedRoleIdxMap:      nil,
			expectedResourcesIdxMap: nil,
			expectedAccessMap:       nil,
			wantErr:                 true,
			expectedErr:             "validate config: " + ErrNoResources.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := os.CreateTemp(".", "*.json")
			defer os.Remove(f.Name())
			f.Write([]byte(tt.jsonContent))

			r, err := NewFromJsonConfig(f.Name())

			if tt.wantErr == false {
				require.NoError(t, err)
				require.NotNil(t, r.roleIdxMap)
				require.NotNil(t, r.resourceIdxMap)
				require.NotNil(t, r.accessMap)
				assert.Equal(t, tt.expectedRoleIdxMap, r.roleIdxMap[:roles])
				assert.Equal(t, tt.expectedResourcesIdxMap, r.resourceIdxMap[:roles])
				assert.Equal(t, tt.expectedAccessMap, r.accessMap[:roles*maxActions])
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				require.Nil(t, r)
			}
		})
	}
}

func Test_NewFromYamlConfig(t *testing.T) {
	tests := []struct {
		name                    string
		yamlContent             string
		expectedRoleIdxMap      []string
		expectedResourcesIdxMap []string
		expectedAccessMap       []resourceSet
		wantErr                 bool
		expectedErr             string
	}{
		{
			name:                    "create rbac from yaml config",
			yamlContent:             rolesYaml,
			expectedRoleIdxMap:      []string{"Admin", "Auditor", "Instance Manager"},
			expectedResourcesIdxMap: []string{"applications", "audit-logs", "instances"},
			expectedAccessMap: []resourceSet{
				allResourceAccess, allResourceAccess,
				allResourceAccess, allResourceAccess,
				allResourceAccess,
				3, 0, 0, 0, 0, 4, 4, 4, 4, 4,
			},
			wantErr:     false,
			expectedErr: "",
		},
		{
			name:                    "invalid yaml config",
			yamlContent:             "rol",
			expectedRoleIdxMap:      nil,
			expectedResourcesIdxMap: nil,
			expectedAccessMap:       nil,
			wantErr:                 true,
			expectedErr:             "read config",
		},
		{
			name:                    "invalid config resources",
			yamlContent:             `resources:`,
			expectedRoleIdxMap:      nil,
			expectedResourcesIdxMap: nil,
			expectedAccessMap:       nil,
			wantErr:                 true,
			expectedErr:             "validate config: " + ErrNoResources.Error(),
		},
		{
			name: "invalid config roles",
			yamlContent: `
resources:
- "instances"
- "applications"
- "audit-logs"`,
			expectedRoleIdxMap:      nil,
			expectedResourcesIdxMap: nil,
			expectedAccessMap:       nil,
			wantErr:                 true,
			expectedErr:             "validate config: " + ErrNoRoles.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := os.CreateTemp(".", "*.yaml")
			defer os.Remove(f.Name())
			f.Write([]byte(tt.yamlContent))

			r, err := NewFromYamlConfig(f.Name())

			if tt.wantErr == false {
				require.NoError(t, err)
				require.NotNil(t, r.roleIdxMap)
				require.NotNil(t, r.resourceIdxMap)
				require.NotNil(t, r.accessMap)
				assert.Equal(t, tt.expectedRoleIdxMap, r.roleIdxMap[:roles])
				assert.Equal(t, tt.expectedResourcesIdxMap, r.resourceIdxMap[:roles])
				assert.Equal(t, tt.expectedAccessMap, r.accessMap[:roles*maxActions])
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				require.Nil(t, r)
			}
		})
	}
}

func Test_Check(t *testing.T) {
	f, _ := os.CreateTemp(".", "*.json")
	defer os.Remove(f.Name())
	f.Write([]byte(rolesJson))

	testcases := []struct {
		name           string
		role           string
		resource       string
		action         string
		expectedAccess bool
		expectedError  string
	}{
		{
			name:           "role has action access for resource",
			role:           "Instance Manager",
			resource:       "instances",
			action:         "POST",
			expectedAccess: true,
		},
		{
			name:           "role does not have action access for resource",
			role:           "Auditor",
			resource:       "instances",
			action:         "POST",
			expectedAccess: false,
		},
		{
			name:           "role not found",
			role:           "Operator",
			resource:       "instances",
			action:         "POST",
			expectedAccess: false,
			expectedError:  "unknown role: Operator",
		},
		{
			name:           "resource not found",
			role:           "Instance Manager",
			resource:       "orders",
			action:         "POST",
			expectedAccess: false,
			expectedError:  "unknown resource: orders",
		},
	}

	r, err := NewFromJsonConfig(f.Name())
	if err != nil {
		t.Error(err)
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			access, err := r.Check(tt.role, tt.resource, tt.action)
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			}
			assert.Equal(t, tt.expectedAccess, access)
		})
	}
}
