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

type LzmaRecipe struct {
	*templates.BaseRecipe
}

func init() {
	builder.RegisterRecipe(&LzmaRecipe{})
}

func (r *LzmaRecipe) Info() *types.RecipeInfo {
	return &types.RecipeInfo{
		Name:         "lzma",
		Version:      "5.0.8",
		Dependencies: nil,
		Sources: []string{
			"http://tukaani.org/xz/xz-${version}.tar.gz",
		},
		Sums: []string{
			"cac71b31ed322a487f1da1f10dfcf47f8855f97ff2c23b92680c7ae7be58babb",
		},
		Library: true,
	}
}

func (r *LzmaRecipe) Build(ctx *types.BuildContext) error {
	log.Info("Building LZMA")
	srcdir := filepath.Join(ctx.SourceDir, fmt.Sprintf("xz-%s", r.Info().Version))

	cmd := exec.Command(
		"./configure",
		"--disable-shared",
		"--enable-static",
		"--host="+ctx.CrossPrefix,
		"--build=i686",
	)
	cmd.Dir = srcdir
	cmd.Env = ctx.Env.
		Set("CFLAGS", ctx.StaticFlags).
		Set("CXXFLAGS", ctx.StaticFlags).
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

	log.Info("Finished building LZMA")
	return nil
}

func (r *LzmaRecipe) Finalize(ctx *types.BuildContext, outDir string) error {
	srcdir := filepath.Join(ctx.SourceDir, fmt.Sprintf("xz-%s", r.Info().Version))
	ctx.AddDependentEnvVar(
		"CPPFLAGS",
		"-I"+filepath.Join(srcdir, "src", "liblzma", "api"),
	)
	ctx.AddDependentEnvVar(
		"LDFLAGS",
		"-L"+filepath.Join(srcdir, "src", "liblzma", ".libs")+" -llzma",
	)
	return nil
}
