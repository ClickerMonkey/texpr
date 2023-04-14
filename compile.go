package texpr

import (
	"fmt"
	"strings"
)

// A compiler is a function that is given an expression, the root type, a previously compiled expression (CE),
// argument CEs, and returns a CE for the given expression.
type Compiler[CE any] func(e *Expr, root *Type, previous CE, arguments []CE) (CE, error)

// A helper to the compile function.
type CompileSource[CE any] interface {
	// Returns the initial compiled expression value. This is passed to the compiler functions for the
	// first expressions in a chain.
	GetInitial(e *Expr) (CE, error)
	// Returns a compiled value for a constant expression.
	GetConstantCompiled(e *Expr, root *Type, previous CE, arguments []CE) (CE, error)
	// Returns a compiler for a value expression.
	GetValueCompiler(e *Expr, root *Type, previous CE) (Compiler[CE], error)
}

// A set of compilers mapped by their lowecased paths.
type ValueCompilers[CE any] map[string]Compiler[CE]

// A set of value compilers mapped by a type.
type TypeCompilers[CE any] map[TypeName]ValueCompilers[CE]

// A CompileSource implementation where compilers are looked up based on type->value.
type CompileSourceLookup[CE any] struct {
	// The initial compiled expression value. This is passed to the compiler functions for the
	// first expressions in a chain.
	Initial CE
	// Compilers for each type and their values.
	TypeCompilers TypeCompilers[CE]
	// A compiler for a constant expression.
	ConstantCompiler Compiler[CE]
}

var _ CompileSource[int] = CompileSourceLookup[int]{}

func (csl CompileSourceLookup[CE]) GetInitial(e *Expr) (CE, error) {
	return csl.Initial, nil
}
func (csl CompileSourceLookup[CE]) GetConstantCompiled(e *Expr, root *Type, previous CE, arguments []CE) (CE, error) {
	return csl.ConstantCompiler(e, root, previous, arguments)
}
func (csl CompileSourceLookup[CE]) GetValueCompiler(e *Expr, root *Type, previous CE) (Compiler[CE], error) {
	parent := e.ParentType
	if e.Prev != nil {
		parent = e.Prev.Type
	}
	typeCompiler := csl.TypeCompilers[parent.Name]
	if typeCompiler == nil {
		return nil, fmt.Errorf("no value compilers specified for %s", parent.Name)
	}
	valueCompiler := typeCompiler[strings.ToLower(e.Value.Path)]
	if valueCompiler == nil {
		return nil, fmt.Errorf("no value %s specified for %s", e.Value.Path, parent.Name)
	}
	return valueCompiler, nil
}

// Compiles the given expression into the desired compiled expression (CE). If there was any error
// or a type or value compiler was not specified an error will be returned.
func Compile[CE any](e *Expr, source CompileSource[CE]) (CE, error) {
	last, err := source.GetInitial(e)
	if err != nil {
		return last, err
	}

	current := e
	root := e.ParentType

	for current != nil {
		if current.Constant {
			last, err = source.GetConstantCompiled(current, root, last, nil)
			if err != nil {
				break
			}
		} else {
			valueCompiler, valueErr := source.GetValueCompiler(current, root, last)
			if valueErr != nil {
				err = valueErr
				break
			}

			args := make([]CE, len(current.Arguments))
			if len(args) > 0 {
				for i, arg := range current.Arguments {
					args[i], err = Compile(arg, source)
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
		current = current.Next
	}

	return last, err
}
