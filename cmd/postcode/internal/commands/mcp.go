package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewMCPCmd creates the mcp subcommand.
func NewMCPCmd(v *viper.Viper) *cobra.Command {
	var stdio bool

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start a Model Context Protocol (MCP) server for AI agents",
		Long: `Start an in-process Model Context Protocol (MCP) server over standard input/output (stdio).

This command enables seamless integration with Claude Desktop, Cursor, Antigravity IDE,
and any autonomous AI agent framework supporting the MCP standard.

The server exposes tools for postcode validation, diagnostic self-correction,
location resolution, reverse geocoding, and autocomplete, as well as resources
and prompts for Nigerian address normalization.`,
		Example: `  # Run in stdio mode (default for Claude Desktop, Cursor, and IDEs)
  postcode mcp

  # Example Claude Desktop configuration (~/Library/Application Support/Claude/claude_desktop_config.json):
  # {
  #   "mcpServers": {
  #     "postcode": {
  #       "command": "postcode",
  #       "args": ["mcp"],
  #       "env": {
  #         "POSTCODE_API_KEY": "YOUR_NIPOST_KEY"
  #       }
  #     }
  #   }
  # }`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _ := buildClient(v)

			logger := log.New(cmd.ErrOrStderr(), "[postcode-mcp] ", log.LstdFlags)
			server := postcode.NewMCPServer(client, postcode.WithMCPLogger(logger))

			fmt.Fprintf(cmd.ErrOrStderr(), "postcode MCP server running on stdio (PID: %d)...\n", os.Getpid())
			return server.Serve(cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&stdio, "stdio", true, "run server over stdio (standard input/output)")
	return cmd
}
