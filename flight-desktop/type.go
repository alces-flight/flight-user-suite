package main

import (
	"fmt"
	"os"
	"path/filepath"

	"charm.land/log/v2"
	"gopkg.in/yaml.v3"
)

func loadAllTypes() ([]*Type, error) {
	glob := filepath.Join(env.FlightRoot, "usr", "lib", "desktop", "types", "*", "metadata.yml")
	log.Debug("Loading all desktop types", "glob", glob)
	types := make([]*Type, 0)
	matches, err := filepath.Glob(glob)
	if err != nil {
		return nil, fmt.Errorf("loading types: %w", err)
	}
	for _, match := range matches {
		id := filepath.Base(filepath.Dir(match))
		typ, err := loadType(id)
		if err != nil {
			log.Debug("Skipping bad type", "match", match, "err", err)
			continue
		}
		types = append(types, typ)
	}
	return types, nil
}

func loadType(id string) (*Type, error) {
	typ := &Type{ID: id}
	log.Debug("Loading desktop type", "dir", typ.dir())
	info, err := os.Stat(typ.dir())
	if err != nil {
		log.Debug("Error checking dir", "dir", typ.dir(), "err", err)
		return typ, UnknownType{Type: id}
	}
	if !info.IsDir() {
		log.Debug("Desktop type dir is not a directory", "dir", typ.dir())
		return typ, UnknownType{Type: id}
	}

	data, err := os.ReadFile(typ.metadataFile())
	if err != nil {
		log.Debug("Reading desktop type metadata", "metadataFile", typ.metadataFile(), "err", err)
		return typ, nil
	}
	err = yaml.Unmarshal(data, &typ)
	if err != nil {
		log.Debug("Loading desktop type metadata", "metadataFile", typ.metadataFile(), "err", err)
		return typ, nil
	}
	return typ, nil
}

type Type struct {
	ID           string `yaml:"id"`
	Summary      string `yaml:"summary"`
	URL          string `yaml:"url"`
	dependencies []dependency
}

func (t *Type) loadDependencies() error {
	if t.dependencies != nil {
		return nil
	}
	path := filepath.Join(t.dir(), "dependencies.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("loading dependencies for %s: %w", t.ID, err)
	}
	var deps []dependency
	err = yaml.Unmarshal(data, &deps)
	if err != nil {
		return fmt.Errorf("loading config from %s: %w", path, err)
	}
	t.dependencies = deps
	return nil
}

func (t *Type) dir() string {
	return filepath.Join(env.FlightRoot, "usr", "lib", "desktop", "types", t.ID)
}

func (t *Type) metadataFile() string {
	return filepath.Join(t.dir(), "metadata.yml")
}
