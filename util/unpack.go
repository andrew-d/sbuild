package util

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

var (
	ErrUnknownArchive = errors.New("unpack: unknown archive format")
)

func UnpackArchive(archive, intoDir string) error {
	if strings.HasSuffix(archive, ".tar") {
		return unpackTar(archive, intoDir)

	} else if strings.HasSuffix(archive, ".tar.gz") || strings.HasSuffix(archive, ".tgz") {
		return unpackTarGz(archive, intoDir)

	} else if strings.HasSuffix(archive, ".tar.bz2") {
		return unpackTarBz2(archive, intoDir)

	} else if strings.HasSuffix(archive, ".tar.lzma") {
		// TODO

	} else if strings.HasSuffix(archive, ".tar.xz") {
		// TODO

	} else if strings.HasSuffix(archive, ".zip") {
		return unpackZip(archive, intoDir)

	}

	return ErrUnknownArchive
}

func simpleUnpackCommand(f func(string, string) *exec.Cmd) func(string, string) error {
	fn := func(archive, intoDir string) error {
		cmd := f(archive, intoDir)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			// TODO: better logging
			//fmt.Fprintf(msg.Output, "Stdout:\n%s", stdout.String())
			//fmt.Fprintf(msg.Output, "Stderr:\n%s", stderr.String())
			return err
		}

		return nil
	}
	return fn
}

var (
	unpackTar = simpleUnpackCommand(func(archive, intoDir string) *exec.Cmd {
		return exec.Command("tar", "-C", intoDir, "-xf", archive)
	})

	unpackTarGz = simpleUnpackCommand(func(archive, intoDir string) *exec.Cmd {
		return exec.Command("tar", "-C", intoDir, "-xzf", archive)
	})

	unpackTarBz2 = simpleUnpackCommand(func(archive, intoDir string) *exec.Cmd {
		return exec.Command("tar", "-C", intoDir, "-xjf", archive)
	})

	unpackZip = simpleUnpackCommand(func(archive, intoDir string) *exec.Cmd {
		return exec.Command("unzip", archive, "-d", intoDir)
	})
)
