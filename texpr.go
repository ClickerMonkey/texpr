package texpr

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
)

// A name for a type.
type TypeName string

// A data type in an expression system. It can have values, with and without parameters.
// It can also be automatically cast to another type with the `As` field.
type Type struct {
	// The name of this type, should be unique.
	Name TypeName `json:"name"`
	// A description of this type.
	Description string `json:"description,omitempty"`
	// All values of this type.
	Values []Value `json:"values,omitempty"`
	// All types that this type can be converted to, and which value path can be used to do it.
	As map[TypeName]string `json:"as,omitempty"`
	// The type might be an enumerated value which means it has to be one of the specified values.
	// Parse can be specified to validate this and return a different data type other than string.
	Enums []string `json:"enums,omitempty"`
	// A custom parse function that converts a constant into a real value that is stored in Expression.Parsed.
	// If the given input does not match the type an error must be returned.
	Parse func(x string) (any, error) `json:"-"`
	// The parse order of the type. By default all types are considered equal and have an order of 0.
	// Higher parse orders are used first. For all types with the same parse order they are ordered
	// whether they have a Parse function (it prefers this). For two types with equivalent parse function
	// specificity they are ordered by type name length (preferring longer types before shorter).
	ParseOrder int `json:"parseOrder,omitempty"`

	values map[string]*Value
	as     map[TypeName]*Value
	enums  map[string]string
}

// Returns the value with the given path, case insensitive. If this type was not given
// to a system then a nil panic will occur.
func (t Type) Value(path string) *Value {
	return t.values[strings.ToLower(path)]
}

// Returns the value that's used to convert to the given type. If this type was not given
// to a system then a nil panic will occur.
func (t Type) AsValue(other TypeName) *Value {
	return t.as[other]
}

// Returns the enum value that matches the given text. If this type was not given
// to a system then a nil panic will occur.
func (t Type) EnumFor(input string) (string, bool) {
	value, ok := t.enums[strings.ToLower(input)]
	return value, ok
}

// Parses the constant input and returns a matching value. If there is no parse or matching
// enum option then an error is returned.
func (t Type) ParseInput(input string) (any, error) {
	if t.Parse == nil {
		value, exists := t.EnumFor(input)
		if exists {
			return value, nil
		}
		return nil, fmt.Errorf("parsing is not supported for %v", t.Name)
	}
	return t.Parse(input)
}

// A value (possibly with parameters) on a type.
type Value struct {
	// The main path for the value. Alternatives can be specified with Aliases.
	Path string `json:"path"`
	// The aliases to the path, to allow for more than one way to refer to the value.
	Aliases []string `json:"aliases,omitempty"`
	// The description of the value.
	Description string `json:"description,omitempty"`
	// The type of the value.
	Type TypeName `json:"type,omitempty"`
	// If the value is has a generic type that's determined based on one or more generic parameters.
	Generic bool `json:"generic,omitempty"`
	// The parameters for the value.
	Parameters []Parameter `json:"parameters,omitempty"`
	// If the last parameter can be specified any number of times.
	Variadic bool `json:"variadic,omitempty"`

	valueType *Type
}

// The calculated type of the value. This will only be non-nil when the value is passed to a system.
func (v Value) ValueType() *Type {
	return v.valueType
}

// Returns the maximum number of possible parameters. If this value is not parameterized
// this returns 0. If this value is parameterized and variadic it returns the largest possible int.
func (v Value) MaxParameters() int {
	if v.Variadic && len(v.Parameters) > 0 {
		return math.MaxInt
	}
	return len(v.Parameters)
}

// Returns the minimum number of required parameters
func (v Value) MinParameters() int {
	min := 0
	if v.Parameters != nil {
		for i, p := range v.Parameters {
			if p.Default != nil {
				min = i + 1
			}
		}
	}
	return min
}

// Returns the parameter at the given index. If this value is variadic and `i` goes
// beyond the defined parameters the last parameter is defined. If no parameter exists
// at the given index then nil is returned.
func (v Value) Parameter(i int) *Parameter {
	if i >= v.MaxParameters() || i < 0 {
		return nil
	}
	if i >= len(v.Parameters) {
		return &v.Parameters[len(v.Parameters)-1]
	}
	return &v.Parameters[i]
}

