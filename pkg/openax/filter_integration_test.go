package openax

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestApplyFilter_Integration(t *testing.T) {
	// Load the sample OpenAPI specification
	wd, err := os.Getwd()
	require.NoError(t, err)

	// Construct the absolute path to the test file
	filePath := filepath.Join(filepath.Dir(wd), "../testdata/specs/petstore.yaml")

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(filePath)
	require.NoError(t, err)

	t.Run("filter by tags", func(t *testing.T) {
		// Define filter options
		opts := FilterOptions{
			Tags: []string{"pet"},
		}

		// Apply the filter
		filteredDoc, err := applyFilter(doc, opts)
		require.NoError(t, err)

		// Load the expected output
		expectedFilePath := filepath.Join(filepath.Dir(wd), "../testdata/expected/petstore_filtered_by_tag.yaml")
		expectedData, err := os.ReadFile(expectedFilePath)
		require.NoError(t, err)

		// Marshal the filtered doc to YAML
		filteredData, err := yaml.Marshal(filteredDoc)
		require.NoError(t, err)

		// Compare the filtered output with the expected output
		require.YAMLEq(t, string(expectedData), string(filteredData))
	})

	t.Run("filter by paths", func(t *testing.T) {
		// Define filter options
		opts := FilterOptions{
			Paths: []string{"/pet"},
		}

		// Apply the filter
		filteredDoc, err := applyFilter(doc, opts)
		require.NoError(t, err)

		// Load the expected output
		expectedFilePath := filepath.Join(filepath.Dir(wd), "../testdata/expected/petstore_filtered_by_path.yaml")
		expectedData, err := os.ReadFile(expectedFilePath)
		require.NoError(t, err)

		// Marshal the filtered doc to YAML
		filteredData, err := yaml.Marshal(filteredDoc)
		require.NoError(t, err)

		// Compare the filtered output with the expected output
		require.YAMLEq(t, string(expectedData), string(filteredData))
	})

	t.Run("filter by operations", func(t *testing.T) {
		// Define filter options
		opts := FilterOptions{
			Operations: []string{"getPetById"},
		}

		// Apply the filter
		filteredDoc, err := applyFilter(doc, opts)
		require.NoError(t, err)

		// Load the expected output
		expectedFilePath := filepath.Join(filepath.Dir(wd), "../testdata/expected/petstore_filtered_by_operation.yaml")
		expectedData, err := os.ReadFile(expectedFilePath)
		require.NoError(t, err)

		// Marshal the filtered doc to YAML
		filteredData, err := yaml.Marshal(filteredDoc)
		require.NoError(t, err)

		// Compare the filtered output with the expected output
		require.YAMLEq(t, string(expectedData), string(filteredData))
	})
}
