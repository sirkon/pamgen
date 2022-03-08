package imports

import "github.com/sirkon/gogh"

// New creates new Imports
func New(i *gogh.Imports) *Imports {
	return &Imports{i: i}
}

// Imports a gogh.Importer implementation to deal with imports of frequent imports needed in mocks
type Imports struct {
	i *gogh.Imports
}

// Imports to satisfy gogh.Importer
func (i *Imports) Imports() *gogh.Imports {
	return i.i
}

// Add to satisfy gogh.Importer
func (i *Imports) Add(pkgpath string) *gogh.ImportAliasControl {
	return i.i.Add(pkgpath)
}

// Module to satisfy gogh.Importer
func (i *Imports) Module(relpath string) *gogh.ImportAliasControl {
	return i.i.Module(relpath)
}

// Errors imports stdlib errors
func (i *Imports) Errors() *gogh.ImportAliasControl {
	return i.i.Add("errors")
}

// Reflect imports stdlib reflect
func (i *Imports) Reflect() *gogh.ImportAliasControl {
	return i.i.Add("reflect")
}

// Deepequal imports github.com/sirkon/deepequal
func (i *Imports) Deepequal() *gogh.ImportAliasControl {
	return i.i.Add("github.com/sirkon/deepequal")
}

// Gomock imports github.com/golang/mock/gomock
func (i *Imports) Gomock() *gogh.ImportAliasControl {
	return i.i.Add("github.com/golang/mock/gomock")
}

var (
	_ gogh.Importer = &Imports{}
)
