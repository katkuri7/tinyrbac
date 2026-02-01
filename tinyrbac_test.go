package tinyrbac

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const roles = 3
const rolesJson = `{
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

const rolesYaml = `roles:
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
			name:                    "error reading from json conf",
			jsonContent:             " invalid json ",
			expectedRoleIdxMap:      nil,
			expectedResourcesIdxMap: nil,
			expectedAccessMap:       nil,
			wantErr:                 true,
			expectedErr:             "read config error",
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
			name:                    "error reading from yaml conf",
			yamlContent:             `rol`,
			expectedRoleIdxMap:      nil,
			expectedResourcesIdxMap: nil,
			expectedAccessMap:       nil,
			wantErr:                 true,
			expectedErr:             "read config error",
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
		name     string
		role     string
		resource string
		action   string
		expected bool
	}{
		{
			name:     "role has action access for resource",
			role:     "Instance Manager",
			resource: "instances",
			action:   "POST",
			expected: true,
		},
		{
			name:     "role does not have action access for resource",
			role:     "Auditor",
			resource: "instances",
			action:   "POST",
			expected: false,
		},
		{
			name:     "role not found",
			role:     "Operator",
			resource: "instances",
			action:   "POST",
			expected: false,
		},
		{
			name:     "resource not found",
			role:     "Instance Manager",
			resource: "orders",
			action:   "POST",
			expected: false,
		},
	}

	r, err := NewFromJsonConfig(f.Name())
	if err != nil {
		t.Error(err)
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, r.Check(tt.role, tt.resource, tt.action))
		})
	}
}
