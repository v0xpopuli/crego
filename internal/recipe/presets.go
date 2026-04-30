package recipe

import "fmt"

func NewPreset(name string) (*Recipe, error) {
	var r Recipe
	switch name {
	case PresetWebBasic:
		r = presetBase("example-web", "example.com/example-web", ProjectTypeWeb)
	case PresetWebPostgres:
		r = presetBase("example-web", "example.com/example-web", ProjectTypeWeb)
		r.Database.Driver = DatabaseDriverPostgres
		r.Database.Migrations = DatabaseMigrationsMigrate
	case PresetWebMySQL:
		r = presetBase("example-web", "example.com/example-web", ProjectTypeWeb)
		r.Database.Driver = DatabaseDriverMySQL
		r.Database.Migrations = DatabaseMigrationsMigrate
	case PresetWebSQLite:
		r = presetBase("example-web", "example.com/example-web", ProjectTypeWeb)
		r.Database.Driver = DatabaseDriverSQLite
		r.Database.Migrations = DatabaseMigrationsMigrate
	case PresetCLIBasic:
		r = presetBase("example-cli", "example.com/example-cli", ProjectTypeCLI)
	default:
		return nil, fmt.Errorf("unknown recipe preset %q", name)
	}

	Normalize(&r)
	ApplyDefaults(&r)
	if err := Validate(&r); err != nil {
		return nil, err
	}

	return &r, nil
}

func presetBase(name string, module string, projectType string) Recipe {
	r := Recipe{
		Version: VersionV1,
		Project: ProjectConfig{
			Name:   name,
			Module: module,
			Type:   projectType,
		},
		Configuration: ConfigurationConfig{
			Format: ConfigurationFormatEnv,
		},
	}
	if projectType == ProjectTypeWeb {
		r.CI.GitHubActions = true
	}
	return r
}
