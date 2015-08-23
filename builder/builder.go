package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"

	"github.com/andrew-d/sbuild/config"
	"github.com/andrew-d/sbuild/env"
	"github.com/andrew-d/sbuild/logmgr"
	"github.com/andrew-d/sbuild/types"
	"github.com/andrew-d/sbuild/util"
)

const DETERMINISTIC = false

// Keeps information about a single build.
type context struct {
	rootEnv *env.Env
	config  *config.BuildConfig
	cache   *sourceCache

	// Map of package --> environment variable map
	packageEnv map[string]map[string]string
}

var (
	log = logmgr.NewLogger("sbuild/builder")
)

// Build will run a build for the recipe with the given name and using the
// provided configuration.
func Build(recipes []string, config *config.BuildConfig) error {
	log.WithField("recipes", recipes).Info("Starting build")
	cacheDir := filepath.Join(config.BuildDir, ".cache")

	// Ensure build, output, and cache directories exist.
	for _, dir := range []string{config.BuildDir, config.OutputDir, cacheDir} {
		log.WithField("dir", dir).Debug("Ensuring directory exists")
		if err := os.Mkdir(dir, 0700); err != nil {
			if !os.IsExist(err) {
				log.WithFields(logrus.Fields{
					"dir": dir,
					"err": err,
				}).Error("Could not create directory")
				return err
			}
		}
	}

	// Get dependency order for all input recipes.
	deps, err := getRecipeDeps(recipes)
	if err != nil {
		log.WithField("err", err).Error("Could not get recipe dependencies")
		return err
	}

	cache, err := newSourceCache(cacheDir)
	if err != nil {
		log.WithField("err", err).Error("Could not create source cache")
		return err
	}

	// Make our context
	ctx := context{
		rootEnv:    env.FromOS(),
		config:     config,
		cache:      cache,
		packageEnv: make(map[string]map[string]string),
	}

	// For each dependency, we build it.
	for _, dep := range deps {
		if err = buildOne(dep, &ctx); err != nil {
			log.WithFields(logrus.Fields{
				"dep": dep,
				"err": err,
			}).Error("Error building dependency")
			return err
		}
	}

	return nil
}

