package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	flag "github.com/ogier/pflag"

	"github.com/andrew-d/sbuild/builder"
	"github.com/andrew-d/sbuild/config"
	"github.com/andrew-d/sbuild/logmgr"

	_ "github.com/andrew-d/sbuild/recipes"
)

var (
	log = logmgr.NewLogger("main")

	flagPlatform string
	flagArch     string
	flagBuildDir string
	flagVerbose  bool
)

func init() {
	flag.StringVarP(&flagPlatform, "platform", "p", "linux",
		"the platform to build for")
	flag.StringVarP(&flagArch, "arch", "a", "amd64",
		"the architecture to build for")
	flag.StringVar(&flagBuildDir, "build-dir", "/tmp/sbuild",
		"the directory to use as a build directory")
	flag.BoolVarP(&flagVerbose, "verbose", "v", false, "be verbose")
}

func main() {
	logmgr.SetOutput(os.Stderr)
	if flagVerbose {
		logmgr.SetLevel(logrus.DebugLevel)
	} else {
		logmgr.SetLevel(logrus.InfoLevel)
	}

	flag.Parse()
	conf := &config.BuildConfig{
		BuildDir:  flagBuildDir,
		OutputDir: flag.Arg(0),
		Platform:  flagPlatform,
		Arch:      flagArch,
	}

	recipes := flag.Args()[1:]

	// Special case - passing a single 'all' means build all binaries.
	if len(recipes) == 1 && recipes[0] == "all" {
		recipes = builder.AllBinaries()
	}

	log.WithField("recipes", recipes).Info("Starting build")
	err := builder.Build(recipes, conf)
	if err != nil {
		log.WithField("err", err).Error("Error building")
	} else {
		log.Info("Successfully built")
	}
}
