package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Execute runs the root command using context.Background.
func Execute() error {
	return NewRootCmd().Execute()
}

// NewRootCmd creates the configured root cobra.Command hierarchy.
func NewRootCmd() *cobra.Command {
	v := viper.New()

	rootCmd := &cobra.Command{
		Use:           "postcode",
		Short:         "CLI for Nigeria's National Digital Alphanumeric Postcode system",
		Long: `postcode is a comprehensive command-line toolkit for parsing, validating,
formatting, geocoding, and querying Nigeria's 11-character digital postcodes.

Built for high-performance offline address processing and full integration with
the official NIPOST Postcode API (api.postcode.gov.ng).`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:        fmt.Sprintf("%s (commit: %s, built: %s)", Version, GitCommit, BuildDate),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig(v, cmd)
		},
	}

	// Persistent flags (inherited by all subcommands)
	rootCmd.PersistentFlags().String("config", "", "config file path (default is $HOME/.postcode.yaml or ./.postcode.yaml)")
	rootCmd.PersistentFlags().StringP("output", "o", "text", "output format: text, json, yaml, csv")
	rootCmd.PersistentFlags().StringP("apikey", "k", "", "NIPOST postcode API key (env: POSTCODE_API_KEY)")
	rootCmd.PersistentFlags().String("api", "", "NIPOST postcode api base url (env: POSTCODE_BASE_URL)")
	rootCmd.PersistentFlags().DurationP("timeout", "t", 10*time.Second, "HTTP request timeout")

	_ = rootCmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"text", "json", "yaml", "csv"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Command groupings for clear, organized --help output
	rootCmd.AddGroup(&cobra.Group{ID: "offline", Title: "Offline & Transformation Commands:"})
	rootCmd.AddGroup(&cobra.Group{ID: "gateway", Title: "Gateway & Geocoding Commands:"})
	rootCmd.AddGroup(&cobra.Group{ID: "ops", Title: "Operations & Developer Tools:"})

	// Subcommands
	validateCmd := NewValidateCmd(v)
	validateCmd.GroupID = "offline"

	parseCmd := NewParseCmd(v)
	parseCmd.GroupID = "offline"

	formatCmd := NewFormatCmd(v)
	formatCmd.GroupID = "offline"

	coordsCmd := NewCoordsCmd(v)
	coordsCmd.GroupID = "offline"

	mapCmd := NewMapCmd(v)
	mapCmd.GroupID = "offline"

	assembleCmd := NewAssembleCmd(v)
	assembleCmd.GroupID = "offline"

	disassembleCmd := NewDisassembleCmd(v)
	disassembleCmd.GroupID = "offline"

	lookupCmd := NewLookupCmd(v)
	lookupCmd.GroupID = "gateway"

	autocompleteCmd := NewAutocompleteCmd(v)
	autocompleteCmd.GroupID = "gateway"

	nearbyCmd := NewNearbyCmd(v)
	nearbyCmd.GroupID = "gateway"

	reverseCmd := NewReverseCmd(v)
	reverseCmd.GroupID = "gateway"

	statusCmd := NewStatusCmd(v)
	statusCmd.GroupID = "gateway"

	batchCmd := NewBatchCmd(v)
	batchCmd.GroupID = "ops"

	serveCmd := NewServeCmd(v)
	serveCmd.GroupID = "ops"

	mcpCmd := NewMCPCmd(v)
	mcpCmd.GroupID = "ops"

	versionCmd := NewVersionCmd(v)
	versionCmd.GroupID = "ops"

	diagnoseCmd := NewDiagnoseCmd(v)
	diagnoseCmd.GroupID = "offline"

	evalCmd := NewEvalCmd(v)
	evalCmd.GroupID = "ops"

	rootCmd.AddCommand(
		validateCmd,
		diagnoseCmd,
		parseCmd,
		formatCmd,
		coordsCmd,
		mapCmd,
		assembleCmd,
		disassembleCmd,
		lookupCmd,
		autocompleteCmd,
		nearbyCmd,
		reverseCmd,
		statusCmd,
		batchCmd,
		serveCmd,
		mcpCmd,
		evalCmd,
		versionCmd,
	)

	return rootCmd
}

func initConfig(v *viper.Viper, cmd *cobra.Command) error {
	v.SetEnvPrefix("POSTCODE")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()

	// Register known keys so AutomaticEnv is picked up during lookup
	v.SetDefault("apikey", "")
	v.SetDefault("url", "")
	v.SetDefault("output", "text")
	v.SetDefault("timeout", 10*time.Second)

	cfgFile, _ := cmd.Flags().GetString("config")
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			v.AddConfigPath(home)
		}
		v.AddConfigPath(".")
		v.SetConfigType("yaml")
		v.SetConfigName(".postcode")
	}

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) && !os.IsNotExist(err) {
			return fmt.Errorf("reading config: %w", err)
		}
	}

	// Bind persistent flags to viper
	if err := v.BindPFlags(cmd.Root().PersistentFlags()); err != nil {
		return err
	}
	// Bind local command flags to viper
	return v.BindPFlags(cmd.Flags())
}
