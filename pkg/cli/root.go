package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/janfietz/webunits-go-cli/pkg/config"
)

var (
	cfgFile string
	fields  []string
	pretty  bool
)

const (
	ExitOK      = 0
	ExitAuth    = 2
	ExitAPI     = 3
	ExitNetwork = 4
)

var rootCmd = &cobra.Command{
	Use:     "webuntis",
	Short:   "WebUntis CLI for AI Agents",
	Long:    `A Go CLI application optimized for AI agents to interact with the WebUntis API.`,
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(config.InitConfig)
	rootCmd.PersistentFlags().StringSliceVar(&fields, "fields", []string{}, "Comma-separated list of fields to include in JSON output")
	rootCmd.PersistentFlags().BoolVar(&pretty, "pretty", false, "Pretty print JSON output")
}

// Helper to print JSON with filtering
func printJSON(data interface{}) {
	// If no fields specified, just marshal
	if len(fields) == 0 {
		encoder := json.NewEncoder(os.Stdout)
		if pretty {
			encoder.SetIndent("", "  ")
		}
		if err := encoder.Encode(data); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Filtering logic
	// Marshal to JSON first to get map representation
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling data: %v\n", err)
		os.Exit(1)
	}

	// Determine if slice or object
	var asMap map[string]interface{}
	var asSlice []map[string]interface{}

	if err := json.Unmarshal(b, &asMap); err == nil {
		// It's an object
		filtered := filterMap(asMap, fields)
		outputJSON(filtered)
	} else if err := json.Unmarshal(b, &asSlice); err == nil {
		// It's a slice
		filteredSlice := make([]map[string]interface{}, len(asSlice))
		for i, item := range asSlice {
			filteredSlice[i] = filterMap(item, fields)
		}
		outputJSON(filteredSlice)
	} else {
		// Fallback for simple types or failures
		outputJSON(data)
	}
}

func filterMap(m map[string]interface{}, allowed []string) map[string]interface{} {
	filtered := make(map[string]interface{})
	for _, field := range allowed {
		field = strings.TrimSpace(field)
		if val, ok := m[field]; ok {
			filtered[field] = val
		}
	}
	return filtered
}

func outputJSON(data interface{}) {
	encoder := json.NewEncoder(os.Stdout)
	if pretty {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}
