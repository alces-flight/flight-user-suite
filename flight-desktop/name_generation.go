package main

import (
	cryptorand "crypto/rand"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"

	"charm.land/log/v2"
)

type nameGenerator struct {
	config       nameGeneratorConfig
	combinations [][]string
	adjectives   []string
	adverbs      []string
	animals      []string
	things       []string
	verbs        []string
	sessionType  string
	err          error
}

func newNameGenerator(sessionType string) *nameGenerator {
	ng := nameGenerator{sessionType: sessionType}
	ng.config = config.NameGenerator
	if ng.config.Strategy == "absurd" {
		ng.loadCombinations()
		ng.loadFileList("adjectives", &ng.adjectives)
		ng.loadFileList("adverbs", &ng.adverbs)
		ng.loadFileList("animals", &ng.animals)
		ng.loadFileList("things", &ng.things)
		ng.loadFileList("verbs", &ng.verbs)
	}
	return &ng
}

func (ng *nameGenerator) Generate() string {
	if ng.err != nil || ng.config.Strategy == "meaningless" {
		return fmt.Sprintf("%s.%s", ng.sessionType, cryptorand.Text()[0:8])
	}
	b := make([]string, 0, 3)
	combination := sample(ng.combinations)
	for _, c := range combination {
		var part string
		switch c {
		case "adjective":
			part = sample(ng.adjectives)
		case "adverb":
			part = sample(ng.adverbs)
		case "animal":
			part = sample(ng.animals)
		case "thing":
			part = sample(ng.things)
		case "verb":
			part = sample(ng.verbs)
		case "sessiontype":
			part = ng.sessionType
		}
		if part != "" {
			b = append(b, part)
		}
	}
	return strings.Join(b, "-")
}

func sample[S ~[]E, E any](options S) E {
	idx := rand.IntN(len(options))
	return options[idx]
}

func (ng *nameGenerator) loadFileList(name string, dst *[]string) {
	if ng.err != nil {
		return
	}
	path := filepath.Join(env.FlightRoot, "usr", "lib", "desktop", "name-generation", fmt.Sprintf("%s.txt", name))
	data, err := os.ReadFile(path)
	if err != nil {
		log.Debug("Unable to load name generation list", "path", path, "err", err)
		ng.err = err
		return
	}
	lines := make([]string, 0)
	for line := range strings.Lines(string(data)) {
		lines = append(lines, strings.TrimSpace(line))
	}
	*dst = lines
}

func (ng *nameGenerator) loadCombinations() {
	if ng.err != nil {
		return
	}
	path := filepath.Join(env.FlightRoot, "usr", "lib", "desktop", "name-generation", "combinations.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		log.Debug("Unable to load name generation combinations", "path", path, "err", err)
		ng.err = err
		return
	}
	combinations := make([][]string, 0)
	for line := range strings.Lines(string(data)) {
		combination := strings.Split(strings.TrimSpace(line), " ")
		combinations = append(combinations, combination)
	}
	ng.combinations = combinations
}
