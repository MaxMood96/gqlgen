package code

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestPackages(t *testing.T) {
	t.Run("name for existing package does not load again", func(t *testing.T) {
		p := initialState(t)
		require.Equal(t, "a", p.NameForPackage("github.com/99designs/gqlgen/internal/code/testdata/a"))
		require.Equal(t, 1, p.numLoadCalls)
	})

	t.Run("name for unknown package makes name only load", func(t *testing.T) {
		p := initialState(t)
		require.Equal(t, "c", p.NameForPackage("github.com/99designs/gqlgen/internal/code/testdata/c"))
		require.Equal(t, 1, p.numLoadCalls)
		require.Equal(t, 1, p.numNameCalls)
	})

	t.Run("evicting a package causes it to load again", func(t *testing.T) {
		p := initialState(t)
		p.Evict("github.com/99designs/gqlgen/internal/code/testdata/b")
		require.Equal(t, "a", p.Load("github.com/99designs/gqlgen/internal/code/testdata/a").Name)
		require.Equal(t, 1, p.numLoadCalls)
		require.Equal(t, "b", p.Load("github.com/99designs/gqlgen/internal/code/testdata/b").Name)
		require.Equal(t, 2, p.numLoadCalls)
	})

	t.Run("able to load private package with build tags", func(t *testing.T) {
		p := initialState(t, WithBuildTags("private"))
		p.Evict("github.com/99designs/gqlgen/internal/code/testdata/a")
		require.Equal(t, "a", p.Load("github.com/99designs/gqlgen/internal/code/testdata/a").Name)
		require.Equal(t, 2, p.numLoadCalls)
		require.Equal(t, "p", p.Load("github.com/99designs/gqlgen/internal/code/testdata/p").Name)
		require.Equal(t, 3, p.numLoadCalls)
	})
}

func TestPackagesErrors(t *testing.T) {
	loadFirstErr := errors.New("first")
	loadSecondErr := errors.New("second")
	packageErr := packages.Error{Msg: "package"}
	p := &Packages{
		loadErrors: []error{loadFirstErr, loadSecondErr},
		packages: map[string]*packages.Package{"github.com/99designs/gqlgen/internal/code/testdata/a": {
			Errors: []packages.Error{packageErr},
		}},
	}

	errs := p.Errors()

	assert.Equal(t, PkgErrors([]error{loadFirstErr, loadSecondErr, packageErr}), errs)
}

func TestNameForPackage(t *testing.T) {
	var p Packages

	assert.Equal(t, "api", p.NameForPackage("github.com/99designs/gqlgen/api"))

	// does not contain go code, should still give a valid name
	assert.Equal(t, "docs", p.NameForPackage("github.com/99designs/gqlgen/docs"))
	assert.Equal(t, "github_com", p.NameForPackage("github.com"))
}

func TestLoadAllNames(t *testing.T) {
	var p Packages

	p.LoadAllNames("github.com/99designs/gqlgen/api", "github.com/99designs/gqlgen/docs", "github.com")

	// should now be cached
	assert.Equal(t, 0, p.numNameCalls)
	assert.Equal(t, "api", p.importToName["github.com/99designs/gqlgen/api"])
	assert.Equal(t, "docs", p.importToName["github.com/99designs/gqlgen/docs"])
	assert.Equal(t, "github_com", p.importToName["github.com"])
}

func initialState(t *testing.T, opts ...Option) *Packages {
	p := NewPackages(opts...)
	pkgs := p.LoadAll(
		"github.com/99designs/gqlgen/internal/code/testdata/a",
		"github.com/99designs/gqlgen/internal/code/testdata/b",
	)

	require.Empty(t, p.Errors())
	require.Equal(t, 1, p.numLoadCalls)
	require.Equal(t, 0, p.numNameCalls)
	require.Equal(t, "a", pkgs[0].Name)
	require.Equal(t, "b", pkgs[1].Name)
	return p
}
