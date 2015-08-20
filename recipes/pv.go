package recipes

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Sirupsen/logrus"

	"github.com/andrew-d/sbuild/builder"
	"github.com/andrew-d/sbuild/recipes/templates"
	"github.com/andrew-d/sbuild/types"
)

type PvRecipe struct {
	*templates.BaseRecipe
}

func init() {
	builder.RegisterRecipe(&PvRecipe{})
}

func (r *PvRecipe) Info() *types.RecipeInfo {
	return &types.RecipeInfo{
		Name:         "pv",
		Version:      "1.6.0",
		Dependencies: []string{},
		Sources: []string{
			"https://www.ivarch.com/programs/sources/pv-1.6.0.tar.bz2",
		},
		Sums: []string{
			"0ece824e0da27b384d11d1de371f20cafac465e038041adab57fcf4b5036ef8d",
		},
		Binary: true,
	}
}

func (r *PvRecipe) Build(ctx *types.BuildContext) error {
	log.Info("Building pv")
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
		Append("CC", ctx.StaticFlags).
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

	log.Info("Finished building pv")
	return nil
}

func (r *PvRecipe) Finalize(ctx *types.BuildContext, outDir string) error {
	srcdir := r.UnpackedDir(ctx, r.Info())

	source := filepath.Join(srcdir, "pv")
	target := filepath.Join(outDir, "pv")

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
