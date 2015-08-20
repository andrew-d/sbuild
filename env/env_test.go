package env

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstructors(t *testing.T) {
	one := FromOS()
	assert.True(t, len(one.vars) > 0)

	two := Empty()
	assert.Len(t, two.vars, 0)
}

func TestGetSet(t *testing.T) {
	env := Empty()

	env = env.
		Set("one", "1").
		Set("two", "2")

	assert.Len(t, env.vars, 2)

	assert.Equal(t, "2", env.Get("two"))

	s, ok := env.GetOk("two")
	assert.True(t, ok)
	assert.Equal(t, "2", s)

	s, ok = env.GetOk("nonexisting")
	assert.False(t, ok)
	assert.Equal(t, "", s)

	// To get around map ordering, we sort.
	vars := env.AsSlice()
	sort.Strings(vars)
	assert.Equal(t, []string{"one=1", "two=2"}, vars)
}

func TestDelete(t *testing.T) {
	env := Empty()

	env = env.Set("one", "1")
	assert.Equal(t, []string{"one=1"}, env.AsSlice())

	env = env.Delete("one")
	assert.Equal(t, []string{}, env.AsSlice())
}

func TestAppend(t *testing.T) {
	env := Empty()

	env = env.Set("one", "1")
	assert.Equal(t, []string{"one=1"}, env.AsSlice())

	env = env.Append("one", " 1")
	assert.Equal(t, []string{"one=1 1"}, env.AsSlice())
}

// Test that we can append a key that doesn't exist.
func TestAppendEmpty(t *testing.T) {
	env := Empty()

	env = env.Append("one", "1")
	assert.Equal(t, []string{"one=1"}, env.AsSlice())
}

func TestMerge(t *testing.T) {
	first := Empty()
	first = first.Append("one", "1")

	second := Empty()
	second = second.Append("one", "2")

	n := first.Merge(second)
	assert.Equal(t, "2", n.Get("one"))
}

func TestMergeAppend(t *testing.T) {
	first := Empty()
	first = first.Append("key", " foo ")

	second := Empty()
	second = second.Append("key", " bar ")

	n := first.MergeAppend(second)
	assert.Equal(t, " foo  bar ", n.Get("key"))
}
