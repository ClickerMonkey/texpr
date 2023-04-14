package texpr

import (
	"fmt"
	"strings"
)

// A compiler is a function that is given an expression, the root type, a previously compiled expression (CE),
// argument CEs, and returns a CE for the given expression.
type Compiler[CE any] func(e *Expr, root *Type, previous CE, arguments []CE) (CE, error)

// A set of compilers mapped by their lowecased paths.
type ValueCompilers[CE any] map[string]Compiler[CE]

// A set of value compilers mapped by a type.
type TypeCompilers[CE any] map[TypeName]ValueCompilers[CE]

// Options to pass to a compile function.
type CompileOptions[CE any] struct {
	// The initial compiled expression value. This is passed to the compiler functions for the
	// first expressions in a chain.
	Initial CE
	// Compilers for each type and their values.
	TypeCompilers TypeCompilers[CE]
	// A compiler for a constant expression.
	ConstantCompiler Compiler[CE]
}

// Compiles the given expression into the desired compiled expression (CE). If there was any error
// or a type or value compiler was not specified an error will be returned.
func Compile[CE any](e *Expr, options CompileOptions[CE]) (CE, error) {
	current := e
	err := error(nil)
	last := options.Initial
	root := e.ParentType
	parentType := root

	for current != nil {
		if current.Constant {
			last, err = options.ConstantCompiler(current, root, last, nil)
			if err != nil {
				break
			}
		} else {
			typeCompiler := options.TypeCompilers[parentType.Name]
			if typeCompiler == nil {
				err = fmt.Errorf("no value compilers specified for %s", parentType.Name)
				break
			}
			valueCompiler := typeCompiler[strings.ToLower(current.Value.Path)]
			if valueCompiler == nil {
				err = fmt.Errorf("no value %s specified for %s", current.Value.Path, parentType.Name)
				break
			}
			args := make([]CE, len(current.Arguments))
			if len(args) > 0 {
				for i, arg := range current.Arguments {
					args[i], err = Compile(arg, options)
					if err != nil {
						break
					}
				}
				if err != nil {
					break
				}
			}
			last, err = valueCompiler(current, root, last, args)
			if err != nil {
				break
			}
		}
		parentType = current.Type
		current = current.Next
	}

	return last, err
}
