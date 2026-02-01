package tinyrbac

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

			gotConf, err := readFromJson(filename)

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
