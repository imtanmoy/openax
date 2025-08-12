package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"

	"github.com/imtanmoy/openax/pkg/openax"
)

func NewApp() *cli.Command {
	return &cli.Command{
		Name:  "openax",
		Usage: "Filter and validate OpenAPI specifications",
		Description: `OpenAx is a CLI tool that loads an OpenAPI spec, validates it, 
filters it down to specified paths/operations/tags, pulls in only 
the referenced components, and writes the result to JSON or YAML.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "input",
				Aliases:  []string{"i"},
				Usage:    "Input OpenAPI spec file (required)",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file (stdout if not specified)",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Value:   "yaml",
				Usage:   "Output format: json or yaml",
			},
			&cli.StringSliceFlag{
				Name:    "paths",
				Aliases: []string{"p"},
				Usage:   "Filter by paths (e.g., /users, /orders)",
			},
			&cli.StringSliceFlag{
				Name:  "operations",
				Usage: "Filter by operations (e.g., get, post, put, delete)",
			},
			&cli.StringSliceFlag{
				Name:    "tags",
				Aliases: []string{"t"},
				Usage:   "Filter by tags",
			},
			&cli.BoolFlag{
				Name:  "validate-only",
				Usage: "Only validate the spec without filtering",
			},
		},
		Action: runFilter,
	}
}

func runFilter(ctx context.Context, cmd *cli.Command) error {
	inputFile := cmd.String("input")

	client := openax.NewWithOptions(openax.LoadOptions{
		AllowExternalRefs: true,
		Context:           ctx,
	})

	if cmd.Bool("validate-only") {
		if err := client.ValidateOnly(inputFile); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
		fmt.Println("OpenAPI spec is valid")
		return nil
	}

	filteredDoc, err := client.LoadAndFilter(inputFile, openax.FilterOptions{
		Paths:      cmd.StringSlice("paths"),
		Operations: cmd.StringSlice("operations"),
		Tags:       cmd.StringSlice("tags"),
	})
	if err != nil {
		return fmt.Errorf("failed to filter spec: %w", err)
	}

	return writeOutput(cmd, filteredDoc)
}

func writeOutput(cmd *cli.Command, doc *openapi3.T) error {
	var data []byte
	var err error

	format := cmd.String("format")
	switch strings.ToLower(format) {
	case "json":
		data, err = json.MarshalIndent(doc, "", "  ")
	case "yaml", "yml":
		data, err = yaml.Marshal(doc)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}

	if err != nil {
		return err
	}

	outputFile := cmd.String("output")
	if outputFile == "" {
		fmt.Print(string(data))
	} else {
		err = os.WriteFile(outputFile, data, 0644)
	}

	return err
}
