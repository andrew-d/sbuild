package builder

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
)

type sourceCache struct {
	rootDir string
}

func newSourceCache(rootDir string) (*sourceCache, error) {
	ret := &sourceCache{
		rootDir: rootDir,
	}
	return ret, nil
}

// Fetch will attempt to download the given source, and verify that it matches
// the provided hash.  If fetching succeeds, it will symlink the downloaded
// source into the given directory.  If a source for a given package has
// already been downloaded, then it will not be downloaded a second time.  If a
// source fails hash verification, then any cached source will be removed (so
// it will be re-downloaded upon the next attempt).
func (c *sourceCache) Fetch(recipe, source, hash, intoDir string) error {
	filename, source := SplitSource(source)
	recipeCacheDir := filepath.Join(c.rootDir, recipe)
	filePath := filepath.Join(recipeCacheDir, filename)

	// Ensure the cache dir exists.
	if err := os.Mkdir(recipeCacheDir, 0700); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	// If the source already exists, then we don't need to download it.
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File does not exist.  Fetch it.
			log.WithFields(logrus.Fields{
				"recipe": recipe,
				"source": source,
			}).Info("Fetching source")
			if err := c.download(source, filePath); err != nil {
				log.WithFields(logrus.Fields{
					"recipe": recipe,
					"source": source,
					"err":    err,
				}).Error("Error fetching source")
				return err
			}
		} else {
			// An actual error - return.
			return err
		}
	} else {
		log.WithFields(logrus.Fields{
			"recipe": recipe,
			"source": source,
		}).Info("Source exists in cache")
	}

	// If we get here, then the source file should exist.  We hash the file and
	// compare it against the given hash.
	if err := c.compareHash(filePath, hash); err != nil {
		// TODO: make configurable
		os.Remove(filePath)
		return err
	}

	// Symlink the file from the cache directory into the source directory.
	if err := os.Symlink(filePath, filepath.Join(intoDir, filename)); err != nil {
		log.WithFields(logrus.Fields{
			"recipe":  recipe,
			"source":  source,
			"err":     err,
			"oldname": filePath,
			"newname": filepath.Join(intoDir, filename),
		}).Error("Could not symlink")
		return err
	}

	return nil
}

func (c *sourceCache) download(url, intoPath string) error {
	cmd := exec.Command(
		"curl",
		"-L",
		"-o", intoPath,
		url,
	)
	cmd.Env = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (c *sourceCache) compareHash(path, hash string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, f)
	if err != nil {
		return err
	}

	sum := hasher.Sum(nil)
	ssum := hex.EncodeToString(sum)

	if !strings.EqualFold(ssum, hash) {
		return fmt.Errorf(
			"hash of file %s (%s) does not match expected value",
			filepath.Base(path),
			ssum,
		)
	}

	return nil
}

// Splits the given source string into a filename to save to, and the source
// URL to be fetched.
func SplitSource(in string) (filename, source string) {
	if strings.Contains(in, "::") {
		parts := strings.SplitN(in, "::", 2)
		filename = parts[0]
		source = parts[1]
	} else {
		slashPos := strings.LastIndex(in, "/")
		filename = in[slashPos+1:]
		source = in
	}

	return
}
