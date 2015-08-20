package recipes

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/andrew-d/sbuild/builder"
	"github.com/andrew-d/sbuild/recipes/templates"
	"github.com/andrew-d/sbuild/types"
)

type ZlibRecipe struct {
	*templates.BaseRecipe
}

func init() {
	builder.RegisterRecipe(&ZlibRecipe{})
}

func (r *ZlibRecipe) Info() *types.RecipeInfo {
	return &types.RecipeInfo{
		Name:         "zlib",
		Version:      "1.2.8",
		Dependencies: nil,
		Sources: []string{
			"http://zlib.net/zlib-1.2.8.tar.gz",
		},
		Sums: []string{
			"36658cb768a54c1d4dec43c3116c27ed893e88b02ecfcb44f2166f9c0b7f2a0d",
		},
	}
}

func (r *ZlibRecipe) Build(ctx *types.BuildContext) error {
	log.Info("Building zlib")
	srcdir := r.UnpackedDir(ctx, r.Info())

	// 1. Configure
	cmd := exec.Command("./configure", "--static")
	cmd.Dir = srcdir
	cmd.Env = ctx.Env.
		Set("CHOST", ctx.CrossPrefix).
		Set("CFLAGS", ctx.StaticFlags).
		Append("CC", ctx.StaticFlags).
		AsSlice()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not run configure")
		return err
	}

	// 2. Fix path to libtool when cross-compiling to Darwin
	if ctx.Platform == "darwin" {
		cmd = exec.Command(
			"sed",
			"-i",
			"-e",
			fmt.Sprintf("s|AR=/usr/bin/libtool|AR=%s-ar|g", ctx.CrossPrefix),
			"-e",
			"s|ARFLAGS=-o|ARFLAGS=rc|g",
			filepath.Join(srcdir, "Makefile"),
		)
		if err := cmd.Run(); err != nil {
			log.WithField("err", err).Error("Could not patch Makefile")
			return err
		}
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

	log.Info("Finished building zlib")
	return nil
}

func (r *ZlibRecipe) Finalize(ctx *types.BuildContext, outDir string) error {
	srcdir := r.UnpackedDir(ctx, r.Info())
	ctx.AddDependentEnvVar(
		"CPPFLAGS",
		"-I"+srcdir,
	)
	ctx.AddDependentEnvVar(
		"LDFLAGS",
		"-L"+srcdir+" -lz",
	)
	return nil
}
