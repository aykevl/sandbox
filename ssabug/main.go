package main

import (
	"log"
	"os"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func main() {
	// Load the packages.
	// (Specifying all Need* constants here because I don't know which are
	// actually needed.)
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedTypes | packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedDeps}
	initial, err := packages.Load(cfg, "ssabug/testa", "ssabug/value")
	if err != nil {
		log.Fatal(err)
	}

	// Create SSA packages for all well-typed packages.
	_, pkgs := ssautil.Packages(initial, ssa.SanityCheckFunctions|ssa.BareInits|ssa.InstantiateGenerics)

	// Check for the bug.
	for _, pkg := range pkgs {
		if pkg.Pkg.Path() != "ssabug/testa" {
			continue
		}
		pkg.Build()
		readPackage(pkg)
	}
}

func readPackage(pkg *ssa.Package) {
	pkg.Build()
	test := pkg.Members["Test"].(*ssa.Function)
	println("Test:", test.RelString(nil))
	test.WriteTo(os.Stdout)
	mapCall := test.Blocks[0].Instrs[0].(*ssa.Call)
	println("map call:", mapCall.String())
	mapFn := mapCall.Common().Value.(*ssa.Function)
	println("map fn:", mapFn.RelString(nil))
	mapFn.WriteTo(os.Stdout)
	makeInst := mapFn.Blocks[0].Instrs[1].(*ssa.MakeInterface)
	println("make inst:", makeInst.String())
	ms := pkg.Prog.MethodSets.MethodSet(makeInst.X.Type())
	method := ms.At(0)
	methodFn := pkg.Prog.MethodValue(method)
	println("method:", methodFn.RelString(nil))
	methodFn.WriteTo(os.Stdout)
	invokeInst := methodFn.Blocks[0].Instrs[2].(*ssa.Call)
	invokeParam := invokeInst.Call.Args[0].(*ssa.Function)
	println("param:", invokeParam.String())

	// This call right here result in a panic:
	//     panic: runtime error: index out of range [0] with length 0
	invokeParam.Origin()
}
