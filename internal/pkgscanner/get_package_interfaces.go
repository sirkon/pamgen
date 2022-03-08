package pkgscanner

import (
	"go/token"
	"go/types"

	"github.com/sirkon/errors"
	"golang.org/x/tools/go/packages"
)

// GetPackageInterfaces look for interfaces
func GetPackageInterfaces(pkpgpath string) (*packages.Package, map[string]*types.Interface, error) {
	mode := packages.NeedImports | packages.NeedTypes | packages.NeedName |
		packages.NeedDeps | packages.NeedSyntax | packages.NeedFiles | packages.NeedModule

	pkgs, err := packages.Load(
		&packages.Config{
			Mode:  mode,
			Fset:  token.NewFileSet(),
			Tests: false,
		},
		pkpgpath,
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "load package")
	}

	if len(pkgs) == 0 {
		return nil, nil, errors.New("nothing loaded")
	}
	pkg := pkgs[0]

	res := map[string]*types.Interface{}
	for _, typeName := range pkg.Types.Scope().Names() {
		t := pkg.Types.Scope().Lookup(typeName)
		if v := digInterface(nil, t.Type()); v != nil {
			res[typeName] = v
		}
	}

	return pkg, res, nil
}

func digInterface(prev types.Type, t types.Type) *types.Interface {
	switch v := t.(type) {
	case *types.Interface:
		return v
	default:
		u := t.Underlying()
		if u == t {
			return nil
		}
		if t.Underlying() != nil {
			return digInterface(t, t.Underlying())
		}

		return nil
	}
}
