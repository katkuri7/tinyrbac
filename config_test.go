package tinyrbac

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var unique2Char = []string{
	"aa", "ab", "ac", "ad", "ae", "af", "ag", "ah", "ai", "aj",
	"ak", "al", "am", "an", "ao", "ap", "aq", "ar", "as", "at",
	"au", "av", "aw", "ax", "ay", "az", "ba", "bb", "bc", "bd",
	"be", "bf", "bg", "bh", "bi", "bj", "bk", "bl", "bm", "bn",
	"bo", "bp", "bq", "br", "bs", "bt", "bu", "bv", "bw", "bx",
	"by", "bz", "ca", "cb", "cc", "cd", "ce", "cf", "cg", "ch",
	"ci", "cj", "ck", "cl", "cm", "cn",
}

func Test_readFromJson(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		jsonContent string
		wantConf    *config
		wantErr     string
	}{
		{
			name: "valid config",
			path: "testdata/valid.json",
			jsonContent: `{
                "Description": "RBAC config",
                "Roles": [
                    {
                        "Name": "admin",
                        "Description": "Full access",
                        "Resources": [
                            {"Name": "posts", "Actions": ["GET", "POST", "DELETE"]},
                            {"Name": "users", "Actions": ["GET", "POST"]}
                        ]
                    },
                    {
                        "Name": "user", 
                        "Description": "Read only",
                        "Resources": [
                            {"Name": "posts", "Actions": ["GET"]}
                        ]
                    }
                ]
            }`,
			wantConf: &config{
				Description: "RBAC config",
				Roles: []role{
					{
						Name:        "admin",
						Description: "Full access",
						Resources: []resource{
							{Name: "posts", Actions: []string{"GET", "POST", "DELETE"}},
							{Name: "users", Actions: []string{"GET", "POST"}},
						},
					},
					{
						Name:        "user",
						Description: "Read only",
						Resources: []resource{
							{Name: "posts", Actions: []string{"GET"}},
						},
					},
				},
			},
			wantErr: "",
		},
		{
			name:        "file not found",
			path:        "testdata/nonexistent.json",
			jsonContent: "",
			wantConf:    nil,
			wantErr:     "open json config",
		},
		{
			name:        "invalid json",
			path:        "testdata/invalid.json",
			jsonContent: `{ invalid json }`,
			wantConf:    nil,
			wantErr:     "unmarshal json config",
		},
		{
			name:        "missing description",
			path:        "testdata/missing-desc.json",
			jsonContent: `{"Roles": [{"Name": "admin", "Description": "test", "Resources": [{"Name": "posts", "Actions": ["GET"]}]}]}`,
			wantConf: &config{
				Description: "",
				Roles: []role{
					{
						Name:        "admin",
						Description: "test",
						Resources: []resource{
							{Name: "posts", Actions: []string{"GET"}},
						},
					},
				},
			},
			wantErr: "",
		},
		{
			name:        "empty file path",
			path:        "",
			jsonContent: "",
			wantConf:    nil,
			wantErr:     ErrConfigFileNotProvided.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filename string
			if tt.path != "" {
				f, err := os.CreateTemp(".", "*.json")
				defer os.Remove(f.Name())
				if err != nil {
					t.Error(err)
				}
				f.Write([]byte(tt.jsonContent))

				filename = f.Name()
				// A small work around to test the open config error
				if tt.jsonContent == "" {
					filename = "nocontent" + filename
				}
			}

			gotConf, err := newConfigFromJson(filename)

			if tt.wantErr == "" {
				require.NoError(t, err)
				require.NotNil(t, gotConf)
				assert.Equal(t, tt.wantConf, gotConf)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, gotConf)
			}
		})
	}
}

