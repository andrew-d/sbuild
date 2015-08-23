package recipes

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"

	"github.com/andrew-d/sbuild/builder"
	"github.com/andrew-d/sbuild/recipes/templates"
	"github.com/andrew-d/sbuild/types"
)

type BinutilsRecipe struct {
	*templates.BaseRecipe
}

func init() {
	builder.RegisterRecipe(&BinutilsRecipe{})
}

func (r *BinutilsRecipe) Info() *types.RecipeInfo {
	return &types.RecipeInfo{
		Name:    "binutils",
		Version: "2.25",
		Sources: []string{
			"http://ftp.gnu.org/gnu/binutils/binutils-${version}.tar.gz",
		},
		Sums: []string{
			"cccf377168b41a52a76f46df18feb8f7285654b3c1bd69fc8265cb0fc6902f2d",
		},
		Binary: true,
	}
}

func (r *BinutilsRecipe) Dependencies(platform, arch string) []string {
	if platform == "darwin" {
		return []string{"zlib", "libiconv"}
	}

	return nil
}

func (r *BinutilsRecipe) Build(ctx *types.BuildContext) error {
	log.Info("Building binutils")
	unpackedDir := r.UnpackedDir(ctx, r.Info())

	var cmd *exec.Cmd

	// 1. Make a directory for out-of-tree building.
	buildDir := filepath.Join(ctx.SourceDir, "binutils-build")
	if err := os.Mkdir(buildDir, 0700); err != nil {
		return err
	}

	// 2. Figure out what options we have available
	log.Info("Running ./configure to get options")
	var stdout bytes.Buffer
	cmd = exec.Command(
		filepath.Join(unpackedDir, "configure"),
		"--help",
	)
	cmd.Dir = buildDir
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not get configure options")
		return err
	}

	// 3. Determine what options we're using.
	configureHelp := stdout.Bytes()
	configureOpts := []string{
		"--host=" + ctx.CrossPrefix,
		"--build=i686",
		"--target=",
	}
	for _, opt := range []string{
		"disable-nls",
		"enable-static-link",
		"disable-shared-plugins",
		"disable-dynamicplugin",
		"disable-tls",
		"disable-pie",
	} {
		if bytes.Contains(configureHelp, []byte(opt)) {
			configureOpts = append(configureOpts, "--"+opt)
		}
	}
	for _, opt := range []string{
		"enable-static",
	} {
		if bytes.Contains(configureHelp, []byte(opt)) {
			configureOpts = append(configureOpts, "--"+opt+"=yes")
		}
	}
	for _, opt := range []string{
		"enable-shared",
	} {
		if bytes.Contains(configureHelp, []byte(opt)) {
			configureOpts = append(configureOpts, "--"+opt+"=no")
		}
	}

	// 4. Make static flag.
	/*
		var staticCC, staticCXX string

		if ctx.Platform == "darwin" {
			staticCC = ctx.StaticFlags
			staticCXX = ctx.StaticFlags
		} else {
			staticCC = "-static -fPIC"
			staticCXX = "-static -static-libstdc++ -fPIC"
		}
	*/

	// 5. Actually run the configure
	log.WithField("opts", configureOpts).Info("Running ./configure")
	cmd = exec.Command(
		filepath.Join(unpackedDir, "configure"),
		configureOpts...,
	)
	cmd.Dir = buildDir
	cmd.Env = ctx.Env.
		//Append("CC", staticCC).
		//Append("CXX", staticCXX).
		Set("CFLAGS", ctx.StaticFlags).
		Set("CXXFLAGS", ctx.StaticFlags).
		AsSlice()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		log.WithField("err", err).Error("Could not run configure")
		return err
	}

	// 6. Run build.  This strange dance is actually required to get things to
	// be statically linked.
	commands := [][]string{
		{"make"},
	}

	if ctx.Platform != "darwin" {
		commands = append(commands, []string{"make", "clean"})
		commands = append(commands, []string{"make", "LDFLAGS=-all-static"})
	}

	for _, commandArr := range commands {
		cmd = exec.Command(commandArr[0], commandArr[1:]...)
		cmd.Dir = buildDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout

		if err := cmd.Run(); err != nil {
			log.WithField("err", err).Errorf("Could not run command: %s", commandArr)
			return err
		}
	}

	log.Info("Finished building binutils")
	return nil
}

func (r *BinutilsRecipe) Finalize(ctx *types.BuildContext, outDir string) error {
	buildDir := filepath.Join(ctx.SourceDir, "binutils-build")

	filenames := []string{
		filepath.Join("binutils", "ar"),
		filepath.Join("binutils", "nm-new"),
		filepath.Join("binutils", "objcopy"),
		filepath.Join("binutils", "objdump"),
		filepath.Join("binutils", "ranlib"),
		filepath.Join("binutils", "readelf"),
		filepath.Join("binutils", "size"),
		filepath.Join("binutils", "strings"),
	}

	// No 'ld' when cross-compiling to darwin.
	if ctx.Platform != "darwin" {
		filenames = append(filenames, filepath.Join("ld", "ld-new"))
	}

	for _, fname := range filenames {
		source := filepath.Join(buildDir, fname)

		// Put all files in the same directory (i.e. no subdir), and remove the
		// '-new' suffix from all binaries.
		targetFname := strings.TrimSuffix(filepath.Base(fname), "-new")
		target := filepath.Join(outDir, targetFname)

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
	}

	return nil
}
