package builder

import (
	"fmt"

	"github.com/andrew-d/sbuild/env"
)

// Default Darwin version supported
const DARWIN_VERSION = 12

// Get the cross-compiler prefix for a given platform/arch combination.
// Returns the empty string if unknown.
func CrossPrefix(platform, arch string) string {
	switch platform {
	case "linux":
		switch arch {
		// TODO: x86
		case "amd64":
			return "x86_64-linux-musl"
		case "arm":
			return "arm-linux-musleabihf"
		}

	case "android":
		return "arm-linux-musleabihf"

	case "darwin":
		return fmt.Sprintf("x86_64-apple-darwin%d", DARWIN_VERSION)
	}

	return ""
}

// Sets the appropriate tool flags in the provided environment, given the cross
// compiler prefix.
func setCrossEnv(prefix string, env *env.Env) *env.Env {
	env = env.
		Set("AR", prefix+"-ar").
		Set("CC", prefix+"-gcc").
		Set("CXX", prefix+"-g++").
		Set("LD", prefix+"-ld").
		Set("RANLIB", prefix+"-ranlib").
		Set("STRIP", prefix+"-strip")
	return env
}
