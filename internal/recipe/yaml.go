package recipe

import "gopkg.in/yaml.v3"

func yamlMappingContains(value *yaml.Node, key string) bool {
	if value == nil || value.Kind != yaml.MappingNode {
		return false
	}

	for i := 0; i+1 < len(value.Content); i += 2 {
		if value.Content[i].Value == key {
			return true
		}
	}
	return false
}

func yamlMappingKeys(value *yaml.Node) []string {
	if value == nil || value.Kind != yaml.MappingNode {
		return nil
	}

	keys := make([]string, 0, len(value.Content)/2)
	for i := 0; i+1 < len(value.Content); i += 2 {
		keys = append(keys, value.Content[i].Value)
	}
	return keys
}
