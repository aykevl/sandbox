package main

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func main() {
	cfg := &packages.Config{
		Mode: packages.NeedImports | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedTypes,
	}

	pkgs, err := packages.Load(cfg, "./main")
	if err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		os.Exit(1)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	// Print the names of the source files
	// for each package listed on the command line.
	allPkgs := map[string]*packages.Package{}
	packages.Visit(pkgs, func(pkg *packages.Package) bool {
		allPkgs[pkg.Types.Path()] = pkg
		return true
	}, nil)

	_, ssaPkgs := ssautil.AllPackages([]*packages.Package{
		allPkgs["ssa-generics-race/testa"],
		allPkgs["ssa-generics-race/testb"],
		allPkgs["ssa-generics-race/value"],
	}, ssa.BareInits|ssa.InstantiateGenerics)
	testa := ssaPkgs[0]
	testb := ssaPkgs[1]
	value := ssaPkgs[2]

	// Build the common dependency before all others.
	value.Build()

	// Build testa and testb in parallel.
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		testa.Build()
		wg.Done()
	}()
	go func() {
		// Build testb while testa is also building.
		// This in itself does not seem to be a problem.
		testb.Build()

		// The testb package is now finished building.
		// So *presumably* we could just use it now, right?

		// To trigger the race condition, we're going to read from the
		// value.New[int] function in the Test function, which looks like this:
		//
		//    func Test() {
		//        value.New(1)
		//    }
		fn := testb.Members["Test"].(*ssa.Function)
		for _, inst := range fn.Blocks[0].Instrs {
			// Look for the call instruction.
			if inst, ok := inst.(*ssa.Call); ok {
				// Found the call instruction. Get the callee (value.New[int]).
				callee := inst.Common().StaticCallee()

				// Read from the callee.
				// RACE: this triggers the race condition.
				_ = callee.Blocks
			}
		}

		wg.Done()
	}()

	// Wait until the goroutines are done.
	wg.Wait()

	fmt.Println("done")
}
