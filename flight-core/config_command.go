package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/adrg/xdg"
	"github.com/concertim/flight-user-suite/flight/cliui"
	"github.com/hashicorp/go-envparse"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
)

var (
	permittedKeys   = []string{"autostart"}
	permittedValues = []string{"on", "off"}
	// There must be an entry here for each entry in `permittedKeys`.
	permittedKeyDescriptions = []string{"If set to 'on', the Flight environment is automatically started when logging in."}
)

func configCommand() *cli.Command {
	settings := make([]string, 0)
	for i, key := range permittedKeys {
		var b strings.Builder
		for j, v := range permittedValues {
			b.WriteString("'")
			b.WriteString(v)
			b.WriteString("'")
			if j == len(permittedValues)-2 {
				b.WriteString(" or ")
			} else if j == len(permittedValues)-1 {
			} else {
				b.WriteString(", ")
			}
		}
		values := fmt.Sprintf("%s. ", b.String())

		setting := lipgloss.JoinHorizontal(
			lipgloss.Top,
			cliui.Bullet.Render(fmt.Sprintf("* %s:", key)),
			values,
			permittedKeyDescriptions[i],
		)
		settings = append(settings, setting)
	}
	settingsList := lipgloss.JoinVertical(lipgloss.Left, settings...)

	return &cli.Command{
		Name:        "config",
		Usage:       "Manage Flight User Suite settings",
		Description: wordwrap.String(fmt.Sprintf("Manage Flight User Suite settings. Available settings are shown below:\n\n%s", settingsList), maxTextWidth),
		Category:    "Configuration",
		Commands: []*cli.Command{
			{
				Name:        "list",
				Usage:       "List all Flight User Suite settings",
				Description: wordwrap.String("List all Flight User Suite settings.", maxTextWidth),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "global",
						Usage: "only show global settings",
					},
				},
				Action: configList,
			},
			{
				Name:  "set",
				Usage: "Set a Flight User Suite setting",
				Description: wordwrap.String(fmt.Sprintf(`Set a Flight User Suite setting. Available settings are shown below:

%s`, settingsList), maxTextWidth),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "global",
						Usage: "change the setting for all users",
					},
				},
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "key", UsageText: "<key>"},
					&cli.StringArg{Name: "value", UsageText: "<value>"},
				},
				Before: assertArgPresent("key", "value"),
				ShellComplete: func(ctx context.Context, cmd *cli.Command) {
					switch cmd.NArg() {
					case 0:
						for _, key := range permittedKeys {
							fmt.Println(key)
						}
					case 1:
						for _, value := range permittedValues {
							fmt.Println(value)
						}
					}
				},
				Action: configSet,
			},
			{
				Name:  "get",
				Usage: "Display a Flight User Suite setting",
				Description: wordwrap.String(`Display a Flight User Suite setting.

Currently supported keys are 'autostart'.`, maxTextWidth),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "global",
						Usage: "only get global settings",
					},
				},
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "key", UsageText: "<key>"},
				},
				Before: assertArgPresent("key"),
				ShellComplete: func(ctx context.Context, cmd *cli.Command) {
					switch cmd.NArg() {
					case 0:
						for _, key := range permittedKeys {
							fmt.Println(key)
						}
					}
				},
				Action: configGet,
			},
		},
	}
}

func configList(ctx context.Context, cmd *cli.Command) error {
	var config map[string]string
	if cmd.Bool("global") {
		config = loadGlobalConfig()
	} else {
		config = loadMergedConfigs()
	}
	for _, key := range permittedKeys {
		key = strings.ToUpper(fmt.Sprintf("FLIGHT_%s", key))
		if _, found := config[key]; !found {
			config[key] = "(unset)"
		}
	}
	for k, v := range config {
		k = strings.ToLower(strings.TrimPrefix(k, "FLIGHT_"))
		fmt.Printf("%s=%s\n", k, v)
	}
	return nil
}

