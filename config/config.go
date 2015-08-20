package config

// Information that must be provided in order to run a build.
type BuildConfig struct {
	// The working directory for the build.
	BuildDir string

	// The output directory for build products.
	OutputDir string

	// The operating system to build for.
	Platform string

	// The architecture to build for.
	Arch string
}
