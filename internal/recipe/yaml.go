package recipe

import (
	"bytes"
	"strings"

	"gopkg.in/yaml.v3"
)

func MarshalYAML(r *Recipe) ([]byte, error) {
	var data bytes.Buffer
	encoder := yaml.NewEncoder(&data)
	encoder.SetIndent(2)
	if err := encoder.Encode(r); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}

	return spaceTopLevelYAML(data.String()), nil
}

func spaceTopLevelYAML(data string) []byte {
	data = strings.TrimRight(data, "\n")
	if data == "" {
		return []byte("\n")
	}

	lines := strings.Split(data, "\n")
	var builder strings.Builder
	for i, line := range lines {
		if i > 0 && isTopLevelYAMLKey(line) {
			builder.WriteString("\n")
		}
		builder.WriteString(line)
		builder.WriteString("\n")
	}

	return []byte(builder.String())
}

func isTopLevelYAMLKey(line string) bool {
	return line != "" &&
		line[0] != ' ' &&
		line[0] != '\t' &&
		line != "---" &&
		line != "..." &&
		strings.Contains(line, ":")
}

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
