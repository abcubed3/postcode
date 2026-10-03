package commands

import (
	"fmt"
	"io"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Version   = "v0.1.0"
	GitCommit = "dev"
	BuildDate = "unknown"
)

type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// NewVersionCmd creates the version subcommand.
func NewVersionCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display the postcode CLI version, commit, and build details",
		RunE: func(cmd *cobra.Command, args []string) error {
			info := VersionInfo{
				Version:   Version,
				GitCommit: GitCommit,
				BuildDate: BuildDate,
				GoVersion: runtime.Version(),
				OS:        runtime.GOOS,
				Arch:      runtime.GOARCH,
			}

			return PrintOutput(cmd, v, info, func(w io.Writer) error {
				_, _ = fmt.Fprintf(w, "postcode version %s (commit: %s, built: %s, %s/%s, %s)\n",
					info.Version, info.GitCommit, info.BuildDate, info.OS, info.Arch, info.GoVersion)
				return nil
			})
		},
	}

	return cmd
}
