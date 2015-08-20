package builder

import (
	"fmt"

	"github.com/andrew-d/sbuild/types"
)

var (
	recipes = make(map[string]types.Recipe)
)

func RegisterRecipe(r types.Recipe) {
	name := r.Info().Name
	if _, exists := recipes[name]; exists {
		panic("recipe with this name already exists")
	}

	recipes[name] = r
}

// Returns all dependency names for the given named recipe.  Will panic if the
// recipe name given doesn't exist in the map, or if any dependencies don't
// exist.
func dependencyNames(name string) []string {
	depNames := []string{}

	var visit func(string)
	visit = func(name string) {
		recipe, ok := recipes[name]
		if !ok {
			panic(fmt.Sprintf("recipe with name '%s' not found", name))
		}

		for _, dep := range recipe.Info().Dependencies {
			depNames = append(depNames, dep)
			visit(dep)
		}
	}

	visit(name)
	return depNames
}
