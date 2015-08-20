package util

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/andrew-d/sbuild/util/assets"
)

// Finds all files named 'config.sub' in the given directory, and, if they
// don't support the given triple, replaces them with a version that does.
func ReplaceConfigSub(dir, triple string) error {
	adir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	err = filepath.Walk(adir, func(path string, info os.FileInfo, err error) error {
		if filepath.Base(path) == "config.sub" {
			cmd := exec.Command(path, "arm-linux-musleabihf")
			cmd.Dir = filepath.Dir(path)
			err := cmd.Run()

			// No errors, and no problems - we're good.
			if err == nil {
				return nil
			}

			// If this is not an exit status error, we return this as a real
			// error.
			if _, ok := err.(*exec.ExitError); !ok {
				return err
			}

			// Replace the file.
			return ioutil.WriteFile(path, assets.MustAsset("config.sub"), 0755)
		}

		return nil
	})

	return err
}
