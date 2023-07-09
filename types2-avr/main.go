package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
)

func main() {
	// Parse a single source file.
	const input = `
package main
import "unsafe"

var Foo = unsafe.Pointer(uintptr(0x1234))
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "celsius.go", input, 0)
	if err != nil {
		log.Fatal(err)
	}

	// Type-check a package consisting of this file.
	// Type information for the imported packages
	// comes from $GOROOT/pkg/$GOOS_$GOOARCH/fmt.a.
	conf := types.Config{
		Importer: importer.Default(),
		Sizes: &stdSizes{
			IntSize:  4,
			PtrSize:  2,
			MaxAlign: 1,
		},
	}
	_, err = conf.Check("main", fset, []*ast.File{f}, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("parse OK")
}
