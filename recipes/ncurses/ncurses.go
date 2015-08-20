package ncurses

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Sirupsen/logrus"

	"github.com/andrew-d/sbuild/builder"
	"github.com/andrew-d/sbuild/logmgr"
	"github.com/andrew-d/sbuild/recipes/templates"
	"github.com/andrew-d/sbuild/types"
	"github.com/andrew-d/sbuild/util"
)

type NcursesRecipe struct {
	*templates.BaseRecipe
}

var (
	log = logmgr.NewLogger("sbuild/recipes/ncurses")
)

func init() {
	builder.RegisterRecipe(&NcursesRecipe{})
}

func (r *NcursesRecipe) Info() *types.RecipeInfo {
	return &types.RecipeInfo{
		Name:         "ncurses",
		Version:      "5.9",
		Dependencies: nil,
		Sources: []string{
			"ncurses-5.9.tar.gz::http://invisible-island.net/datafiles/release/ncurses.tar.gz",
		},
		Sums: []string{
			"9046298fb440324c9d4135ecea7879ffed8546dd1b58e59430ea07a4633f563b",
		},
	}
}

func (r *NcursesRecipe) Prepare(ctx *types.BuildContext) error {
	srcdir := r.UnpackedDir(ctx, r.Info())

	// Apply patches
	for _, patch := range AssetNames() {
		log.WithField("patch", patch).Info("Applying patch")

		cmd := exec.Command("patch", "-p1")
		cmd.Dir = srcdir
		cmd.Stdin = bytes.NewBuffer(MustAsset(patch))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.WithFields(logrus.Fields{
				"patch": patch,
				"err":   err,
			}).Error("Error applying patch")
			return err
		}
	}

	// Replace config.sub in the directory.
	if err := util.ReplaceConfigSub(srcdir, ctx.CrossPrefix); err != nil {
		return err
	}

	return nil
}

func (r *NcursesRecipe) Build(ctx *types.BuildContext) error {
	log.Info("Building ncurses")
	srcdir := r.UnpackedDir(ctx, r.Info())

	cmd := exec.Command(
		"./configure",
		"--disable-shared",
		"--enable-static",
		"--with-normal",
		"--without-debug",
		"--without-ada",
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

	log.Info("Finished building ncurses")
	return nil
}

func (r *NcursesRecipe) Finalize(ctx *types.BuildContext, outDir string) error {
	srcdir := r.UnpackedDir(ctx, r.Info())
	ctx.AddDependentEnvVar(
		"CPPFLAGS",
		"-I"+srcdir,
	)
	ctx.AddDependentEnvVar(
		"LDFLAGS",
		"-L"+filepath.Join(srcdir, "lib")+" -lncurses",
	)
	return nil
}
