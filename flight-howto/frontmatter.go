package main

// Based on code taken from https://github.com/charmbracelet/glow/blob/master/utils/utils.go
// Licensed under the MIT License.

import "regexp"

// SplitFrontmatter returns the front matter header of a markdown file and the remainder.
func SplitFrontmatter(content []byte) ([]byte, []byte) {
	if frontmatterBoundaries := detectFrontmatter(content); frontmatterBoundaries[0] == 0 {
		return content[frontmatterBoundaries[0]:frontmatterBoundaries[1]], content[frontmatterBoundaries[1]:]
	}
	return nil, content
}

var yamlPattern = regexp.MustCompile(`(?m)^---\r?\n(\s*\r?\n)?`)

func detectFrontmatter(c []byte) []int {
	if matches := yamlPattern.FindAllIndex(c, 2); len(matches) > 1 {
		return []int{matches[0][0], matches[1][1]}
	}
	return []int{-1, -1}
}
