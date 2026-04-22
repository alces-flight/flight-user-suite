package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Tool struct {
	Name        string
	Description string
	URL         string
	IconPath    string
}

var toolDefs = map[string]Tool{
	// Keys should match the unprefixed name of FUS tools.
	"desktop": {
		Name:        "Flight Desktop",
		Description: "Access interactive desktop sessions",
		URL:         "/desktop",
		IconPath:    "/assets/images/desktop.png",
	},
	"howto": {
		Name:        "Flight Howto",
		Description: "Learn about the Flight User Suite and using your cluster",
		URL:         "/howto",
		IconPath:    "/assets/images/howto.png",
	},
}

// Virtually all of the rest of this file is duplicated from flight-core
// TODO fix whatever caused commit 00252fd to not work so we can properly share
// functionality!

type FlightTool struct {
	Enabled  bool
	Name     string
	Synopsis string
}

func getTools(onlyEnabled bool) ([]*FlightTool, error) {
	//log.Debug("getting tools", "dir", toolDir, "onlyEnabled", onlyEnabled)

	toolSynopsisDir := filepath.Join(flightRoot, "usr", "share", "doc", "tools")
	toolDir := filepath.Join(flightRoot, "usr", "lib", "flight-core")

	entries, err := os.ReadDir(toolDir)
	if err != nil {
		return nil, fmt.Errorf("listing tools: %w", err)
	}
	tools := make([]*FlightTool, 0)
	for _, entry := range entries {
		if toolName, hasPrefix := strings.CutPrefix(entry.Name(), "flight-"); hasPrefix {
			info, err := entry.Info()
			if err != nil {
				return nil, fmt.Errorf("reading tool info: %w", err)
			}
			enabled := info.Mode()&0111 != 0

			synopsisFile := filepath.Join(toolSynopsisDir, entry.Name())
			synopsis, _ := os.ReadFile(synopsisFile)

			tool := &FlightTool{Enabled: enabled, Name: toolName, Synopsis: strings.TrimSpace(string(synopsis))}

			if !onlyEnabled || tool.Enabled {
				tools = append(tools, tool)
			}
		}
	}
	return tools, nil
}
