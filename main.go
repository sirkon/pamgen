package main

import (
	"fmt"
	"os"
	"path"
	"runtime/debug"
	"sort"

	"github.com/alecthomas/kong"
	"github.com/sirkon/errors"
	"github.com/sirkon/gogh"
	"github.com/sirkon/message"
	"github.com/sirkon/pamgen/internal/app"
	"github.com/sirkon/pamgen/internal/imports"
	"github.com/sirkon/pamgen/internal/pkgscanner"
)

func main() {
	// first check if there is "-v" or "--version" flag in parameters
	for _, p := range os.Args[1:] {
		switch p {
		case "-v", "--version":
			info, ok := debug.ReadBuildInfo()
			var version string
			if !ok || info.Main.Version == "" {
				version = "(devel)"
			} else {
				version = info.Main.Version
			}
			fmt.Println(app.Name, "version", version)

			os.Exit(0)
		}
	}

	// the actual work now
	var cli cliDefinition
	parser := kong.Must(
		&cli,
		kong.Name(app.Name),
		kong.Description("proto-aware mocks generation utility"),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary: true,
			Compact: true,
		}),
		kong.UsageOnError(),
	)

	if _, err := parser.Parse(os.Args[1:]); err != nil {
		parser.FatalIfErrorf(err)
	}

	uniqIfaces := map[string]string{}
	for _, iface := range cli.Interfaces {
		uniqIfaces[iface.base] = iface.replacement
	}

	if err := run(cli.SourcePackage, cli.Destination, cli.Package, uniqIfaces, cli.Whatever); err != nil {
		message.Fatal(err)
	}
}

func run(
	srcpkgpath string,
	dest string,
	pkgname string,
	interfaces map[string]string,
	genWhatever bool,
) error {
	// generation setup
	mod, err := gogh.New(
		gogh.GoFmt,
		func(r *gogh.Imports) *imports.Imports {
			return imports.New(r)
		},
	)
	if err != nil {
		return errors.Wrap(err, "setup module")
	}

	cpkg, err := mod.Current(pkgname)
	if err != nil {
		return errors.Wrap(err, "setup a package related to the current directory")
	}

	modfilepath := path.Clean(path.Join(cpkg.Path(), dest))
	pkgpath, filename := path.Split(modfilepath)
	pkg, err := mod.Package(pkgname, path.Clean(pkgpath))
	if err != nil {
		return errors.Wrap(err, "setup destination package")
	}

	// the source package can be one of:
	//   1. standard library package
	//   2. package of one of dependencies
	//   3. some of module packages
	// Will try to open the package with the srcpkgpath first,
	srcpkg, ifaces, err := pkgscanner.GetPackageInterfaces(srcpkgpath)
	if err != nil {
		message.Debug(errors.Wrapf(err, "look for source package '%s'", srcpkgpath))
		relsrcpkgpath, m, err := pkgscanner.GetPackageInterfaces(path.Join(cpkg.Path(), srcpkgpath))
		srcpkg, ifaces, err = relsrcpkgpath, m, err
		if err != nil {
			message.Debug(errors.Wrapf(
				err,
				"look for source package '%s'",
				relsrcpkgpath.PkgPath,
			))
			return errors.New("source package not found")
		}
	}

	if len(ifaces) == 0 {
		message.Warning("no interface found")
		return nil
	}

	// dependencies are required
	if err := mod.GetDependencyLatest(gomockDep); err != nil {
		return errors.Wrap(err, "get gomock dependency")
	}
	if err := mod.GetDependencyLatest(deepequalDep); err != nil {
		return errors.Wrap(err, "get deepequal dependency")
	}

	// everything was found, ready to generate a code

	// but first warn about requested interfaces not found in the source package
	var sortedIfaces []string
	for iface := range interfaces {
		sortedIfaces = append(sortedIfaces, iface)
	}
	sort.Strings(sortedIfaces)
	for _, iface := range sortedIfaces {
		if _, ok := ifaces[iface]; !ok {
			message.Warningf("interface %s was not found in %s", iface, srcpkg.PkgPath, iface)
			continue
		}
	}

	// this can be no interface name was given, use mocks for all generated interfaces then
	if len(interfaces) == 0 {
		for name := range ifaces {
			if interfaces == nil {
				interfaces = map[string]string{}
			}
			interfaces[name] = name
			sortedIfaces = append(sortedIfaces, name)
		}
	}
	sort.Strings(sortedIfaces)

	// generate mocks
	r := pkg.Go(filename, gogh.Autogen(app.Name))

	for _, ifacename := range sortedIfaces {
		iface, ok := ifaces[ifacename]
		if !ok {
			message.Warningf("passing %s", ifacename)
			continue
		}

		generate(r, iface, interfaces[ifacename], srcpkg, ifacename, genWhatever)
	}

	if err := mod.Render(); err != nil {
		return errors.Wrap(err, "render generated code")
	}

	return nil
}
