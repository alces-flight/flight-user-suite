package main

import (
	"testing"

	"github.com/concertim/flight-user-suite/flight/configenv"
	"github.com/concertim/flight-user-suite/flight/doctor"
)

func TestLoadDefaultConfigIncludesWebSuiteDesktopDependencies(t *testing.T) {
	prevEnv := env
	t.Cleanup(func() {
		env = prevEnv
	})
	env = configenv.RepoLocalFlightEnv(t.TempDir())

	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}

	assertHasDependency(t, cfg.Dependencies, doctor.Dependency{
		Type:        doctor.TypeExecutable,
		Description: "Websockify",
		Optional:    true,
		Paths:       []string{"/usr/bin/websockify"},
	})
	assertHasDependency(t, cfg.Dependencies, doctor.Dependency{
		Type:        doctor.TypeExecutable,
		Description: "ImageMagick - screenshot support",
		Optional:    true,
		Paths:       []string{"/usr/bin/import"},
	})
}

func assertHasDependency(t *testing.T, deps []doctor.Dependency, want doctor.Dependency) {
	t.Helper()
	for _, dep := range deps {
		if dep.Type == want.Type &&
			dep.Description == want.Description &&
			dep.Optional == want.Optional &&
			len(dep.Paths) == len(want.Paths) &&
			dep.Paths[0] == want.Paths[0] {
			return
		}
	}
	t.Fatalf("expected dependency %#v in %#v", want, deps)
}
