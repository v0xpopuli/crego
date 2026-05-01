package recipe

import "strings"

func Normalize(r *Recipe) *Recipe {
	if r == nil {
		return nil
	}

	r.Version = normalizeEnum(r.Version)
	r.Project.Name = strings.TrimSpace(r.Project.Name)
	r.Project.Module = strings.TrimSpace(r.Project.Module)
	r.Project.Type = normalizeEnum(r.Project.Type)
	r.Go.Version = strings.TrimSpace(r.Go.Version)
	r.Layout.Style = normalizeEnum(r.Layout.Style)
	r.Server.Framework = normalizeEnum(r.Server.Framework)
	r.Configuration.Format = normalizeEnum(r.Configuration.Format)
	r.SQLDatabase = normalizeEnum(r.SQLDatabase)
	r.ORMFramework = normalizeEnum(r.ORMFramework)
	for index, driver := range r.NoSQLDatabase {
		r.NoSQLDatabase[index] = normalizeEnum(driver)
	}
	r.Migrations = normalizeEnum(r.Migrations)
	mergeTopLevelDatabaseConfig(r)
	r.Database.Driver = normalizeEnum(r.Database.Driver)
	for index, driver := range r.Database.Drivers {
		r.Database.Drivers[index] = normalizeEnum(driver)
	}
	r.Database.Framework = normalizeEnum(r.Database.Framework)
	if r.Database.Framework == "database/sql" {
		r.Database.Framework = DatabaseFrameworkDatabaseSQL
	}
	r.Database.Migrations = normalizeEnum(r.Database.Migrations)
	r.Logging.Framework = normalizeEnum(r.Logging.Framework)
	r.Logging.Format = normalizeEnum(r.Logging.Format)

	return r
}

func mergeTopLevelDatabaseConfig(r *Recipe) {
	if r.SQLDatabase != "" && len(r.Database.Drivers) == 0 {
		r.Database.Driver = r.SQLDatabase
	}
	if len(r.NoSQLDatabase) > 0 && len(r.Database.Drivers) == 0 {
		drivers := make([]string, 0, 1+len(r.NoSQLDatabase))
		if r.SQLDatabase != "" && r.SQLDatabase != DatabaseDriverNone {
			drivers = append(drivers, r.SQLDatabase)
		} else if r.Database.Driver != "" && r.Database.Driver != DatabaseDriverNone {
			drivers = append(drivers, r.Database.Driver)
		}
		drivers = append(drivers, r.NoSQLDatabase...)
		r.Database.Drivers = drivers
	}
	if r.ORMFramework != "" {
		r.Database.Framework = r.ORMFramework
	}
	if r.Migrations != "" {
		r.Database.Migrations = r.Migrations
	}
}

func normalizeEnum(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
