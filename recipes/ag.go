package recipes

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Sirupsen/logrus"

	"github.com/andrew-d/sbuild/builder"
	"github.com/andrew-d/sbuild/recipes/templates"
	"github.com/andrew-d/sbuild/types"
)

type AgRecipe struct {
	*templates.BaseRecipe
}

func init() {
	builder.RegisterRecipe(&AgRecipe{})
}

func (r *AgRecipe) Info() *types.RecipeInfo {
	return &types.RecipeInfo{
		Name:    "the_silver_searcher",
		Version: "0.30.0",
		Sources: []string{
			"${name}-${version}.tar.gz::https://github.com/ggreer/the_silver_searcher/archive/${version}.tar.gz",
		},
		Sums: []string{
			"a3b61b80f96647dbe89c7e89a8fa7612545db6fa4a313c0ef8a574d01e7da5db",
		},
		Binary: true,
	}
}

func (r *AgRecipe) Dependencies(platform, arch string) []string {
	return []string{"zlib", "lzma", "pcre"}
}

func (r *AgRecipe) Build(ctx *types.BuildContext) error {
	log.Info("Building the_silver_searcher")
	srcdir := r.UnpackedDir(ctx, r.Info())

	var cmd *exec.Cmd

	// Run autotools
	cmd = exec.Command("autoreconf", "-i")
	cmd.Dir = srcdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not run autotools")
		return err
	}

	log.Infof("Running ./configure")
	cmd = exec.Command(
		"./configure",
		"--host="+ctx.CrossPrefix,
		"--build=i686",
		"PKG_CONFIG=/bin/true",
	)
	cmd.Dir = srcdir
	cmd.Env = ctx.Env.
		Append("CC", ctx.StaticFlags).
		Set("CFLAGS", "-fPIC "+ctx.StaticFlags).
		Set("PCRE_LIBS", ctx.DependencyEnv["pcre"]["LDFLAGS"]).
		Set("PCRE_CFLAGS", ctx.DependencyEnv["pcre"]["CFLAGS"]).
		Set("LZMA_LIBS", ctx.DependencyEnv["lzma"]["LDFLAGS"]).
		Set("LZMA_CFLAGS", ctx.DependencyEnv["lzma"]["CFLAGS"]).
		Set("ZLIB_LIBS", ctx.DependencyEnv["zlib"]["LDFLAGS"]).
		Set("ZLIB_CFLAGS", ctx.DependencyEnv["zlib"]["CFLAGS"]).
		AsSlice()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not run configure")
		return err
	}

	cmd = exec.Command("make")
	cmd.Dir = srcdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not run make")
		return err
	}

	log.Info("Finished building the_silver_searcher")
	return nil
}

func (r *AgRecipe) Finalize(ctx *types.BuildContext, outDir string) error {
	srcdir := r.UnpackedDir(ctx, r.Info())

	source := filepath.Join(srcdir, "ag")
	target := filepath.Join(outDir, "ag")

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