// Determines the type for this value for the given expression. If this value
// is generic the types of the generic parameters will be used to determine the returned type.
func (v Value) GetType(e *Expr) *Type {
	if !v.Generic {
		return v.valueType
	}
	genericTypes := make([]*Type, 0)
	if len(e.Arguments) > 0 {
		for _, arg := range e.Arguments {
			if arg.Type != nil && arg.Parameter.Generic {
				genericTypes = append(genericTypes, arg.Type)
			}
		}
	}
	return getBaseType(genericTypes)
}

// A parameter to a parameterized value. Type or Generic is required.
type Parameter struct {
	// The expected type for the parameter. Either this or Generic is required.
	Type TypeName `json:"type,omitempty"`
	// If there is no expected type, but this parameter and potentially others need to be the same type.
	// The generic parameters can also decide the type of generic value.
	Generic bool `json:"generic,omitempty"`
	// The name of the parameter.
	Name string `json:"name,omitempty"`
	// A more detailed description of the parameter.
	Description string `json:"description,omitempty"`
	// A default value, making this an optional parameter. This must be a valid value that can be parsed by the type.
	Default *string `json:"default,omitempty"`

	parameterType *Type
}

func (p Parameter) ParameterType() *Type {
	return p.parameterType
}

// The position of a character in a multi-line string.
type Position struct {
	// The index of the character in Options.Expression
	Index int
	// The line of the character, where \n delimits lines.
	Line int
	// The column of the character in its line.
	Column int
}

// The string representation of a position.
func (p Position) String() string {
	return fmt.Sprintf("(index: %d, line: %d, column: %d)", p.Index, p.Line, p.Column)
}

type Expr struct {
	// The string parsed from the expression input.
	Token string
	// The start position of the expression in the input.
	Start Position
	// The end position of the expression in the input.
	End Position
	// If this expression is a constant value and not a value.
	Constant bool
	// The parsed value if this expression is a constant.
	Parsed any
	// The value this expression is in the parent type.
	Value *Value
	// The parent type if any. If prev is nil this represents the root type.
	ParentType *Type
	// The type of this value/constant.
	Type *Type
	// The arguments to pass as the parameters to the value.
	Arguments []*Expr
	// The next expression in the chain on the result of this one.
	Next *Expr
	// The previous expression in the chain or nil. When nil this is either a constant
	// or a value on the root type.
	Prev *Expr
	// The expression this is an argument for, which is only set on the first expression in a chain.
	Parent *Expr
	// The parameter this expression is on if any. This is only set for the first expression in the chain.
	Parameter *Parameter
	// The system that created the expression.
	System *System
}

// Converts the expression to a string.
func (e Expr) String() string {
	out := strings.Builder{}
	c := &e
	for c != nil {
		if c.Prev != nil && wordChars[c.Token[0]] {
			out.WriteString(".")
		}
		if c.Constant {
			out.WriteString("'" + strings.ReplaceAll(c.Token, "'", "\\'") + "'")
		} else {
			out.WriteString(c.Token)
		}
		if len(c.Arguments) > 0 {
			out.WriteString("(")
			for i, arg := range c.Arguments {
				argSerialized := arg.String()
				if i > 0 {
					out.WriteString(",")
				}
				out.WriteString(argSerialized)
			}
			out.WriteString(")")
		}
		c = c.Next
	}

	return out.String()
}

// Returns the last expression in this chain.
func (e *Expr) Last() *Expr {
	c := e
	for c.Next != nil {
		c = c.Next
	}
	return c
}

// Returns a slice of all expressions in the chain starting with this expression.
func (e *Expr) Chain() []*Expr {
	chain := make([]*Expr, 0)
	c := e
	for c != nil {
		chain = append(chain, c)
		c = c.Next
	}
	return chain
}

// Returns if the type on this expression is one of the given types.
// If this expression is nil or has no type then this will return whether the given types are empty.
// Otherwise the type on the expression must match one of the given types.
func (e *Expr) TypeOneOf(types []*Type) bool {
	if e == nil || e.Type == nil {
		return len(types) == 0
	}
	for _, t := range types {
		if t.Name == e.Type.Name {
			return true
		}
	}
	return false
}

// An error occurred during the parsing or linking of System.Parse.
type ParseError struct {
	Message   string
	Expr      *Expr
	Parameter *Parameter
	Start     *Position
	End       *Position
}

var _ error = ParseError{}

// Creates a new parse error given the expression (if any) and the message.
func NewParseError(expr *Expr, message string) ParseError {
	e := ParseError{
		Message: message,
		Expr:    expr,
	}
	if expr != nil {
		e.Start = &expr.Start
		e.End = &expr.End
	}
	return e
}

// The parse error message.
func (e ParseError) Error() string {
	return e.Message
}

