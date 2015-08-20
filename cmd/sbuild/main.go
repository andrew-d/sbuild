package main

import (
	"os"

	"github.com/Sirupsen/logrus"

	"github.com/andrew-d/sbuild/builder"
	"github.com/andrew-d/sbuild/config"
	"github.com/andrew-d/sbuild/logmgr"

	_ "github.com/andrew-d/sbuild/recipes"
)

var (
	log = logmgr.NewLogger("main")
)

func main() {
	logmgr.SetOutput(os.Stderr)
	logmgr.SetLevel(logrus.DebugLevel)

	conf := &config.BuildConfig{
		BuildDir:  "/tmp/sbuild",
		OutputDir: "/tmp/sout",
		Platform:  "linux",
		Arch:      "arm",
	}

	err := builder.Build(os.Args[1], conf)
	if err != nil {
		log.WithField("err", err).Error("Error building")
	} else {
		log.Info("Successfully built")
	}
}
