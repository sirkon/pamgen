package main

import (
	"go/types"
	"strings"
	"unicode"

	"github.com/sirkon/gogh"
	"github.com/sirkon/pamgen/internal/imports"
	"golang.org/x/tools/go/packages"
)

// generate generates mock for the given interface. Uses given "replace" as a base for mock interface.
func generate(
	r *gogh.GoRenderer[*imports.Imports],
	iface *types.Interface,
	replace string,
	srcpkg *packages.Package,
	ifacename string,
	whatever bool,
) {
	constructor := "New" + replace
	if replace[:1] != strings.ToUpper(replace[:1]) {
		constructor = "new" + strings.Title(replace)
	}

	r.Imports().Gomock().Ref("gomock")

	r.L(`// $0Mock interface $1.$2 mock`, replace, srcpkg.PkgPath, ifacename)
	r.L(`type $0Mock struct{`, replace)
	r.L(`    ctrl *$gomock.Controller`)
	r.L(`    recorder *$0MockRecorder`, replace)
	r.L(`}`)
	r.N()
	r.L(`// $0MockRecorder records expected calls of $1.$2`, replace, srcpkg.PkgPath, ifacename)
	r.L(`type $0MockRecorder struct{`, replace)
	r.L(`    mock *$0Mock`, replace)
	r.L(`}`)
	r.N()
	r.L(`// $0Mock creates $1Mock instance`, constructor, replace)
	r.L(`func $0Mock(ctrl *$gomock.Controller) *$1Mock {`, constructor, replace)
	r.L(`    mock := &$0Mock{`, replace)
	r.L(`        ctrl: ctrl,`)
	r.L(`    }`)
	r.L(`    mock.recorder = &$0MockRecorder{mock: mock}`, replace)
	r.L(`    return mock`)
	r.L(`}`)
	r.N()
	r.L(`// EXPECT returns expected calls recorder`)
	r.L(`func (m *$0Mock) EXPECT() *$0MockRecorder {`, replace)
	r.L(`    return m.recorder`)
	r.L(`}`)
	r.N()

	for i := 0; i < iface.NumMethods(); i++ {
		m := iface.Method(i)
		s := m.Type().(*types.Signature)

		q := r.Scope()
		q.Uniq("m")
		q.Uniq("mr")
		q.Uniq("ret")

		if !s.Variadic() {
			generateRegularMethod(q, iface, replace, srcpkg, ifacename, m, s, whatever)
		} else {
			generateVariadicMethod(q, iface, replace, srcpkg, ifacename, m, s, whatever)
		}
	}
}

func generateRegularMethod(
	r *gogh.GoRenderer[*imports.Imports],
	iface *types.Interface,
	replace string,
	srcpkg *packages.Package,
	ifacename string,
	m *types.Func,
	s *types.Signature,
	whatever bool,
) {
	r.N()
	r.L(`// $0 method to implement $1.$2`, m.Name(), srcpkg.PkgPath, ifacename)

	var params gogh.Params
	var args gogh.Commas
	var argnames []string

	for i := 0; i < s.Params().Len(); i++ {
		p := s.Params().At(i)

		var name string
		switch p.Name() {
		case "_", "":
			name = "arg"
		default:
			name = p.Name()
		}
		argname := r.Uniq(name)
		argnames = append(argnames, argname)
		args.Add(argname)
		params.Add(argname, r.Type(p.Type()))
	}

	var returns gogh.Params
	var rets gogh.Commas
	var retnames []string
	for i := 0; i < s.Results().Len(); i++ {
		v := s.Results().At(i)

		base := v.Name()
		if base == "" || base == "_" {
			if s.Results().At(i).Type().String() == "error" {
				base = "err"
			} else {
				base = "r"
			}
		}
		retname := r.Uniq(base)
		retnames = append(retnames, retname)
		returns.Add(retname, r.Type(v.Type()))
		rets.Add(retname)
	}

	r.L(`func (m *$0Mock) $1($2) ($3) {`, replace, m.Name(), params, returns)
	r.L(`    m.ctrl.T.Helper()`)
	if s.Results().Len() > 0 {
		r.L(`    ret := m.ctrl.Call(m, "$0", $1)`, m.Name(), args)
	} else {
		r.L(`    m.ctrl.Call(m, "$0", $1)`, m.Name(), args)
	}
	for i := 0; i < s.Results().Len(); i++ {
		r.L(`    $0, _ = ret[$1].($2)`, retnames[i], i, r.Type(s.Results().At(i).Type()))
	}
	r.L(`    return $0`, rets)
	r.L(`}`)

	recorderArgs := args.String()
	if recorderArgs != "" {
		recorderArgs += " any"
	}

	r.N()
	r.L(` // $0 register expected call of method $1.$2.$0`, m.Name(), srcpkg, ifacename)

	var methodName string
	if n := []rune(m.Name()); n[0] == unicode.ToUpper(n[0]) {
		methodName = m.Name()
	} else {
		methodName = "Private_" + m.Name()
	}
	r.L(`func (mr *$0MockRecorder) $1($2) *$gomock.Call {`, replace, methodName, recorderArgs)

	for i, argname := range argnames {
		r.Imports().Deepequal().Ref("deepequal")

		if i > 0 {
			r.N()
		}
		r.L(`    if $0 != nil {`, argname)
		r.L(`        if _, ok := $0.($gomock.Matcher); !ok {`, argname)
		r.L(`            $0 = $deepequal.NewEqMatcher($0)`, argname)
		r.L(`        }`)
		r.L(`    }`)
	}

	r.Imports().Reflect().Ref("reflect")
	r.L(
		`    return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "$0", reflect.TypeOf((*$1Mock)(nil).$2), $3)`,
		m.Name(),
		replace,
		m.Name(),
		args,
	)

	r.L(`}`)

	if !whatever {
		return
	}

	r.L(`// $0Whatever records a call with arbitrary arguments.`, methodName)
	r.L(`func (mr *$0MockRecorder) $1Whatever() *$gomock.Call {`, replace, methodName)

	wargs := &gogh.Commas{}
	for i := 0; i < s.Params().Len(); i++ {
		wargs.Add(r.S("$gomock.Any()"))
	}

	r.L(`    return mr.$0($1)`, methodName, wargs)
	r.L(`}`)
}

