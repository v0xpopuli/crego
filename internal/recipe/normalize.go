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
	r.Database.Driver = normalizeEnum(r.Database.Driver)
	r.Database.Framework = normalizeEnum(r.Database.Framework)
	r.Database.Migrations = normalizeEnum(r.Database.Migrations)
	r.Configuration.Format = normalizeEnum(r.Configuration.Format)
	r.Logging.Provider = normalizeEnum(r.Logging.Provider)
	r.Logging.Format = normalizeEnum(r.Logging.Format)

	return r
}

func normalizeEnum(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
