package recipe

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(path string) (*Recipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load recipe %q: %w", path, err)
	}

	var r Recipe
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&r); err != nil {
		return nil, fmt.Errorf("load recipe %q: %w", path, err)
	}

	Normalize(&r)
	ApplyDefaults(&r)
	if err := Validate(&r); err != nil {
		return nil, err
	}

	return &r, nil
}
