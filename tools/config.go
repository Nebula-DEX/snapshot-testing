package tools

import (
	"fmt"

	"github.com/tomwright/dasel"
	"github.com/tomwright/dasel/storage"
)

func UpdateConfig(filePath, configType string, newValues map[string]interface{}) error {
	root, err := dasel.NewFromFile(filePath, configType)
	if err != nil {
		return fmt.Errorf("failed to open %s config file with dasel: %w", filePath, err)
	}
	for k, v := range newValues {
		if err := root.Put(fmt.Sprintf(".%s", k), v); err != nil {
			return fmt.Errorf(
				"failed to update value for %s parameter in the %s file: %w",
				k,
				filePath,
				err,
			)
		}
	}

	if err := root.WriteToFile(filePath, "toml", []storage.ReadWriteOption{
		storage.IndentOption("  "),
		storage.PrettyPrintOption(true),
	}); err != nil {
		return fmt.Errorf("failed to write updated config to file %s: %w", filePath, err)
	}
	return nil
}
