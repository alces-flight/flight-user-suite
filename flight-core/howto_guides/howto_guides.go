package howto_guides

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"charm.land/log/v2"
)

var (
	flightRoot string = "/opt/flight"
)

func init() {
	if root, ok := os.LookupEnv("FLIGHT_ROOT"); ok {
		flightRoot = root
	}
}

func howTosDir(name string, isTool bool) string {
	if isTool {
		return filepath.Join(flightRoot, "usr", "share", "doc", fmt.Sprintf("flight-%s", name))
	} else {
		return filepath.Join(flightRoot, "usr", "share", "doc", name)
	}
}

func CreateHowtoSymlinks(name string, isTool bool) error {
	tgtDir := filepath.Join(flightRoot, "usr", "share", "doc", "howtos-enabled")
	srcDir := howTosDir(name, isTool)
	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == srcDir {
			return nil
		}
		if d.IsDir() {
			srcRel, _ := filepath.Rel(srcDir, path)
			dir := filepath.Join(tgtDir, srcRel)
			log.Debug("Creating howto directory", "dir", dir)
			err := os.MkdirAll(dir, 0o755)
			if err != nil {
				log.Debug("Failed to create directory", "dir", dir, "err", err)
				return filepath.SkipDir
			}
		}
		matched, _ := filepath.Match("*.md", d.Name())
		if matched {
			srcRel, _ := filepath.Rel(srcDir, path)
			linkName := filepath.Join(tgtDir, srcRel)
			log.Debug("Creating howto symlink", "target", path, "linkName", linkName)
			if err = os.Symlink(path, linkName); err != nil {
				return fmt.Errorf("creating howto symlink: %w", err)
			}
		}
		return nil
	})
	return err
}

func RemoveHowtoSymlinks(name string, isTool bool) error {
	symDir := filepath.Join(flightRoot, "usr", "share", "doc", "howtos-enabled")
	srcDir := filepath.Clean(howTosDir(name, isTool))

	err := filepath.WalkDir(symDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == srcDir {
			return nil
		}
		if d.IsDir() {
			// Nothing to do. Leaving empty directories around is OK and it
			// might not even be empty.
			return nil
		}
		matched, _ := filepath.Match("*.md", d.Name())
		if matched {
			symTgt, err := os.Readlink(path)
			if err != nil {
				return err
			}
			symTgt = filepath.Clean(symTgt)
			if strings.HasPrefix(symTgt, srcDir) {
				log.Debug("Removing howto symlink", "target", symTgt, "linkName", path)
				err = os.Remove(path)
				if err != nil {
					log.Debug("Error removing symlink", "linkName", path, "err", err)
				}
				return err
			}
		}
		return nil
	})
	return err
}