func buildOne(name string, ctx *context) error {
	log.WithField("recipe", name).Info("Building single recipe")
	recipe := recipesRegistry[name]

	// Remove and re-create the source directory for this build.
	sourceDir := filepath.Join(ctx.config.BuildDir, name)
	if err := os.RemoveAll(sourceDir); err != nil {
		log.WithFields(logrus.Fields{
			"recipe": name,
			"err":    err,
		}).Error("Could not remove source directory")
		return err
	}
	if err := os.Mkdir(sourceDir, 0700); err != nil {
		log.WithFields(logrus.Fields{
			"recipe": name,
			"err":    err,
		}).Error("Could not create source directory")
		return err
	}

	info := recipe.Info()
	for i, source := range info.Sources {
		// Expand the source.
		expandedSource := os.Expand(source, func(vname string) string {
			if vname == "name" {
				return info.Name
			} else if vname == "version" {
				return info.Version
			}

			panic(fmt.Sprintf("unknown expansion variable: %s", vname))
		})

		// Fetch the source
		if err := ctx.cache.Fetch(
			name,
			expandedSource,
			info.Sums[i],
			sourceDir,
		); err != nil {
			log.WithFields(logrus.Fields{
				"recipe": name,
				"source": expandedSource,
				"hash":   info.Sums[i],
				"err":    err,
			}).Error("Could not fetch source")
			return err
		}

		filename, _ := SplitSource(expandedSource)
		sourcePath := filepath.Join(sourceDir, filename)

		// Unpack it.
		if err := util.UnpackArchive(sourcePath, sourceDir); err != nil {
			log.WithFields(logrus.Fields{
				"recipe": name,
				"source": expandedSource,
				"err":    err,
			}).Error("Could not unpack source")
			return err
		}
	}

	// Make the environment for this build.  We do this by taking the root
	// environment, and then merging in all flags from the recursive tree of
	// dependencies.
	deps := dependencyNames(name)
	env := ctx.rootEnv
	envMap := make(map[string]map[string]string)
	for _, dep := range deps {
		if flags, ok := ctx.packageEnv[dep]; ok {
			envMap[dep] = flags
			for k, v := range flags {
				env = env.Append(k, " "+v+" ")
			}
		}
	}

	// Set up cross compiler environment.
	prefix := CrossPrefix(ctx.config.Platform, ctx.config.Arch)
	env = setCrossEnv(prefix, env)

	// We special-case darwin here, since we're using osxcross.
	if ctx.config.Platform == "darwin" {
		env = env.
			Set("CC", prefix+"-clang").
			Set("CXX", prefix+"-clang++").
			Set("OSXCROSS_NO_INCLUDE_PATH_WARNINGS", "1")
	}

	// Set up the static flag
	var staticFlag string
	if ctx.config.Platform == "darwin" {
		staticFlag = " -flto -O3 -mmacosx-version-min=10.6 "
	} else {
		staticFlag = " -static "
	}

	// Set up the random seed we pass to each build.  This is a part of the
	// compiler command to ensure that it gets passed correctly.
	if DETERMINISTIC {
		randomSeed := fmt.Sprintf(
			" -frandom-seed=build-%s-%s-%s ",
			name,
			ctx.config.Platform,
			ctx.config.Arch,
		)
		env = env.Append("CC", randomSeed).Append("CXX", randomSeed)
	}

	// Run the build in this directory.
	buildCtx := types.BuildContext{
		SourceDir:     sourceDir,
		Env:           env,
		CrossPrefix:   prefix,
		StaticFlags:   staticFlag,
		Platform:      ctx.config.Platform,
		Arch:          ctx.config.Arch,
		DependencyEnv: envMap,
	}

	if err := recipe.Prepare(&buildCtx); err != nil {
		log.WithFields(logrus.Fields{
			"recipe": name,
			"err":    err,
		}).Error("Prepare failed")
		return err
	}
	if err := recipe.Build(&buildCtx); err != nil {
		log.WithFields(logrus.Fields{
			"recipe": name,
			"err":    err,
		}).Error("Build failed")
		return err
	}

	// Now, fill in the environment variable function.
	buildCtx.AddDependentEnvVar = func(key, value string) {
		mm, ok := ctx.packageEnv[name]
		if !ok {
			mm = make(map[string]string)
			ctx.packageEnv[name] = mm
		}

		mm[key] = value
	}

	// Create the output directory for this recipe.
	outDir := filepath.Join(
		ctx.config.OutputDir,
		name,
		info.Version,
	)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.WithFields(logrus.Fields{
			"recipe": name,
			"outDir": outDir,
			"err":    err,
		}).Error("Could not create output directory")
		return err
	}

	if err := recipe.Finalize(&buildCtx, outDir); err != nil {
		log.WithFields(logrus.Fields{
			"recipe": name,
			"err":    err,
		}).Error("Finalize failed")
		return err
	}

	return nil
}

// Returns a sorted list of dependencies for the given recipe name, or an error
// describing a dependency cycle.
func getRecipeDeps(recipes []string) ([]string, error) {
	depgraph := make(graph)

	var visit func(string) error
	visit = func(curr string) error {
		recipe, found := recipesRegistry[curr]
		if !found {
			return fmt.Errorf("builder: recipe %s does not exist", curr)
		}

		for _, dep := range recipe.Info().Dependencies {
			depgraph[dep] = append(depgraph[dep], curr)
			if err := visit(dep); err != nil {
				return err
			}
		}

		// Ensure that the current map entry exists.
		depgraph[curr] = depgraph[curr]
		return nil
	}

	// Calculate dependency graph.
	for _, recipe := range recipes {
		if err := visit(recipe); err != nil {
			return nil, err
		}
	}

	// Toplogically sort dependencies
	order, cycle := topologicalSort(depgraph)
	if len(cycle) > 0 {
		return nil, fmt.Errorf("builder: dependency cycle detected: %+v", cycle)
	}

	return order, nil
}
