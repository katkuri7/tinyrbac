package tinyrbac

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

			gotConf, err := readFromYaml(filename)

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
