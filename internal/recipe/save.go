package recipe

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Save(path string, r *Recipe) error {
	r = Normalize(r)
	r = ApplyDefaults(r)
	if err := Validate(r); err != nil {
		return err
	}

	data, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("save recipe %q: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("save recipe %q: %w", path, err)
	}

	return nil
}