func Test_readFromYaml(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		yamlContent string
		wantConf    *config
		wantErr     string
	}{
		{
			name: "valid config",
			path: "testdata/valid.yaml",
			yamlContent: `description: "RBAC config"
roles:
  - name: admin
    description: "Full access"
    resources:
      - name: posts
        actions: [GET, POST, DELETE]
      - name: users
        actions: [GET, POST]
  - name: user
    description: "Read only" 
    resources:
      - name: posts
        actions: [GET]`,
			wantConf: &config{
				Description: "RBAC config",
				Roles: []role{
					{
						Name:        "admin",
						Description: "Full access",
						Resources: []resource{
							{Name: "posts", Actions: []string{"GET", "POST", "DELETE"}},
							{Name: "users", Actions: []string{"GET", "POST"}},
						},
					},
					{
						Name:        "user",
						Description: "Read only",
						Resources: []resource{
							{Name: "posts", Actions: []string{"GET"}},
						},
					},
				},
			},
			wantErr: "",
		},
		{
			name:        "file not found",
			path:        "testdata/nonexistent.yaml",
			yamlContent: "",
			wantConf:    nil,
			wantErr:     "open yaml config",
		},
		{
			name:        "invalid yaml",
			path:        "testdata/invalid.yaml",
			yamlContent: `rol`, // Invalid YAML syntax
			wantConf:    nil,
			wantErr:     "unmarshal yaml config",
		},
		{
			name: "missing description",
			path: "testdata/missing-desc.yaml",
			yamlContent: `roles:
  - name: admin
    description: "test"
    resources:
      - name: posts
        actions: [GET]`,
			wantConf: &config{
				Description: "",
				Roles: []role{
					{
						Name:        "admin",
						Description: "test",
						Resources: []resource{
							{Name: "posts", Actions: []string{"GET"}},
						},
					},
				},
			},
			wantErr: "",
		},
		{
			name:        "empty file path",
			path:        "",
			yamlContent: "",
			wantConf:    nil,
			wantErr:     ErrConfigFileNotProvided.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filename string
			if tt.path != "" {
				f, err := os.CreateTemp(".", "*.yaml")
				if err != nil {
					t.Error(err)
				}
				defer os.Remove(f.Name())

				if tt.yamlContent != "" {
					f.Write([]byte(tt.yamlContent))
				}
				f.Close()

				filename = f.Name()
				// Workaround for file not found test
				if tt.yamlContent == "" {
					filename = "nocontent" + filepath.Ext(f.Name())
				}
			}

			gotConf, err := newConfigFromYaml(filename)

			if tt.wantErr == "" {
				require.NoError(t, err)
				require.NotNil(t, gotConf)
				assert.Equal(t, tt.wantConf, gotConf)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, gotConf)
			}
		})
	}
}

func Test_validate(t *testing.T) {
	c := role{
		Name: "Auditor",
		Resources: []resource{
			{
				Name: "instances",
			},
		},
	}

	// To test config with roles greater than maximum.
	moreThanMaxRoles := make([]role, maxRoles+1)
	for i := range maxRoles + 1 {
		moreThanMaxRoles[i] = c
	}

	tests := []struct {
		name        string
		c           *config
		wantErr     bool
		expectedErr string
	}{
		{
			name: "succesful validation",
			c: &config{
				Resources: []string{"instances", "applications", "audit-logs", ""},
				Roles: []role{
					{
						Name: "Admin",
						Resources: []resource{
							{
								Name:    "*",
								Actions: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
							},
						},
					},
					{
						Name: "Instance Manager",
						Resources: []resource{
							{
								Name:    "instances",
								Actions: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
							},
						},
					},
				},
			},
		},
		{
			name: "no resources",
			c: &config{
				Resources: []string{},
			},
			wantErr:     true,
			expectedErr: ErrNoResources.Error(),
		},
		{
			name: "resources exceed maximum",
			c: &config{
				Resources: unique2Char,
			},
			wantErr:     true,
			expectedErr: fmt.Sprintf("resources exceeded: maximum %d but config has %d", maxResources, len(unique2Char)),
		},
		{
			name: "no roles",
			c: &config{
				Resources: []string{"instances", "applications", "audit-logs"},
				Roles:     []role{},
			},
			wantErr:     true,
			expectedErr: ErrNoRoles.Error(),
		},
		{
			name: "empty role name",
			c: &config{
				Resources: []string{"instances", "applications", "audit-logs"},
				Roles: []role{
					{
						Name: "",
					},
				},
			},
			wantErr:     true,
			expectedErr: "empty role: name not defined at index 0",
		},
		{
			name: "empty resources for role",
			c: &config{
				Resources: []string{"instances", "applications", "audit-logs"},
				Roles: []role{
					{
						Name: "Auditor",
					},
				},
			},
			wantErr:     true,
			expectedErr: "empty resources: not defined for role Auditor",
		},
		{
			name: "undefined resource",
			c: &config{
				Resources: []string{"instances", "applications", "audit-logs"},
				Roles: []role{
					{
						Name: "Auditor",
						Resources: []resource{
							{
								Name:    "storage",
								Actions: []string{"GET"},
							},
						},
					},
				},
			},
			wantErr:     true,
			expectedErr: "undefined resource: storage for role Auditor: storage not defined in resources",
		},
		{
			name: "roles exceeded",
			c: &config{
				Resources: []string{"instances"},
				Roles:     moreThanMaxRoles,
			},
			wantErr:     true,
			expectedErr: fmt.Sprintf("roles exceeded: maximum %d but config has %d", maxRoles, len(moreThanMaxRoles)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.c.validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
