package commands

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// readInputs extracts postcodes from positional arguments, or reads from stdin if no arguments are provided.
func readInputs(cmd *cobra.Command, args []string) ([]string, error) {
	if len(args) > 0 {
		return args, nil
	}

	in := cmd.InOrStdin()
	if in == nil {
		return nil, errors.New("no postcode input provided (specify as arguments or pipe via stdin)")
	}

	// Avoid hanging when run interactively in a terminal without pipe or arguments
	if f, ok := in.(*os.File); ok {
		stat, err := f.Stat()
		if err == nil && (stat.Mode()&os.ModeCharDevice) != 0 {
			return nil, errors.New("no postcode input provided (specify as arguments or pipe via stdin)")
		}
	}

	var inputs []string
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			inputs = append(inputs, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading stdin: %w", err)
	}

	if len(inputs) == 0 {
		return nil, errors.New("no postcode input found in stdin")
	}

	return inputs, nil
}

// getAPIKey retrieves the API key from viper across supported alias keys.
func getAPIKey(v *viper.Viper) string {
	if k := v.GetString("api-key"); k != "" {
		return k
	}
	if k := v.GetString("apikey"); k != "" {
		return k
	}
	return ""
}

// getBaseURL retrieves the base URL from viper across supported alias keys.
func getBaseURL(v *viper.Viper) string {
	if u := v.GetString("base-url"); u != "" {
		return u
	}
	if u := v.GetString("api"); u != "" {
		return u
	}
	if u := v.GetString("url"); u != "" {
		return u
	}
	return ""
}

// getGoogleMapsKey retrieves the Google Maps API key across alias keys.
func getGoogleMapsKey(v *viper.Viper) string {
	if k := v.GetString("google-maps-api-key"); k != "" {
		return k
	}
	if k := v.GetString("google-maps-key"); k != "" {
		return k
	}
	if k := v.GetString("google_maps_api_key"); k != "" {
		return k
	}
	if k := v.GetString("google_maps_key"); k != "" {
		return k
	}
	return ""
}

// buildClient instantiates a postcode.Client wired to the active Viper configuration.
func buildClient(v *viper.Viper) (*postcode.Client, error) {
	var opts []postcode.ClientOption

	if apiKey := getAPIKey(v); apiKey != "" {
		opts = append(opts, postcode.WithAPIKey(apiKey))
	}

	if baseURL := getBaseURL(v); baseURL != "" {
		opts = append(opts, postcode.WithBaseURL(baseURL))
	}

	if timeout := v.GetDuration("timeout"); timeout > 0 {
		opts = append(opts, postcode.WithTimeout(timeout))
	} else {
		opts = append(opts, postcode.WithTimeout(10*time.Second))
	}

	if gKey := getGoogleMapsKey(v); gKey != "" {
		opts = append(opts, postcode.WithGoogleMapsKey(gKey))
	}

	return postcode.NewClient(opts...)
}

// isOffline checks whether offline mode is active.
// Priority order:
// 1. If command explicitly has a local --online flag set to true, online wins.
// 2. If command explicitly has a local --offline flag set to true, offline wins.
// 3. Otherwise, return the persistent global root --offline flag or POSTCODE_OFFLINE environment variable.
func isOffline(v *viper.Viper, cmd *cobra.Command) bool {
	if cmd != nil {
		if flag := cmd.Flags().Lookup("online"); flag != nil && flag.Changed {
			if online, err := cmd.Flags().GetBool("online"); err == nil && online {
				return false
			}
		}
		if flag := cmd.Flags().Lookup("offline"); flag != nil && flag.Changed {
			if offline, err := cmd.Flags().GetBool("offline"); err == nil && offline {
				return true
			}
		}
	}
	return v.GetBool("offline")
}
