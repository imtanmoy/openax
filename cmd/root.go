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
			&cli.BoolFlag{
				Name:    "prune-components",
				Aliases: []string{"prune"},
				Usage:   "Remove unused components from the filtered specification",
			},
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"n"},
				Usage:   "Preview filtering results without writing the output file",
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
		Paths:           cmd.StringSlice("paths"),
		Operations:      cmd.StringSlice("operations"),
		Tags:            cmd.StringSlice("tags"),
		PruneComponents: cmd.Bool("prune-components"),
	})
	if err != nil {
		return fmt.Errorf("failed to filter spec: %w", err)
	}

	// Handle dry run mode
	if cmd.Bool("dry-run") {
		return showDryRunSummary(filteredDoc, cmd)
	}

	return writeOutput(cmd, filteredDoc)
}

func showDryRunSummary(doc *openapi3.T, cmd *cli.Command) error {
	fmt.Println("🔍 Dry Run Mode - Filtering Results Summary")
	fmt.Println("==========================================")

	// Show basic info
	fmt.Printf("API Title: %s\n", doc.Info.Title)
	fmt.Printf("API Version: %s\n", doc.Info.Version)
	fmt.Printf("OpenAPI Version: %s\n", doc.OpenAPI)
	fmt.Println()

	// Show paths
	pathCount := len(doc.Paths.Map())
	fmt.Printf("📁 Paths included: %d\n", pathCount)
	if pathCount > 0 {
		for path := range doc.Paths.Map() {
			fmt.Printf("  • %s\n", path)
		}
	}
	fmt.Println()

	// Show components
	if doc.Components != nil {
		fmt.Println("🧩 Components included:")

		schemaCount := len(doc.Components.Schemas)
		fmt.Printf("  • Schemas: %d\n", schemaCount)
		if schemaCount > 0 && schemaCount <= 10 {
			for name := range doc.Components.Schemas {
				fmt.Printf("    - %s\n", name)
			}
		} else if schemaCount > 10 {
			count := 0
			for name := range doc.Components.Schemas {
				if count < 10 {
					fmt.Printf("    - %s\n", name)
					count++
				} else {
					fmt.Printf("    ... and %d more\n", schemaCount-10)
					break
				}
			}
		}

		paramCount := len(doc.Components.Parameters)
		if paramCount > 0 {
			fmt.Printf("  • Parameters: %d\n", paramCount)
		}

		responseCount := len(doc.Components.Responses)
		if responseCount > 0 {
			fmt.Printf("  • Responses: %d\n", responseCount)
		}

		requestBodyCount := len(doc.Components.RequestBodies)
		if requestBodyCount > 0 {
			fmt.Printf("  • Request Bodies: %d\n", requestBodyCount)
		}
	}
	fmt.Println()

	// Show applied filters
	fmt.Println("🎯 Applied Filters:")
	if paths := cmd.StringSlice("paths"); len(paths) > 0 {
		fmt.Printf("  • Paths: %v\n", paths)
	}
	if operations := cmd.StringSlice("operations"); len(operations) > 0 {
		fmt.Printf("  • Operations: %v\n", operations)
	}
	if tags := cmd.StringSlice("tags"); len(tags) > 0 {
		fmt.Printf("  • Tags: %v\n", tags)
	}
	if cmd.Bool("prune-components") {
		fmt.Println("  • Component pruning: enabled")
	}

	if len(cmd.StringSlice("paths")) == 0 && len(cmd.StringSlice("operations")) == 0 && len(cmd.StringSlice("tags")) == 0 {
		fmt.Println("  • No filters applied (showing entire specification)")
	}
	fmt.Println()

	// Show output information
	fmt.Println("📄 Output Configuration:")
	fmt.Printf("  • Format: %s\n", cmd.String("format"))
	if outputFile := cmd.String("output"); outputFile != "" {
		fmt.Printf("  • Would write to: %s\n", outputFile)
	} else {
		fmt.Println("  • Would write to: stdout")
	}

	fmt.Println()
	fmt.Println("✅ Dry run completed. Use without --dry-run to generate the filtered specification.")

	return nil
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
		err = os.WriteFile(outputFile, data, 0600)
	}

	return err
}