// An error occurred building a system from types.
type SystemError struct {
	Message   string
	Type      *Type
	Value     *Value
	Parameter *Parameter
	Path      *string
}

var _ error = SystemError{}

func (e SystemError) Error() string {
	return e.Message
}

// A type system that validates types, values, parameters, etc.
type System struct {
	types      []*Type
	typeMap    map[TypeName]*Type
	parseOrder []*Type
}

// Returns a System given a set of types and panics if any of the types, values, parameters, etc are malformed.
func NewSystemRequired(types []Type) System {
	sys, err := NewSystem(types)
	if err != nil {
		panic(err)
	}
	return sys
}

var pathValidator = regexp.MustCompile(`^([a-zA-Z0-9_]+|[^a-zA-Z0-9_,\.\(\)][^,\.\(\)]*)$`)

// Returns a new system and if any errors were found building the system.
func NewSystem(types []Type) (System, error) {
	sys := System{
		types:      make([]*Type, len(types)),
		typeMap:    make(map[TypeName]*Type),
		parseOrder: make([]*Type, 0, len(types)),
	}
	for i := range types {
		t := &types[i]
		t.values = make(map[string]*Value)
		t.as = make(map[TypeName]*Value)
		t.enums = make(map[string]string)

		if len(t.Values) > 0 {
			for k := range t.Values {
				v := &t.Values[k]
				if !pathValidator.MatchString(v.Path) {
					return sys, SystemError{
						Message: fmt.Sprintf("%s is not a valid path in %s", v.Path, t.Name),
						Type:    t,
					}
				}

				t.values[strings.ToLower(v.Path)] = v
				if len(v.Aliases) > 0 {
					for _, a := range v.Aliases {
						t.values[strings.ToLower(a)] = v
					}
				}

				if v.Generic == (v.Type != "") {
					return sys, SystemError{
						Message: fmt.Sprintf("value %s.%s must have either a type or generic but not both", t.Name, v.Path),
						Type:    t,
					}
				}
				if v.Generic {
					genericCount := 0
					if len(v.Parameters) > 0 {
						for _, param := range v.Parameters {
							if param.Generic {
								genericCount++
							}
						}
					}
					if genericCount == 0 {
						return sys, SystemError{
							Message: fmt.Sprintf("value %s.%s cannot have a generic type without one or more generic parameters.", t.Name, v.Path),
							Type:    t,
						}
					}
				}
			}
		}
		if len(t.As) > 0 {
			for typeName, valuePath := range t.As {
				value := t.Value(valuePath)
				if value == nil {
					return sys, SystemError{
						Message: fmt.Sprintf("%s as %s using value %s could not be found", t.Name, typeName, valuePath),
						Type:    t,
						Path:    &valuePath,
					}
				}
				t.as[typeName] = value
			}
		}
		if len(t.Enums) > 0 {
			for _, enumValue := range t.Enums {
				t.enums[strings.ToLower(enumValue)] = enumValue
			}
		}

		sys.types[i] = t
		sys.typeMap[t.Name] = t

		if t.Parse != nil || len(t.Enums) > 0 {
			sys.parseOrder = append(sys.parseOrder, t)
		}
	}

	for _, t := range sys.typeMap {
		for _, v := range t.values {
			v.valueType = sys.Type(v.Type)
			if v.valueType == nil && !v.Generic {
				return sys, SystemError{
					Message: fmt.Sprintf("type %s on %s.%s could not be found", v.Type, t.Name, v.Path),
					Value:   v,
				}
			}

			if len(v.Parameters) > 0 {
				for _, p := range v.Parameters {
					p.parameterType = sys.Type(p.Type)
					if p.parameterType == nil && !v.Generic {
						return sys, SystemError{
							Message:   fmt.Sprintf("type %s on %s.%s (parameter %s) could not be found", v.Type, t.Name, v.Path, p.Name),
							Value:     v,
							Type:      t,
							Parameter: &p,
						}
					}
				}
			}
		}
	}

	// Prefer types with parse logic, then enums. Sort by name length preferring longest.
	sort.Slice(sys.parseOrder, func(i, j int) bool {
		a := sys.parseOrder[i]
		b := sys.parseOrder[j]
		if a.ParseOrder != b.ParseOrder {
			return a.ParseOrder > b.ParseOrder
		}
		if (a.Parse != nil) != (b.Parse != nil) {
			return (a.Parse != nil)
		}
		return len(string(a.Name)) > len(string(b.Name))
	})

	return sys, nil
}