func generateVariadicMethod(
	r *gogh.GoRenderer[*imports.Imports],
	iface *types.Interface,
	replace string,
	srcpkg *packages.Package,
	ifacename string,
	m *types.Func,
	s *types.Signature,
	whatever bool,
) {
	r.N()
	r.L(`// $0 method to implement $1.$2`, m.Name(), srcpkg.PkgPath, ifacename)

	var params gogh.Params
	var args gogh.Commas
	var last string

	for i := 0; i < s.Params().Len(); i++ {
		p := s.Params().At(i)
		t := p.Type()

		var name string
		switch p.Name() {
		case "_", "":
			name = "arg"
		default:
			name = p.Name()
		}
		argname := r.Uniq(name)
		var typePrefix string
		if i == s.Params().Len()-1 {
			typePrefix = "..."
			t = t.(*types.Slice).Elem()
			last = argname
		} else {
			args.Add(argname)
		}

		params.Add(argname, typePrefix+r.Type(t))
	}

	var returns gogh.Params
	var rets gogh.Commas
	var retnames []string
	for i := 0; i < s.Results().Len(); i++ {
		v := s.Results().At(i)

		base := v.Name()
		if base == "" || base == "_" {
			if v.Type().String() == "error" {
				base = "err"
			} else {
				base = "r"
			}
		}
		retname := r.Uniq(base)
		retnames = append(retnames, retname)
		returns.Add(retname, r.Type(v.Type()))
		rets.Add(retname)
	}

	vararg := r.Uniq("varargs")
	itemname := r.Uniq("item")

	r.L(`func (m *$0Mock) $1($2) ($3) {`, replace, m.Name(), params, returns)
	r.L(`    m.ctrl.T.Helper()`)
	r.L(`    $0 := []any{$1}`, vararg, args)
	r.L(`    for _, $0 := range $1 {`, itemname, last)
	r.L(`        $0 = append($0, $1)`, vararg, itemname)
	r.L(`    }`)
	if s.Results().Len() > 0 {
		r.L(`    ret := m.ctrl.Call(m, "$0", $1...)`, m.Name(), vararg)
	} else {
		r.L(`    m.ctlr.Call(m, "$0", $1...)`, m.Name(), vararg)
	}
	for i := 0; i < s.Results().Len(); i++ {
		r.L(`    $0, _ = ret[$1].($2)`, retnames[i], i, r.Type(s.Results().At(i).Type()))
	}
	r.L(`    return $0`, rets)
	r.L(`}`)
	r.N()

	recorderargs := args.String()
	if recorderargs != "" {
		recorderargs += " any,"
	}

	r.L(` // $0 register expected call of method $1.$2.$0`, m.Name(), srcpkg, ifacename)
	var methodName string
	if n := []rune(m.Name()); n[0] == unicode.ToUpper(n[0]) {
		methodName = m.Name()
	} else {
		methodName = "Private_" + m.Name()
	}

	r.Imports().Gomock().Ref("gomock")
	r.Imports().Deepequal().Ref("deepequal")
	r.Imports().Reflect().As("reflect")
	r.L(
		`func (mr *$0MockRecorder) $1($2 $3 ...any) *$gomock.Call {`,
		replace,
		methodName,
		recorderargs,
		last,
	)
	r.L(`    mr.mock.ctrl.T.Helper()`)
	r.L(`    $0 := append([]any{$1}, $2...)`, vararg, args, last)
	r.N()
	r.L(`    for i, v := range $0 {`, vararg)
	r.L(`        if _, ok := v.($gomock.Matcher); ok {`)
	r.L(`            continue`)
	r.L(`        }`)
	r.N()
	r.L(`        $[i] = $deepequal.NewEqMatcher(v)`, vararg)
	r.L(`    }`)
	r.L(
		`    return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "$0", reflect.TypeOf((*$1Mock)(nil).$2), $3...)`,
		m.Name(),
		replace,
		m.Name(),
		vararg,
	)
	r.L(`}`)

	if !whatever {
		return
	}

	r.L(`// $0Whatever registers a call with arbitrary positional and n variadic arguments`, methodName)
	r.L(`func (mr *$0MockRecorder) $1Whatever(n int) $gomock.Call {`, replace, methodName)
	r.L(`    var args []any`)
	r.L(`    for i := 0; i < n; i++ {`)
	r.L(`        args = append(args, $gomock.Any())`)
	r.L(`    }`)

	wargs := &gogh.Commas{}
	for i := 0; i < s.Params().Len()-1; i++ {
		wargs.Add(r.S("$gomock.Any()"))
	}
	r.L(`    return mr.$0($1, args...)`, methodName, wargs)
	r.L(`}`)
}
