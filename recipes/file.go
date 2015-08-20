package recipes

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Sirupsen/logrus"

	"github.com/andrew-d/sbuild/builder"
	"github.com/andrew-d/sbuild/recipes/templates"
	"github.com/andrew-d/sbuild/types"
)

type FileRecipe struct {
	*templates.BaseRecipe
}

func init() {
	builder.RegisterRecipe(&FileRecipe{})
}

func (r *FileRecipe) Info() *types.RecipeInfo {
	return &types.RecipeInfo{
		Name:         "file",
		Version:      "5.24",
		Dependencies: []string{},
		Sources: []string{
			"file-5.24.tar.gz::https://github.com/file/file/archive/FILE5_24.tar.gz",
		},
		Sums: []string{
			"52e160662c45d8b204c583552d80e4ab389a3a641f9745a458da2f6761c9b206",
		},
	}
}

func (r *FileRecipe) UnpackedDir(ctx *types.BuildContext) string {
	return filepath.Join(
		ctx.SourceDir,
		"file-FILE5_24",
	)
}

func (r *FileRecipe) Prepare(ctx *types.BuildContext) error {
	srcdir := r.UnpackedDir(ctx)

	// 1. Don't run tests (which we can't do while cross-compiling)
	if err := ioutil.WriteFile(
		filepath.Join(srcdir, "tests", "Makefile.in"),
		[]byte("all:\n\ttrue\n\ninstall:\n\ttrue\n\n"),
		0666,
	); err != nil {
		return err
	}

	// 2. Fix headers.
	cmd := exec.Command(
		"sed",
		"-i",
		`s/memory.h/string.h/`,
		filepath.Join(srcdir, "src", "encoding.c"),
		filepath.Join(srcdir, "src", "ascmagic.c"),
	)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (r *FileRecipe) Build(ctx *types.BuildContext) error {
	log.Info("Building file")
	srcdir := r.UnpackedDir(ctx)

	var cmd *exec.Cmd

	// 1. Run autotools
	cmd = exec.Command("autoreconf", "-i")
	cmd.Dir = srcdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not run autotools")
		return err
	}

	// 2. Configure in native mode
	log.Infof("Running native ./configure")
	cmd = exec.Command(
		"./configure",
		"--disable-shared",
	)
	cmd.Dir = srcdir
	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not run native configure")
		return err
	}

	// 3. Build native binary.
	log.Infof("Running native build")
	cmd = exec.Command("make")
	cmd.Dir = srcdir
	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not run native build")
		return err
	}

	// 4. Copy the native binary.
	nativePath := filepath.Join(ctx.SourceDir, "file")
	if err := r.CopyFile(
		filepath.Join(srcdir, "src", "file"),
		nativePath,
		0755,
	); err != nil {
		log.WithField("err", err).Error("Could not copy native binary")
		return err
	}

	// 5. Clean up.
	_ = exec.Command("make", "distclean").Run()

	// 6. Configure for cross-compiling.
	log.Infof("Running ./configure")
	cmd = exec.Command(
		"./configure",
		"--disable-shared",
		"--host="+ctx.CrossPrefix,
		"--build=i686",
	)
	cmd.Dir = srcdir
	cmd.Env = ctx.Env.
		Append("CC", ctx.StaticFlags).
		Set("CFLAGS", ctx.StaticFlags).
		AsSlice()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not run configure")
		return err
	}

	// 7. Patch the Makefile to use our native binary.
	cmd = exec.Command(
		"sed",
		"-i",
		fmt.Sprintf("s|FILE_COMPILE = file${EXEEXT}|FILE_COMPILE = %s|", nativePath),
		filepath.Join(srcdir, "magic", "Makefile"),
	)
	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not patch Makefile")
		return err
	}

	// 8. Run native build
	cmd = exec.Command("make")
	cmd.Dir = srcdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not run make")
		return err
	}

	log.Info("Finished building file")
	return nil
}

func (r *FileRecipe) Finalize(ctx *types.BuildContext, outDir string) error {
	srcdir := r.UnpackedDir(ctx)

	source := filepath.Join(srcdir, "src", "file")
	target := filepath.Join(outDir, "file")

	log.WithFields(logrus.Fields{
		"source": source,
		"target": target,
	}).Info("Copying binary")

	if err := r.CopyFile(source, target, 0755); err != nil {
		return err
	}

	if err := r.Strip(ctx, target); err != nil {
		return err
	}

	// Copy magic file too.
	if err := r.CopyFile(
		filepath.Join(srcdir, "magic", "magic.mgc"),
		filepath.Join(outDir, "magic.mgc"),
		0644,
	); err != nil {
		return err
	}

	return nil
}