// Returns the type in the system with the given name, or nil if none exists.
func (s System) Type(name TypeName) *Type {
	return s.typeMap[name]
}

// Returns the types given to the system.
func (s System) Types() []*Type {
	return s.types
}

// Returns the types that can parse constants in the order determined by the system.
func (s System) ParseOrder() []*Type {
	return s.parseOrder
}

// The parse options for an expression string into an Expression struct.
type Options struct {
	// The type that is used as the root of the expressions.
	RootType TypeName
	// The expected types if any. If none are given then the type of the expression is returned.
	// If one or more are given then the expression is automatically cast to a desired type
	// if possible, otherwise errors if it can't meet the expected types.
	ExpectedTypes []TypeName
	// The expression to parse.
	Expression string
}

// No types are defined in the system.
var ErrNoTypes = NewParseError(nil, "undefined types")

// No expression was passed to the parse function.
var ErrNoExpression = NewParseError(nil, "undefined expression")

// No root type was specified in the options for parsing.
var ErrNoRoot = NewParseError(nil, "undefined root type")

// Parses an expression with the given set of options. Even if the expression is invalid it will be
// returned and all attempts of determining types and values will be made to best inform the user
// precisely what is wrong and what is valid.
func (sys System) Parse(opts Options) (*Expr, error) {
	if len(sys.Types()) == 0 {
		return nil, ErrNoTypes
	}
	if len(opts.Expression) == 0 {
		return nil, ErrNoExpression
	}
	if opts.RootType == "" {
		return nil, ErrNoRoot
	}

	root := sys.Type(opts.RootType)
	if root == nil {
		return nil, NewParseError(nil, fmt.Sprintf("undefined root type: %s", opts.RootType))
	}

	expectedTypes := make([]*Type, len(opts.ExpectedTypes))
	if len(opts.ExpectedTypes) >= 0 {
		for i, name := range opts.ExpectedTypes {
			expectedTypes[i] = sys.Type(name)
			if expectedTypes[i] == nil {
				return nil, NewParseError(nil, fmt.Sprintf("undefined expected type: %s", name))
			}
		}
	}

	err := error(nil)
	p := newParser(opts.Expression)

	for p.hasData() && err == nil {
		_, err = p.parseExpr()
	}

	// Always try to link the types, values, parameters, etc to expressions even if there was a parse error
	linkError := sys.link(p.first, expectedTypes, root)
	if err == nil {
		err = linkError
	}

	return p.first, err
}

func (sys System) link(e *Expr, expectedTypes []*Type, root *Type) error {
	current := e
	parentType := root
	var parent *Expr

	for current != nil {
		currentValue := parentType.Value(current.Token)

		current.ParentType = parentType

		// if it matches a value on the parent type and is not a constant
		if currentValue != nil && !current.Constant {
			current.Type = currentValue.ValueType()
			current.Value = currentValue

			err := sys.linkArguments(current, root)
			if err != nil {
				return err
			}

			// For generic values, calculate the type now that the argument types are determined.
			if currentValue.Generic {
				current.Type = currentValue.GetType(current)
				if current.Type == nil {
					return NewParseError(current, fmt.Sprintf("generic type could not be determined for %s", current.Token))
				}
				// Convert the generic arguments to the expected types
				for _, arg := range current.Arguments {
					if arg.Parameter.Generic {
						sys.convertToExpected(arg.Last(), []*Type{current.Type})
					}
				}
			}

			// if it is a constant or does not match a value on the parent type
		} else if current.Constant || currentValue == nil {
			// if its a lone constant and an expected type is given, parse using only that
			if current.Next == nil && len(expectedTypes) > 0 {
				err := sys.setConstant(current, expectedTypes, true)
				if err != nil {
					return err
				}
				// its not a lone constant or there is no expected type
			} else if current.Prev == nil {
				sys.setConstant(current, sys.parseOrder, false)
				if current.Type == nil {
					return NewParseError(current, fmt.Sprintf("type could not be determined for %s", current.Token))
				}
			} else {
				return NewParseError(current, fmt.Sprintf("invalid value %s", current.Token))
			}
		} else {
			return NewParseError(current, fmt.Sprintf("unexpected token %s", current.Token))
		}

		parent = current
		parentType = current.Type
		current = current.Next
	}

	// Try to auto-cast the last expression to an expected type in the order they were given.
	parent = sys.convertToExpected(parent, expectedTypes)

	// If the last expression does not match an expected type, error.
	if parent != nil && len(expectedTypes) > 0 && !parent.TypeOneOf(expectedTypes) {
		return NewParseError(parent, fmt.Sprintf("expected type(s) %s but was given %s instead", getTypeNames(expectedTypes), parent.Type.Name))
	}

	return nil
}

