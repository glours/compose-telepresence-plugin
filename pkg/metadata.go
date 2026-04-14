package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Metadata(description string, up, down *cobra.Command) error {
	return metadata(description, up, down)
}

func metadata(description string, up *cobra.Command, down *cobra.Command) error {
	providerMetadata := ProviderMetadata{
		Description: description,
	}
	providerMetadata.Up = commandParameters(up)
	providerMetadata.Down = commandParameters(down)

	jsonMetadata, err := json.Marshal(providerMetadata)
	if err != nil {
		return err
	}
	fmt.Printf(string(jsonMetadata))
	return nil
}

func commandParameters(cmd *cobra.Command) CommandMetadata {
	cmdMetadata := CommandMetadata{}
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		_, isRequired := f.Annotations[cobra.BashCompOneRequiredFlag]
		cmdMetadata.Parameters = append(cmdMetadata.Parameters, ParameterMetadata{
			Name:        f.Name,
			Description: f.Usage,
			Required:    isRequired,
			Type:        f.Value.Type(),
			Default:     f.DefValue,
		})
	})
	return cmdMetadata
}

type ProviderMetadata struct {
	Description string          `json:"description"`
	Up          CommandMetadata `json:"up"`
	Down        CommandMetadata `json:"down"`
}

type CommandMetadata struct {
	Parameters []ParameterMetadata `json:"parameters"`
}

type ParameterMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
	Default     string `json:"default,omitempty"`
}
