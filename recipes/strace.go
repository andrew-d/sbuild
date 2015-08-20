package recipes

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Sirupsen/logrus"

	"github.com/andrew-d/sbuild/builder"
	"github.com/andrew-d/sbuild/recipes/templates"
	"github.com/andrew-d/sbuild/types"
)

type StraceRecipe struct {
	*templates.BaseRecipe
}

func init() {
	builder.RegisterRecipe(&StraceRecipe{})
}

func (r *StraceRecipe) Info() *types.RecipeInfo {
	return &types.RecipeInfo{
		Name:         "strace",
		Version:      "4.10",
		Dependencies: []string{},
		Sources: []string{
			"http://downloads.sourceforge.net/project/strace/strace/4.10/strace-4.10.tar.xz",
		},
		Sums: []string{
			"e6180d866ef9e76586b96e2ece2bfeeb3aa23f5cc88153f76e9caedd65e40ee2",
		},
		Binary: true,
	}
}

func (r *StraceRecipe) Prepare(ctx *types.BuildContext) error {
	srcdir := r.UnpackedDir(ctx, r.Info())

	// Patch
	cmd := exec.Command("patch", "-p2")
	cmd.Stdin = bytes.NewBuffer(bytes.TrimLeft(stracePatch, "\r\n"))
	cmd.Dir = srcdir

	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not patch source")
		return err
	}

	return nil
}

func (r *StraceRecipe) Build(ctx *types.BuildContext) error {
	log.Info("Building strace")
	srcdir := r.UnpackedDir(ctx, r.Info())

	var cmd *exec.Cmd

	// 1. Configure
	log.Infof("Running ./configure")
	cmd = exec.Command(
		"./configure",
		"--host="+ctx.CrossPrefix,
		"--build=i686",
	)
	cmd.Dir = srcdir
	cmd.Env = ctx.Env.
		Set("CFLAGS", ctx.StaticFlags).
		AsSlice()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not run configure")
		return err
	}

	// 2. Fix linker in Makefile
	cmd = exec.Command(
		"sed",
		"-i",
		fmt.Sprintf("/^CC =/a LD = %s", ctx.Env.Get("LD")),
		filepath.Join(srcdir, "Makefile"),
	)
	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not patch Makefile")
		return err
	}

	// 3. Run build
	cmd = exec.Command("make")
	cmd.Dir = srcdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not run make")
		return err
	}

	log.Info("Finished building strace")
	return nil
}

func (r *StraceRecipe) Finalize(ctx *types.BuildContext, outDir string) error {
	srcdir := r.UnpackedDir(ctx, r.Info())

	source := filepath.Join(srcdir, "strace")
	target := filepath.Join(outDir, "strace")

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

	return nil
}

var stracePatch = []byte(`
diff --git 1/strace-4.10-orig/defs.h 2/strace-4.10/defs.h
index dad4fe8..d1378ab 100644
--- 1/strace-4.10-orig/defs.h
+++ 2/strace-4.10/defs.h
@@ -54,6 +54,8 @@
 #include <time.h>
 #include <sys/time.h>
 #include <sys/syscall.h>
+#include <asm-generic/ioctl.h>
+#include <linux/stat.h>
 
 #ifndef HAVE_STRERROR
 const char *strerror(int);
`)
