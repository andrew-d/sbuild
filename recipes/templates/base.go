package templates

import (
	"fmt"
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