func (sys System) convertToExpected(last *Expr, expectedTypes []*Type) *Expr {
	if last == nil || len(expectedTypes) == 0 || last.TypeOneOf(expectedTypes) {
		return last
	}

	for _, expectedType := range expectedTypes {
		convert := last.Type.AsValue(expectedType.Name)
		if convert != nil {
			next := &Expr{
				Token:      convert.Path,
				Type:       expectedType,
				Value:      convert,
				Prev:       last,
				ParentType: last.Type,
			}
			last.Next = next
			last = next

			break
		}
	}

	return last
}

func (sys System) setConstant(current *Expr, tryTypes []*Type, required bool) error {
	for _, parser := range tryTypes {
		parsed, err := parser.ParseInput(current.Token)
		if err == nil {
			current.Type = parser
			current.Constant = true
			current.Parsed = parsed
			return nil
		}
	}

	if required {
		return NewParseError(current, fmt.Sprintf("constant %s did not match expected type(s) %s", current.Token, getTypeNames(tryTypes)))
	}

	return nil
}

func (sys System) linkArguments(current *Expr, root *Type) error {
	args := current.Arguments
	argCount := len(args)
	argMin := current.Value.MinParameters()
	argMax := current.Value.MaxParameters()

	if argCount < argMin {
		return NewParseError(current, fmt.Sprintf("%s.%s expects at least %d parameters", current.Token, current.ParentType.Name, argMin))
	}
	if argCount > argMax {
		return NewParseError(current, fmt.Sprintf("%s.%s expects no more than %d parameters", current.Token, current.ParentType.Name, argMax))
	}

	for i := 0; i < argCount; i++ {
		param := current.Value.Parameter(i)
		parameterType := make([]*Type, 0)
		if param.parameterType != nil {
			parameterType = append(parameterType, param.parameterType)
		}
		err := sys.link(current.Arguments[i], parameterType, root)
		if err != nil {
			return err
		}
		current.Arguments[i].Parameter = param
	}

	for i := argCount; i < len(current.Value.Parameters); i++ {
		param := current.Value.Parameter(i)
		if param.Default == nil {
			err := NewParseError(current, fmt.Sprintf("parameter %s at %d was not given a value or a default value", param.Name, i))
			err.Parameter = param
			return err
		}
		parsed, parseError := param.parameterType.ParseInput(*param.Default)
		if parseError != nil {
			err := NewParseError(current, parseError.Error())
			err.Parameter = param
			return err
		}
		arg := &Expr{
			Token:     *param.Default,
			Constant:  true,
			Type:      param.parameterType,
			Parameter: param,
			Parent:    current,
			Parsed:    parsed,
		}
		current.Arguments = append(current.Arguments, arg)
	}

	return nil
}

type parser struct {
	// the stack of parameterized expressions the prev expression is in.
	parents []*Expr
	// the previously parsed expression, or nil at the start of a new chain.
	prev *Expr
	// the first parsed expression in the input.
	first *Expr
	// the input
	e string
	// the cached length of the input
	n int
	// the current place in the parser
	i int
	// the position the last time line changed, used to calculate column.
	lineReset int
	// the current line
	line int
}

// Creates a new parser for the given expression.
func newParser(e string) parser {
	return parser{
		e: e,
		n: len(e),
	}
}

// If the parser still has expressions to parse.
func (p *parser) hasData() bool {
	return p.i < p.n
}

// The current position.
func (p parser) position() Position {
	return Position{
		Index:  p.i,
		Column: p.i - p.lineReset,
		Line:   p.line,
	}
}

