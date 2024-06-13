package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

func TestValueAtPath(t *testing.T) {
	assert := assert.New(t)

	testObj := map[string]interface{}{
		"some": map[string]interface{}{
			"nested": map[string]interface{}{
				"property": []string{"a", "b"},
			},
		},
	}

	value, ok := valueAtPath[[]string](testObj, []string{"some", "nested", "property"})
	assert.True(ok)
	assert.Equal([]string{"a", "b"}, value)
}

func TestReadConfig(t *testing.T) {
	assert := assert.New(t)
	configData := Config{
		ApiPrefix: "https://test-api.humanitec.io/",
		Org:       "unittest-org",
		Token:     "unittest-token",
	}

	configBytes, err := yaml.Marshal(configData)
	if err != nil {
		t.Fatal(err)
	}

	configFile, err := os.CreateTemp(".", ".humctl")
	if err != nil {
		t.Fatal(err)
	}
	defer configFile.Close()

	_, err = configFile.Write(configBytes)
	if err != nil {
		t.Fatal(err)
	}

	configPath := configFile.Name()
	defer os.Remove(configPath)

	config, diags := readConfig(HumanitecProviderModel{
		Config: types.StringValue(configPath),
	})
	assert.Len(diags, 0)
	assert.Equal(config, configData)
}

func TestReadConfigNonExistentFile(t *testing.T) {
	assert := assert.New(t)
	currentDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	configPath := currentDirectory + ".humctl"

	_, diags := readConfig(HumanitecProviderModel{
		Config: types.StringValue(configPath),
	})
	assert.Len(diags, 1)
	assert.Equal("Unable to read config file", diags[0].Summary())
}

func TestStrictUnmarshal(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "valid",
			input:     `{"field": "value"}`,
			expectErr: false,
		},
		{
			name:      "additional fields",
			input:     `{"field": "value", "a": "b"}`,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := struct {
				Field string `json:"field"`
			}{}

			err := strictUnmarshal([]byte(tc.input), &f)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOverrideMap(t *testing.T) {
	assert := assert.New(t)

	original := map[string]interface{}{
		"key":         "value",
		"another key": "another value",
	}
	overrides := map[string]interface{}{
		"key":     "new value 1",
		"new key": "new value 2",
	}

	overrideMap(original, overrides)
	assert.Equal(map[string]interface{}{
		"key":     "new value 1",
		"new key": "new value 2",
	}, original)
}
