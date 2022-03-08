package main

import (
	"go/ast"
	"go/parser"
	"strings"

	"github.com/sirkon/errors"
)

type cliDefinition struct {
	SourcePackage string          `short:"s" help:"Source package to look for interfaces in."`
	Destination   string          `short:"d" help:"Destination file to store generated mocks in with the current directory as a root." required:""`
	Package       string          `short:"p" help:"Package name for generated file. Will be ignored if the package exists and has different name."`
	Interfaces    []interfaceName `arg:"" help:"List of interfaces names to generate mocks for" optional:""`
	Version       bool            `short:"v" help:"Show version and exit."`
}

type packagePath string

type interfaceName struct {
	base        string
	replacement string
}

func (n *interfaceName) UnmarshalText(data []byte) error {
	text := string(data)
	name, replacement, found := strings.Cut(text, "=")
	if !found {
		replacement = name
	}

	ifacename, err := parser.ParseExpr(name)
	if err != nil {
		return errors.Newf("invalid interface name '%s'", text)
	}

	switch ifacename.(type) {
	case *ast.Ident:
	default:
		return errors.Newf("invalid interface name '%s'", text)
	}

	replacementname, err := parser.ParseExpr(replacement)
	if err != nil {
		return errors.Newf("invalid replacement name '%s'", text)
	}

	switch replacementname.(type) {
	case *ast.Ident:
	default:
		return errors.Newf("invalid replacement name '%s'", text)
	}

	n.base = name
	n.replacement = replacement
	return nil
}
