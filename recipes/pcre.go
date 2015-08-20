package recipes

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/andrew-d/sbuild/builder"
	"github.com/andrew-d/sbuild/recipes/templates"
	"github.com/andrew-d/sbuild/types"
)

type PcreRecipe struct {
	*templates.BaseRecipe
}

func init() {
	builder.RegisterRecipe(&PcreRecipe{})
}

func (r *PcreRecipe) Info() *types.RecipeInfo {
	return &types.RecipeInfo{
		Name:         "pcre",
		Version:      "8.37",
		Dependencies: nil,
		Sources: []string{
			"http://downloads.sourceforge.net/project/pcre/pcre/8.37/pcre-8.37.tar.bz2",
		},
		Sums: []string{
			"51679ea8006ce31379fb0860e46dd86665d864b5020fc9cd19e71260eef4789d",
		},
	}
}

func (r *PcreRecipe) Build(ctx *types.BuildContext) error {
	log.Info("Building PCRE")
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

	log.Info("Finished building PCRE")
	return nil
}

func (r *PcreRecipe) Finalize(ctx *types.BuildContext) error {
	srcdir := r.UnpackedDir(ctx, r.Info())
	ctx.AddDependentEnvVar(
		"CPPFLAGS",
		"-I"+srcdir,
	)
	ctx.AddDependentEnvVar(
		"LDFLAGS",
		"-L"+filepath.Join(srcdir, ".libs")+" -lpcre",
	)
	return nil
}
