package recipe

import (
	"strings"

	"gopkg.in/yaml.v3"
)

type NoSQLDrivers []string

func (drivers *NoSQLDrivers) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		return nil
	}
	if value.Kind == yaml.ScalarNode {
		driver := strings.ToLower(strings.TrimSpace(value.Value))
		if driver == "" || driver == DatabaseDriverNone {
			*drivers = nil
			return nil
		}
		*drivers = []string{driver}
		return nil
	}

	var values []string
	if err := value.Decode(&values); err != nil {
		return err
	}
	*drivers = values
	return nil
}

func DatabaseDrivers(database DatabaseConfig) []string {
	raw := database.Drivers
	if len(raw) == 0 && database.Driver != "" {
		raw = []string{database.Driver}
	}
	if len(raw) == 0 {
		return []string{DatabaseDriverNone}
	}

	seen := make(map[string]struct{}, len(raw))
	drivers := make([]string, 0, len(raw))
	for _, driver := range raw {
		if driver == "" {
			continue
		}
		if _, ok := seen[driver]; ok {
			continue
		}
		seen[driver] = struct{}{}
		drivers = append(drivers, driver)
	}
	if len(drivers) == 0 {
		return []string{DatabaseDriverNone}
	}
	return drivers
}

func noSQLDatabaseDrivers(drivers []string) []string {
	result := make([]string, 0, len(drivers))
	for _, driver := range drivers {
		switch driver {
		case DatabaseDriverRedis, DatabaseDriverMongoDB:
			result = append(result, driver)
		}
	}
	return result
}

func primarySQLDatabaseDriver(drivers []string) string {
	for _, driver := range drivers {
		switch driver {
		case DatabaseDriverPostgres, DatabaseDriverMySQL, DatabaseDriverSQLite:
			return driver
		}
	}
	return DatabaseDriverNone
}

func primaryDatabaseDriver(drivers []string) string {
	for _, driver := range drivers {
		switch driver {
		case DatabaseDriverPostgres, DatabaseDriverMySQL, DatabaseDriverSQLite:
			return driver
		}
	}
	return firstDatabaseDriver(drivers)
}

func firstDatabaseDriver(drivers []string) string {
	if len(drivers) == 0 || drivers[0] == "" {
		return DatabaseDriverNone
	}
	return drivers[0]
}
