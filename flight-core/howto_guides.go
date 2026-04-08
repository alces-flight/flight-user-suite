package main

import (
	"fmt"
	"os"
	"path/filepath"

	"charm.land/log/v2"
)

func howTosDir(name string, isTool bool) string {
	if isTool {
		return filepath.Join(flightRoot, "usr", "share", "doc", fmt.Sprintf("flight-%s", name))
	} else {
		return filepath.Join(flightRoot, "usr", "share", "doc", name)
	}
}

func createHowtoSymlinks(name string, isTool bool) error {
	tgtDir := filepath.Join(flightRoot, "usr", "share", "doc", "howtos-enabled")
	srcDir := howTosDir(name, isTool)
	matches, err := filepath.Glob(filepath.Join(srcDir, "*.md"))
	if err != nil {
		return fmt.Errorf("globbing howtos: %w", err)
	}
	for _, oldpath := range matches {
		newpath := filepath.Join(tgtDir, filepath.Base(oldpath))
		log.Debug("Creating howto symlink", "target", oldpath, "linkName", newpath)
		if err = os.Symlink(oldpath, newpath); err != nil {
			return fmt.Errorf("creating howto symlink: %w", err)
		}
	}
	return nil
}

func removeHowtoSymlinks(name string, isTool bool) error {
	symDir := filepath.Join(flightRoot, "usr", "share", "doc", "howtos-enabled")
	srcDir := howTosDir(name, isTool)

	entries, err := os.ReadDir(symDir)
	if err != nil {
		return err
	}
	var firstErr error
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			symTgt, err := os.Readlink(filepath.Join(symDir, entry.Name()))
			if err != nil {
				return err
			}
			if !filepath.IsAbs(symTgt) {
				symTgt = filepath.Join(symDir, symTgt)
			}
			symTgtDir := filepath.Dir(filepath.Clean(symTgt))
			if symTgtDir == srcDir {
				linkName := filepath.Join(symDir, entry.Name())
				log.Debug("Removing howto symlink", "target", symTgt, "linkName", linkName)
				err = os.Remove(filepath.Join(symDir, entry.Name()))
				if err != nil {
					log.Debug("Error removing symlink", "linkName", linkName, "err", err)
					if firstErr == nil {
						firstErr = err
					}
				}
			}
		}
	}
	return firstErr
}
