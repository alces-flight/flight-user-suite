package toolset

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Tool struct {
	Enabled  bool
	Name     string
	Synopsis string
}

func GetTools(flightRoot string, onlyEnabled bool) ([]*Tool, error) {
	toolDir := filepath.Join(flightRoot, "usr", "lib", "flight-core")
	toolSynopsisDir := filepath.Join(flightRoot, "usr", "share", "doc", "tools")

	entries, err := os.ReadDir(toolDir)
	if err != nil {
		return nil, fmt.Errorf("listing tools: %w", err)
	}

	tools := make([]*Tool, 0)
	for _, entry := range entries {
		if toolName, hasPrefix := strings.CutPrefix(entry.Name(), "flight-"); hasPrefix {
			info, err := entry.Info()
			if err != nil {
				return nil, fmt.Errorf("reading tool info: %w", err)
			}
			enabled := info.Mode()&0o111 != 0

			synopsisFile := filepath.Join(toolSynopsisDir, entry.Name())
			synopsis, _ := os.ReadFile(synopsisFile)

			tool := &Tool{
				Enabled:  enabled,
				Name:     toolName,
				Synopsis: strings.TrimSpace(string(synopsis)),
			}
			if !onlyEnabled || tool.Enabled {
				tools = append(tools, tool)
			}
		}
	}

	return tools, nil
}

func GetTool(flightRoot, toolName string) (*Tool, error) {
	tools, err := GetTools(flightRoot, false)
	if err != nil {
		return nil, err
	}
	for _, tool := range tools {
		if tool.Name == toolName {
			return tool, nil
		}
	}
	return nil, fmt.Errorf("unknown tool: %s", toolName)
}
