package configenv

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-envparse"
)

const flightConfigPath = "/etc/xdg/flight.config"
const defaultFlightRoot = "/opt/flight"

type Env struct {
	FlightRoot      string
	FlightStateRoot string
	ClusterName     string
}

// InitFlightEnv resolves FLIGHT_ROOT and FLIGHT_STATE_ROOT from the process
// environment first, then from /etc/xdg/flight.config for any still-unset
// values, and finally from built-in defaults. Resolved values are written back
// to the process environment before being returned.
func InitFlightEnv() (Env, error) {
	env := Env{
		FlightRoot:      os.Getenv("FLIGHT_ROOT"),
		FlightStateRoot: os.Getenv("FLIGHT_STATE_ROOT"),
		ClusterName:     os.Getenv("FLIGHT_STARTER_CLUSTER_NAME"),
	}
	if env.FlightRoot != "" && env.FlightStateRoot != "" {
		return env, nil
	}

	configEnv, err := loadFlightConfig()
	if err != nil {
		return Env{}, err
	}

	if env.FlightRoot == "" {
		env.FlightRoot = configEnv["FLIGHT_ROOT"]
		if env.FlightRoot == "" {
			env.FlightRoot = defaultFlightRoot
		}
		if err := os.Setenv("FLIGHT_ROOT", env.FlightRoot); err != nil {
			return Env{}, fmt.Errorf("setting FLIGHT_ROOT: %w", err)
		}
	}
	if env.FlightStateRoot == "" {
		env.FlightStateRoot = configEnv["FLIGHT_STATE_ROOT"]
		if env.FlightStateRoot == "" {
			env.FlightStateRoot = filepath.Join(env.FlightRoot, "var", "lib")
		}
		if err := os.Setenv("FLIGHT_STATE_ROOT", env.FlightStateRoot); err != nil {
			return Env{}, fmt.Errorf("setting FLIGHT_STATE_ROOT: %w", err)
		}
	}

	starterConfig, err := loadFlightStarterConfig(env.FlightRoot)
	if err != nil {
		return Env{}, err
	}
	if env.ClusterName == "" {
		env.ClusterName = starterConfig["FLIGHT_STARTER_CLUSTER_NAME"]
		if env.ClusterName != "" {
			if err := os.Setenv("FLIGHT_STARTER_CLUSTER_NAME", env.ClusterName); err != nil {
				return Env{}, fmt.Errorf("setting FLIGHT_STARTER_CLUSTER_NAME: %w", err)
			}
		}
	}

	return env, nil
}

// loadFlightConfig parses /etc/xdg/flight.config as KEY=VALUE entries and
// expands variable references using both the current process environment and
// other values from the same file. A missing config file is treated as empty;
// errors are returned only for unreadable, unparseable, cyclic, or empty-after-
// expansion configured values.
func loadFlightConfig() (map[string]string, error) {
	file, err := os.Open(flightConfigPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("opening %s: %w", flightConfigPath, err)
	}
	defer file.Close() // nolint:errcheck

	parsed, err := envparse.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", flightConfigPath, err)
	}

	resolved := make(map[string]string, len(parsed))
	resolving := make(map[string]bool, len(parsed))
	var resolve func(string) (string, error)
	resolve = func(key string) (string, error) {
		if value, ok := resolved[key]; ok {
			return value, nil
		}
		raw, ok := parsed[key]
		if !ok {
			return os.Getenv(key), nil
		}
		if resolving[key] {
			return "", fmt.Errorf("resolving %s: circular reference", key)
		}
		resolving[key] = true
		defer delete(resolving, key)

		var expandErr error
		value := os.Expand(raw, func(name string) string {
			if value, ok := os.LookupEnv(name); ok {
				return value
			}
			value, err := resolve(name)
			if err != nil {
				expandErr = err
				return ""
			}
			return value
		})
		if expandErr != nil {
			return "", expandErr
		}
		resolved[key] = value
		return value, nil
	}

	for key := range parsed {
		value, err := resolve(key)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("resolving %s: expands to an empty value", key)
		}
	}
	return resolved, nil
}

func loadFlightStarterConfig(flightRoot string) (map[string]string, error) {
	path := filepath.Join(flightRoot, "etc", "flight-starter.config")
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer file.Close()
	parsed, err := envparse.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return parsed, nil
}

func RepoLocalFlightEnv(root string) Env {
	return Env{
		FlightRoot:      root,
		FlightStateRoot: filepath.Join(root, "var", "lib"),
	}
}

func (e Env) WithStateRoot(root string) Env {
	e.FlightStateRoot = root
	return e
}

func (e Env) Environ() []string {
	env := os.Environ()
	env = append(env, "FLIGHT_ROOT="+e.FlightRoot)
	env = append(env, "FLIGHT_STATE_ROOT="+e.FlightStateRoot)
	return dedupeEnv(env)
}

func dedupeEnv(input []string) []string {
	seen := make(map[string]int, len(input))
	out := make([]string, 0, len(input))
	for _, item := range input {
		key, _, ok := strings.Cut(item, "=")
		if !ok {
			out = append(out, item)
			continue
		}
		if idx, ok := seen[key]; ok {
			out[idx] = item
			continue
		}
		seen[key] = len(out)
		out = append(out, item)
	}
	return out
}
