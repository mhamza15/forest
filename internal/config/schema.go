package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed schema/config.schema.json
var configSchema []byte

//go:embed schema/project.schema.json
var projectSchema []byte

// SchemaDir returns the directory where forest writes its JSON Schema
// files for editor autocomplete.
func SchemaDir() string {
	return filepath.Join(ConfigDir(), "schema")
}

// WriteSchemas writes the embedded JSON Schema files to the config
// directory. This enables yaml-language-server to provide autocomplete
// when config files include a schema modeline.
func WriteSchemas() error {
	dir := SchemaDir()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating schema directory: %w", err)
	}

	files := map[string][]byte{
		"config.schema.json":  configSchema,
		"project.schema.json": projectSchema,
	}

	for name, data := range files {
		p := filepath.Join(dir, name)

		if err := os.WriteFile(p, data, 0o644); err != nil {
			return fmt.Errorf("writing schema %s: %w", name, err)
		}
	}

	return nil
}

// ConfigSchemaModeline returns the yaml-language-server modeline
// comment for the global config schema.
func ConfigSchemaModeline() string {
	return "# yaml-language-server: $schema=" + filepath.Join(SchemaDir(), "config.schema.json")
}

// ProjectSchemaModeline returns the yaml-language-server modeline
// comment for the project config schema.
func ProjectSchemaModeline() string {
	return "# yaml-language-server: $schema=" + filepath.Join(SchemaDir(), "project.schema.json")
}