// Parses the expression at the current character. If the current character
// is the start of an expression the expression is returned. If the character
// represents a different part of an expression string then the internal state
// of the parser moves forward to parse an expression on the next call.
func (p *parser) parseExpr() (expr *Expr, err error) {
	searching := p.i < p.n
	for searching {
		b := p.e[p.i]
		switch b {
		case '\n':
			p.i++
			p.line++
			p.lineReset = p.i
		case ' ', '\t', '\r', '\f', '\v':
			p.i++
		case '(':
			p.parents = append(p.parents, p.prev)
			p.prev = nil
			p.i++
		case ')':
			n := len(p.parents) - 1
			if n == -1 {
				return expr, NewParseError(expr, fmt.Sprintf("unexpected ) at %v", p.position()))
			}
			p.prev = p.parents[n]
			p.parents = p.parents[:n]
			p.i++
		case ',':
			p.prev = nil
			p.i++
		case '.':
			p.i++
		case '"', '\'':
			expr, err = p.parseConstant()
			searching = false
		default:
			expr, err = p.parseToken()
			searching = false
		}
		searching = searching && p.i < p.n
	}

	if p.i == p.n && err == nil && len(p.parents) != 0 {
		err = NewParseError(expr, fmt.Sprintf("expression missing %d terminating parenthesis", len(p.parents)))
	}

	// When an error has occurred and the previous character indicated we expect something
	// next add an empty expression to make that clear that nothing was given when something
	// was expected.
	if p.i > 0 && nextChars[p.e[p.i-1]] {
		expr = p.newExpr(&Expr{Start: p.position(), End: p.position()})
		if err == nil {
			err = NewParseError(expr, fmt.Sprintf("expression expecting a value but found nothing"))
		}
	}

	return expr, err
}

// Returns the expression but updates the Prev, Next, Arguments, and Parent of this expression
// and related expression.
func (p *parser) newExpr(e *Expr) *Expr {
	// The first expression is what Parse returns.
	if p.first == nil {
		p.first = e
	}
	// Keep track of the previous expression
	e.Prev = p.prev
	// Link up Prev's Next to this
	if p.prev != nil {
		p.prev.Next = e
	}
	// This is the new prev
	p.prev = e
	// If this is the first expresion in an argument, add it to the parent expressions
	// argument list and set parent.
	if len(p.parents) > 0 && e.Prev == nil {
		parent := p.parents[len(p.parents)-1]
		parent.Arguments = append(parent.Arguments, e)
		e.Parent = parent
	}
	return e
}

// Parses a token. A token is a value on type (parameterized and non-parameterized)
// or a constant not surrounded with quotes.
func (p *parser) parseToken() (*Expr, error) {
	out := strings.Builder{}
	b := p.e[p.i]
	word := wordChars[b]
	start := p.position()
	for p.i < p.n {
		b = p.e[p.i]
		if stopChars[b] || (word && !wordChars[b]) {
			break
		}
		out.WriteByte(b)
		p.i++
	}
	return p.newExpr(&Expr{Token: out.String(), Start: start, End: p.position()}), nil
}

// Parses a constant surrounded with quotes.
func (p *parser) parseConstant() (*Expr, error) {
	out := strings.Builder{}
	escaped := false
	end := p.e[p.i]
	start := p.position()
	for p.i < p.n {
		p.i++
		b := p.e[p.i]
		if b == '\\' && !escaped {
			escaped = true
			continue
		}
		if escaped {
			switch b {
			case 'n':
				b = '\n'
			case 'r':
				b = '\r'
			case 't':
				b = '\t'
			}
		}
		if b == end && !escaped {
			p.i++
			return p.newExpr(&Expr{Token: out.String(), Constant: true, Start: start, End: p.position()}), nil
		}
		out.WriteByte(b)
		escaped = false
	}

	return nil, NewParseError(nil, fmt.Sprintf("quoted constant starting at %v did not have a terminating %s", start, string([]byte{end})))
}

// Any chars that end a token.
var stopChars = charsToMap(".,()")

// Any chars that are valid ".name" values.
var wordChars = charsToMap("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_")

// Any chars where you would expect another expression to follow
var nextChars = charsToMap("(,.")

// converts every byte in the given string into a map where each byte given has a true value and any byte not found in the map will be false
func charsToMap(x string) map[byte]bool {
	m := make(map[byte]bool, len(x))
	for i := range x {
		m[x[i]] = true
	}
	return m
}

func getTypeNames(types []*Type) string {
	names := make([]string, len(types))
	for i, t := range types {
		names[i] = string(t.Name)
	}
	return strings.Join(names, ", ")
}

func getBaseType(types []*Type) *Type {
	if len(types) == 0 {
		return nil
	}
	if len(types) == 1 {
		return types[0]
	}
	allSame := true
	for i := 1; i < len(types); i++ {
		allSame = allSame && types[i] == types[0]
	}
	if allSame {
		return types[0]
	}
	for i := range types {
		a := types[i]
		allA := true
		for k := range types {
			b := types[k]
			allA = allA && ((a.Name == b.Name) || (b.as[a.Name] != nil))
		}
		if allA {
			return a
		}
	}
	return nil
}