func configSet(ctx context.Context, cmd *cli.Command) error {
	key := cmd.StringArg("key")
	value := cmd.StringArg("value")
	if !slices.Contains(permittedKeys, key) {
		return cli.Exit(fmt.Sprintf("Setting '%s' is not permitted", key), 1)
	}
	if !slices.Contains(permittedValues, value) {
		return cli.Exit(fmt.Sprintf("Value '%s' is not a valid value for '%s'", value, key), 1)
	}
	return saveConfig(key, value, cmd.Bool("global"))
}

func configGet(ctx context.Context, cmd *cli.Command) error {
	var config map[string]string
	if cmd.Bool("global") {
		config = loadGlobalConfig()
	} else {
		config = loadMergedConfigs()
	}
	key := cmd.StringArg("key")
	if !slices.Contains(permittedKeys, key) {
		return cli.Exit(fmt.Sprintf("Key '%s' is not known.\n", key), 1)
	}

	v, found := config[fmt.Sprintf("FLIGHT_%s", strings.ToUpper(key))]
	if found {
		fmt.Println(v)
	} else {
		fmt.Println("(unset)")
	}
	return nil
}

func loadMergedConfigs() map[string]string {
	var globalConfig map[string]string
	var userConfig map[string]string

	globalConfig = loadGlobalConfig()
	path, err := userConfigPath()
	if err != nil {
		log.Debug("Not merging user config", "path", path, "err", err)
	} else {
		userConfig, err = loadConfig(path)
		if err != nil {
			log.Debug("Not merging user config", "path", path, "err", err)
		}
	}

	config := globalConfig
	if userConfig != nil {
		maps.Copy(config, userConfig)
	}
	return config
}

func loadGlobalConfig() map[string]string {
	var globalConfig map[string]string

	path, err := globalConfigPath()
	if err != nil {
		log.Debug("Not merging global config", "path", path, "err", err)
		return make(map[string]string)
	}
	globalConfig, err = loadConfig(path)
	if err != nil {
		log.Debug("Not merging global config", "path", path, "err", err)
	}
	if globalConfig == nil {
		globalConfig = make(map[string]string)
	}
	return globalConfig
}

func loadConfig(path string) (map[string]string, error) {
	log.Debug("Loading config", "file", path)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	env, err := envparse.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return env, nil
}

func saveConfig(key, value string, global bool) error {
	var err error
	var path string
	var config map[string]string

	if global {
		path, err = globalConfigPath()
		if err != nil {
			return err
		}
	} else {
		path, err = userConfigPath()
		if err != nil {
			return err
		}
	}
	config, err = loadConfig(path)
	if err != nil {
		// This is fine.  We'll create the file below.
		config = make(map[string]string)
	}

	key = fmt.Sprintf("FLIGHT_%s", strings.ToUpper(key))
	config[key] = value
	log.Debug("Saving config", "file", path, "config", config, "global", global)
	var b strings.Builder
	for k, v := range config {
		fmt.Fprintf(&b, "%s=%s\n", k, v)
	}
	if global {
		return os.WriteFile(path, []byte(b.String()), 0o666)
	} else {
		return os.WriteFile(path, []byte(b.String()), 0o600)
	}
}

func globalConfigPath() (string, error) {
	name := filepath.Join("flight", "settings.config")
	paths := []string{"/etc/xdg"}
	return configPath(name, paths)
}

func userConfigPath() (string, error) {
	name := filepath.Join("flight", "settings.config")
	paths := []string{xdg.ConfigHome}
	return configPath(name, paths)
}

func configPath(name string, paths []string) (string, error) {
	searchedPaths := make([]string, 0, len(paths))
	for _, p := range paths {
		p = filepath.Join(p, name)

		dir := filepath.Dir(p)
		if exists(dir) {
			return p, nil
		}
		if err := os.MkdirAll(dir, os.ModeDir|0o777); err == nil {
			return p, nil
		}

		searchedPaths = append(searchedPaths, dir)
	}

	return "", fmt.Errorf("could not create any of the following paths: %v",
		searchedPaths)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || errors.Is(err, fs.ErrExist)
}
