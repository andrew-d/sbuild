package tar

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
)

type TarRecipe struct {
	*templates.BaseRecipe
}

var (
	log = logmgr.NewLogger("sbuild/recipes/ncurses")
)

func init() {
	builder.RegisterRecipe(&TarRecipe{})
}

func (r *TarRecipe) Info() *types.RecipeInfo {
	return &types.RecipeInfo{
		Name:         "tar",
		Version:      "1.28",
		Dependencies: []string{"libiconv"},
		Sources: []string{
			"https://ftp.gnu.org/gnu/tar/tar-${version}.tar.xz",
		},
		Sums: []string{
			"64ee8d88ec1b47a0961033493f919d27218c41b580138fd6802327462aff22f2",
		},
		Binary: true,
	}
}

func (r *TarRecipe) Prepare(ctx *types.BuildContext) error {
	srcdir := r.UnpackedDir(ctx, r.Info())

	patches := []string{
		"tar-0001-fix-build-failure.patch",
	}

	if ctx.Platform == "darwin" {
		patches = append(patches, "gnutar-configure-xattrs.patch")
	}

	// Apply patches
	for _, patch := range patches {
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

	return nil
}

func (r *TarRecipe) Build(ctx *types.BuildContext) error {
	log.Info("Building tar")
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

	// 2. Run build
	cmd = exec.Command("make")
	cmd.Dir = srcdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not run make")
		return err
	}

	log.Info("Finished building tar")
	return nil
}

func (r *TarRecipe) Finalize(ctx *types.BuildContext, outDir string) error {
	srcdir := r.UnpackedDir(ctx, r.Info())

	source := filepath.Join(srcdir, "src", "tar")
	target := filepath.Join(outDir, "tar")

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
