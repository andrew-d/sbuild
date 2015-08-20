package templates

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/andrew-d/sbuild/types"
)

type BaseRecipe struct{}

func (r *BaseRecipe) Prepare(ctx *types.BuildContext) error {
	// Do nothing by default.
	return nil
}

func (r *BaseRecipe) Finalize(ctx *types.BuildContext, outDir string) error {
	// Do nothing by default.
	return nil
}

// UnpackedDir returns the path to the unpacked source, assuming the default
// form (i.e. $SourceDir/$Name-$Version
func (r *BaseRecipe) UnpackedDir(ctx *types.BuildContext, info *types.RecipeInfo) string {
	return filepath.Join(
		ctx.SourceDir,
		fmt.Sprintf("%s-%s", info.Name, info.Version),
	)
}

// Strip will run the environment's strip command on the given file.
func (r *BaseRecipe) Strip(ctx *types.BuildContext, file string) error {
	cmd := exec.Command(
		ctx.Env.Get("STRIP"),
		file,
	)
	return cmd.Run()
}

// CopyFile will copy a file from one location to another.
func (r *BaseRecipe) CopyFile(source, target string, mode os.FileMode) error {
	sourcef, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourcef.Close()

	targetf, err := os.OpenFile(
		target,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		mode)
	if err != nil {
		return err
	}
	defer targetf.Close()

	if _, err := io.Copy(targetf, sourcef); err != nil {
		return err
	}

	return nil
}
