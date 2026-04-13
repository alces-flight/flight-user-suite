package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"charm.land/log/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

type Howto struct {
	Path        string
	content     []byte
	frontMatter *FrontMatter
}

type FrontMatter struct {
	Admin bool   `yaml:"admin"`
	Name  string `yaml:"name"`
}

func (h *Howto) FullPath() string {
	return filepath.Join(howtoDir, h.Path)
}

func (h *Howto) Content() ([]byte, error) {
	if h.content != nil {
		return h.content, nil
	}
	err := h.Read()
	if err != nil {
		return nil, err
	}
	return h.content, nil
}

func (h *Howto) Read() error {
	if h.content != nil {
		// We've already read the howto.
		return nil
	}
	markdown, err := os.ReadFile(h.FullPath())
	if err != nil {
		return fmt.Errorf("reading howto: %w", err)
	}
	frontMatterBytes, markdown := SplitFrontmatter(markdown)
	if frontMatterBytes != nil {
		var frontMatter FrontMatter
		err = yaml.Unmarshal(frontMatterBytes, &frontMatter)
		if err != nil {
			log.Debug("Unable to parse frontmatter", "path", h.FullPath(), "err", err)
			h.frontMatter = nil
		} else {
			h.frontMatter = &frontMatter
		}
	}
	h.content = markdown
	return nil
}

func (h *Howto) IsAdminOnly() bool {
	err := h.Read()
	if err != nil {
		log.Debug("Unable to determine if howto is admin only", "err", err)
		return false
	}
	if h.frontMatter == nil {
		return false
	}
	return h.frontMatter.Admin
}

func (h *Howto) Name() string {
	err := h.Read()
	if err != nil {
		return h.nameFromPath()
	}
	if h.frontMatter != nil && h.frontMatter.Name != "" {
		return h.frontMatter.Name
	}
	return h.nameFromPath()
}

func (h *Howto) nameFromPath() string {
	leadingDigits := regexp.MustCompile(`^\d+-\s*`)
	otherDigits := regexp.MustCompile(`/\d+-\s*`)
	name := strings.TrimSuffix(h.Path, ".md")

	name = leadingDigits.ReplaceAllString(name, "")
	name = otherDigits.ReplaceAllString(name, "/")
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "/", " > ")
	return cases.
		Title(language.English, cases.Compact).
		String(name)
}

// Interface for sorting by howto path.
type ByPath []*Howto

func (a ByPath) Len() int           { return len(a) }
func (a ByPath) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPath) Less(i, j int) bool { return a[i].Path < a[j].Path }
