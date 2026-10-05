package commands

import (
	"fmt"
	"io"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Version   = "dev"
	GitCommit = "dev"
	BuildDate = "unknown"
)

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		if (Version == "" || Version == "dev" || Version == "v0.1.0") && info.Main.Version != "" && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				if GitCommit == "" || GitCommit == "dev" {
					GitCommit = s.Value
				}
			case "vcs.time":
				if BuildDate == "" || BuildDate == "unknown" {
					BuildDate = s.Value
				}
			}
		}
	}
}

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
