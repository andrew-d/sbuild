package recipes

import (
	"os"
	"os/exec"

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
