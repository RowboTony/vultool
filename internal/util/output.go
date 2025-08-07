package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"gopkg.in/yaml.v3"
)

// OutputResult outputs data in the specified format to the given writer.
// This is the centralized serialization function for all vultool CLI output.
//
// Supported formats:
//   - "json": Pretty-printed JSON using encoding/json
//   - "yaml": YAML output using gopkg.in/yaml.v3
//   - "toml": Not yet implemented (returns error with TODO message)
//
// Parameters:
//   - data: The data structure to serialize (must be JSON/YAML serializable)
//   - format: Output format string ("json", "yaml", or "toml")
//   - out: io.Writer to write the serialized output to
//
// Returns:
//   - error: nil on success, or error describing serialization/format issues
//
// Usage:
//   err := OutputResult(vaultInfo, "json", os.Stdout)
//   err := OutputResult(diffData, "yaml", fileWriter)
func OutputResult(data interface{}, format string, out io.Writer) error {
	switch format {
	case "json":
		return outputJSON(data, out)
	case "yaml":
		return outputYAML(data, out)
	case "toml":
		return errors.New("TOML not yet supported; TODO implement")
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func outputJSON(data interface{}, out io.Writer) error {
	serialized, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize to JSON: %w", err)
	}

	_, err = out.Write(serialized)
	if err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	return nil
}

func outputYAML(data interface{}, out io.Writer) error {
	serialized, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to serialize to YAML: %w", err)
	}

	_, err = out.Write(serialized)
	if err != nil {
		return fmt.Errorf("failed to write YAML: %w", err)
	}

	return nil
}
