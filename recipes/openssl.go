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

type OpenSSLRecipe struct {
	*templates.BaseRecipe
}

func init() {
	builder.RegisterRecipe(&OpenSSLRecipe{})
}

func (r *OpenSSLRecipe) Info() *types.RecipeInfo {
	return &types.RecipeInfo{
		Name:         "openssl",
		Version:      "1.0.2d",
		Dependencies: nil,
		Sources: []string{
			"https://openssl.org/source/openssl-1.0.2d.tar.gz",
		},
		Sums: []string{
			"671c36487785628a703374c652ad2cebea45fa920ae5681515df25d9f2c9a8c8",
		},
	}
}

func (r *OpenSSLRecipe) Build(ctx *types.BuildContext) error {
	log.Info("Building openssl")
	srcdir := r.UnpackedDir(ctx, r.Info())

	// 1. Figure out what OpenSSL target we're using based on the input
	// configuration.
	var target string

	if ctx.Platform == "linux" {
		if ctx.Arch == "amd64" {
			target = "linux-x86_64"
		} else if ctx.Arch == "arm" {
			target = "linux-armv4"
		}
	} else if ctx.Platform == "android" {
		target = "android-armv7"
	} else if ctx.Platform == "darwin" {
		if ctx.Arch == "amd64" {
			target = "darwin64-x86_64-cc"
		} else if ctx.Arch == "x86" {
			target = "darwin-i386-cc"
		}
	}

	if target == "" {
		return fmt.Errorf("cannot build openssl for platform/arch: %s/%s",
			ctx.Platform, ctx.Arch)
	}

	// 2. Configure OpenSSL
	cmd := exec.Command(
		"perl",
		"./Configure",
		"no-shared",
		target,

		// Accelerated NIST P-224 and P-256 encryption support.
		"enable-ec_nistp_64_gcc_12",
	)
	cmd.Dir = srcdir
	cmd.Env = ctx.Env.
		Set("CFLAGS", ctx.StaticFlags).
		Set("CXXFLAGS", ctx.StaticFlags).
		AsSlice()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not run Configure")
		return err
	}

	cmd = exec.Command("make", "build_libs")
	cmd.Dir = srcdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not run make")
		return err
	}

	log.Info("Finished building openssl")
	return nil
}

func (r *OpenSSLRecipe) Finalize(ctx *types.BuildContext, outDir string) error {
	srcdir := r.UnpackedDir(ctx, r.Info())
	ctx.AddDependentEnvVar(
		"CPPFLAGS",
		"-I"+filepath.Join(srcdir, "include"),
	)
	ctx.AddDependentEnvVar(
		"LDFLAGS",
		"-L"+srcdir+" -lcrypto -lssl",
	)
	return nil
}
