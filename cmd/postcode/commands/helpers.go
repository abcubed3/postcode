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

// buildClient instantiates a postcode.Client wired to the active Viper configuration.
func buildClient(v *viper.Viper) (*postcode.Client, error) {
	var opts []postcode.ClientOption

	if apiKey := v.GetString("apikey"); apiKey != "" {
		opts = append(opts, postcode.WithAPIKey(apiKey))
	}

	baseURL := v.GetString("api")
	if baseURL == "" {
		baseURL = v.GetString("url")
	}
	if baseURL != "" {
		opts = append(opts, postcode.WithBaseURL(baseURL))
	}

	if timeout := v.GetDuration("timeout"); timeout > 0 {
		opts = append(opts, postcode.WithTimeout(timeout))
	} else {
		opts = append(opts, postcode.WithTimeout(10*time.Second))
	}

	return postcode.NewClient(opts...)
}
