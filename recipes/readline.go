package recipes

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/andrew-d/sbuild/builder"
	"github.com/andrew-d/sbuild/recipes/templates"
	"github.com/andrew-d/sbuild/types"
)

type ReadlineRecipe struct {
	*templates.BaseRecipe
}

func init() {
	builder.RegisterRecipe(&ReadlineRecipe{})
}

func (r *ReadlineRecipe) Info() *types.RecipeInfo {
	return &types.RecipeInfo{
		Name:         "readline",
		Version:      "6.3",
		Dependencies: nil,
		Sources: []string{
			"ftp://ftp.gnu.org/gnu/readline/readline-${version}.tar.gz",
		},
		Sums: []string{
			"56ba6071b9462f980c5a72ab0023893b65ba6debb4eeb475d7a563dc65cafd43",
		},
		Library: true,
	}
}

func (r *ReadlineRecipe) Prepare(ctx *types.BuildContext) error {
	srcdir := r.UnpackedDir(ctx, r.Info())

	// Prevent building examples, which don't work when cross-compiling.
	cmd := exec.Command(
		"sed",
		"-i",
		"s|examples/Makefile||g",
		filepath.Join(srcdir, "configure.ac"),
	)
	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not patch configure.ac")
		return err
	}

	cmd = exec.Command("autoconf")
	cmd.Dir = srcdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not run autoconf")
		return err
	}

	return nil
}

func (r *ReadlineRecipe) Build(ctx *types.BuildContext) error {
	log.Info("Building readline")
	srcdir := r.UnpackedDir(ctx, r.Info())

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

	log.Info("Finished building readline")
	return nil
}

func (r *ReadlineRecipe) Finalize(ctx *types.BuildContext, outDir string) error {
	srcdir := r.UnpackedDir(ctx, r.Info())
	ctx.AddDependentEnvVar(
		"LDFLAGS",
		"-L"+srcdir+" -lreadline",
	)

	// Note: most things will attempt to #include readline as
	// `readline/readline.h`.  We create a symlink such that this works.
	oldname := srcdir
	newname := filepath.Join(ctx.SourceDir, "readline")
	if err := os.Symlink(oldname, newname); err != nil {
		return err
	}

	ctx.AddDependentEnvVar(
		"CPPFLAGS",
		"-I"+ctx.SourceDir,
	)
	return nil
}
