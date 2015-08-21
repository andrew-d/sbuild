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

type SocatRecipe struct {
	*templates.BaseRecipe
}

func init() {
	builder.RegisterRecipe(&SocatRecipe{})
}

func (r *SocatRecipe) Info() *types.RecipeInfo {
	return &types.RecipeInfo{
		Name:         "socat",
		Version:      "1.7.3.0",
		Dependencies: []string{"openssl", "readline", "ncurses"},
		Sources: []string{
			"http://www.dest-unreach.org/socat/download/socat-${version}.tar.gz",
		},
		Sums: []string{
			"f8de4a2aaadb406a2e475d18cf3b9f29e322d4e5803d8106716a01fd4e64b186",
		},
		Binary: true,
	}
}

func (r *SocatRecipe) Build(ctx *types.BuildContext) error {
	log.Info("Building socat")
	srcdir := r.UnpackedDir(ctx, r.Info())

	var cmd *exec.Cmd

	log.Infof("Running ./configure")
	cmd = exec.Command(
		"./configure",
		"--host="+ctx.CrossPrefix,
		"--build=i686",
	)
	cmd.Dir = srcdir
	cmd.Env = ctx.Env.
		Append("CC", ctx.StaticFlags).
		Append("CPPFLAGS", "-DNETDB_INTERNAL=-1").
		Set("CFLAGS", "-fPIC "+ctx.StaticFlags).
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

	log.Info("Finished building socat")
	return nil
}

func (r *SocatRecipe) Finalize(ctx *types.BuildContext, outDir string) error {
	srcdir := r.UnpackedDir(ctx, r.Info())

	source := filepath.Join(srcdir, "socat")
	target := filepath.Join(outDir, "socat")

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
