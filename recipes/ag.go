package recipes

import (
	"io"
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
		Name:         "the_silver_searcher",
		Version:      "0.30.0",
		Dependencies: []string{"zlib", "lzma", "pcre"},
		Sources: []string{
			"the_silver_searcher-0.30.0.tar.gz::https://github.com/ggreer/the_silver_searcher/archive/0.30.0.tar.gz",
		},
		Sums: []string{
			"a3b61b80f96647dbe89c7e89a8fa7612545db6fa4a313c0ef8a574d01e7da5db",
		},
	}
}

func (r *AgRecipe) Build(ctx *types.BuildContext) error {
	log.Info("Building the_silver_searcher")
	srcdir := r.UnpackedDir(ctx, r.Info())

	var cmd *exec.Cmd
	for _, runme := range []string{"aclocal", "autoconf", "autoheader"} {
		log.Infof("Running command: %s", runme)
		cmd = exec.Command(runme)
		cmd.Dir = srcdir
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	log.Infof("Running command: automake --add-missing")
	cmd = exec.Command("automake", "--add-missing")
	cmd.Dir = srcdir
	if err := cmd.Run(); err != nil {
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

	log.Info("dep env = %+v", cmd.Env)

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

	binary, err := os.Open(source)
	if err != nil {
		return err
	}
	defer binary.Close()

	output, err := os.OpenFile(
		target,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0755)
	if err != nil {
		return err
	}
	defer output.Close()

	if _, err := io.Copy(output, binary); err != nil {
		return err
	}

	if err := r.Strip(ctx, target); err != nil {
		return err
	}

	return nil
}
