package libiconv

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

type IconvRecipe struct {
	*templates.BaseRecipe
}

var (
	log = logmgr.NewLogger("sbuild/recipes/ncurses")
)

func init() {
	builder.RegisterRecipe(&IconvRecipe{})
}

func (r *IconvRecipe) Info() *types.RecipeInfo {
	return &types.RecipeInfo{
		Name:         "libiconv",
		Version:      "1.14",
		Dependencies: nil,
		Sources: []string{
			"http://ftp.gnu.org/pub/gnu/libiconv/libiconv-${version}.tar.gz",
		},
		Sums: []string{
			"72b24ded17d687193c3366d0ebe7cde1e6b18f0df8c55438ac95be39e8a30613",
		},
		Library: true,
	}
}

func (r *IconvRecipe) Prepare(ctx *types.BuildContext) error {
	srcdir := r.UnpackedDir(ctx, r.Info())

	patches := []string{
		// From Homebrew
		"patch-Makefile.devel.patch",

		// Our patch - does the following:
		//  - Don't build a preloadable extension
		//  - Don't build the iconv executable
		//  - Remove the __inline flag that causes a build failure.
		"libiconv-build-fixes.patch",

		// Stop using gets (causes a build failure)
		"libiconv-1.14_srclib_stdio.in.h-remove-gets-declarations.patch",
	}

	// Support additional locales on darwin.
	if ctx.Platform == "darwin" {
		patches = append(patches, "patch-utf8mac.patch")
		patches = append(patches, "patch-utf8mac-flags.patch")
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

	// Fix Makefile
	cmd := exec.Command(
		"sed",
		"-i",
		"/cd preload && /d",
		filepath.Join(srcdir, "Makefile.in"),
	)
	cmd.Dir = srcdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// Replace config.sub in the directory.
	if err := util.ReplaceConfigSub(srcdir, ctx.CrossPrefix); err != nil {
		return err
	}

	return nil
}

func (r *IconvRecipe) Build(ctx *types.BuildContext) error {
	log.Info("Building libiconv")
	srcdir := r.UnpackedDir(ctx, r.Info())

	cmd := exec.Command(
		"./configure",
		"--disable-shared",
		"--enable-static",
		"--disable-debug",
		"--disable-dependency-tracking",
		"--enable-extra-encodings",
		"--host="+ctx.CrossPrefix,
		"--build=i686",
	)
	cmd.Dir = srcdir
	cmd.Env = ctx.Env.
		Set("CFLAGS", ctx.StaticFlags).
		Set("CXXFLAGS", ctx.StaticFlags).
		Set("LDFLAGS", ctx.StaticFlags).
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

	log.Info("Finished building libiconv")
	return nil
}

func (r *IconvRecipe) Finalize(ctx *types.BuildContext, outDir string) error {
	srcdir := r.UnpackedDir(ctx, r.Info())
	ctx.AddDependentEnvVar(
		"CPPFLAGS",
		"-I"+filepath.Join(srcdir, "include"),
	)
	ctx.AddDependentEnvVar(
		"LDFLAGS",
		"-L"+filepath.Join(srcdir, "lib", ".libs")+" -liconv",
	)
	return nil
}
