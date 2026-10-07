package commands

import (
	"fmt"
	"os"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewCacheCmd creates the cache subcommand group.
func NewCacheCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage the persistent offline geocoding cache and datasets",
		Long: `Cache inspects, imports, exports, and flushes locally cached high-precision
building-level postcode coordinates stored at ~/.postcode/cache.json.

Any postcodes resolved online are automatically persisted to this local cache,
enabling instant offline resolution on subsequent queries.`,
		Example: `  postcode cache status
  postcode cache list
  postcode cache import custom_buildings.csv
  postcode cache import national_addresses.json
  postcode cache clear`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCacheStatus(cmd)
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Display local cache statistics and storage location",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCacheStatus(cmd)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all cached building postcodes and coordinates",
		RunE: func(cmd *cobra.Command, args []string) error {
			count, path := postcode.GetCacheStats()
			if count == 0 {
				fmt.Printf("Cache is empty (%s)\n", path)
				return nil
			}
			fmt.Printf("Cached Building Records (%d entries in %s):\n", count, path)
			fmt.Println("--------------------------------------------------------------------------------")
			return postcode.ExportBuildingsJSON(cmd.OutOrStdout())
		},
	}

	importCmd := &cobra.Command{
		Use:   "import [file.json|file.csv]",
		Short: "Import custom building datasets (JSON or CSV) into local offline cache",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			f, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("opening data file: %w", err)
			}
			defer f.Close()

			var imported int
			if len(filePath) > 4 && filePath[len(filePath)-4:] == ".csv" {
				imported, err = postcode.LoadBuildingsCSV(f)
			} else {
				imported, err = postcode.LoadBuildingsJSON(f)
			}
			if err != nil {
				return fmt.Errorf("importing data from %s: %w", filePath, err)
			}

			fmt.Printf("✓ Successfully imported %d building record(s) into local cache (%s)\n", imported, postcode.GetCachePath())
			return nil
		},
	}

	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear all local cached building records",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := postcode.ClearDiskCache(); err != nil {
				return fmt.Errorf("clearing cache: %w", err)
			}
			fmt.Println("✓ Local postcode cache cleared.")
			return nil
		},
	}

	cmd.AddCommand(statusCmd, listCmd, importCmd, clearCmd)
	return cmd
}

func runCacheStatus(cmd *cobra.Command) error {
	count, path := postcode.GetCacheStats()
	fmt.Fprintf(cmd.OutOrStdout(), "Local Cache Path:  %s\n", path)
	fmt.Fprintf(cmd.OutOrStdout(), "Cached Buildings:  %d entries\n", count)
	fmt.Fprintf(cmd.OutOrStdout(), "Offline Status:    Active (hierarchical fallback: building -> LGA -> state)\n")
	return nil
}
