package config

const schemaBase = "https://raw.githubusercontent.com/mhamza15/forest/main/internal/config/schema"

// ConfigSchemaModeline returns the yaml-language-server modeline
// comment for the global config schema.
func ConfigSchemaModeline() string {
	return "# yaml-language-server: $schema=" + schemaBase + "/config.schema.json"
}

// ProjectSchemaModeline returns the yaml-language-server modeline
// comment for the project config schema.
func ProjectSchemaModeline() string {
	return "# yaml-language-server: $schema=" + schemaBase + "/project.schema.json"
}
