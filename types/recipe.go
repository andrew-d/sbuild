package types

import (
	"github.com/andrew-d/sbuild/env"
)

// BuildContext encapsulates information about a particular build.
type BuildContext struct {
	// Contains a copy (or symlink) of the fetched sources, and should be used
	// to perform all build operations.
	SourceDir string

	// Environment for commands to be run in.
	Env *env.Env

	// Cross compiler prefix.
	CrossPrefix string

	// Flags to make a build static.
	StaticFlags string

	// Input configuration
	Platform string
	Arch     string

	// Environment variables from the dependencies.
	DependencyEnv map[string]map[string]string

	// Call this during Finalize() in order to add environment variables to
	// this recipe's dependents.
	AddDependentEnvVar func(key, value string)
}

// Recipe is the main interface that must be implemented by things that can
// build a library or binary.
type Recipe interface {
	// Info() retrieves information about this recipe.  It must not return nil.
	Info() *RecipeInfo

	// Prepare the build.  This step is where you should apply patches to the
	// fetched/extracted source code, for example.
	Prepare(ctx *BuildContext) error

	// Run the build.  This is where all compilation should occur.  Note that,
	// for packages with a `./configure` script, that should also be performed
	// in this function.
	Build(ctx *BuildContext) error

	// Finalize the build.  If building a binary, files from the build should
	// be copied to the output directory.  If building a library, the recipe
	// should set the appropriate flags in the per-recipe environment.
	Finalize(ctx *BuildContext, outDir string) error
}

// RecipeInfo is a struct containing information about a recipe.
type RecipeInfo struct {
	// The name of this recipe.  Cannot conflict with other names.
	Name string

	// The version of this recipe.
	Version string

	// Any dependencies of this recipe.
	Dependencies []string

	// Sources for this recipe.
	// Format:
	//    http://www.site.com/path/to/file.tar.gz
	//
	// You can prefix a path with `filename::` in order to specify the filename
	// of the downloaded file.
	Sources []string

	// SHA256 hashes for each source in `Sources`.
	Sums []string

	// Whether this is a library or binary recipe (can be both).
	Library bool
	Binary  bool
}
