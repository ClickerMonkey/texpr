package texpr

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

type TypeName string

func (tn TypeName) ToKey() string {
	return strings.ToLower(string(tn))
}

type Type struct {
	Name        TypeName                    `json:"name"`
	Description string                      `json:"description,omitempty"`
	Values      []Value                     `json:"values,omitempty"`
	As          map[TypeName]string         `json:"as,omitempty"`
	Enums       []string                    `json:"enums,omitempty"`
	Parse       func(x string) (any, error) `json:"-"`
	ParseOrder  int                         `json:"parseOrder,omitempty"`

	values map[string]*Value
	as     map[TypeName]*Value
	enums  map[string]string
}

func (t Type) Value(path string) *Value {
	return t.values[strings.ToLower(path)]
}

func (t Type) AsValue(other TypeName) *Value {
	return t.as[other]
}

func (t Type) EnumFor(input string) (string, bool) {
	value, ok := t.enums[strings.ToLower(input)]
	return value, ok
}

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

type Value struct {
	Path        string      `json:"path"`
	Aliases     []string    `json:"aliases,omitempty"`
	Description string      `json:"description,omitempty"`
	Type        TypeName    `json:"type"`
	Parameters  []Parameter `json:"parameters,omitempty"`
	Variadic    bool        `json:"variadic,omitempty"`

	valueType *Type
}

func (v Value) ValueType() *Type {
	return v.valueType
}

func (v Value) MaxParameters() int {
	if v.Variadic {
		return math.MaxInt
	}
	return len(v.Parameters)
}

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

func (v Value) Parameter(i int) *Parameter {
	if i >= v.MaxParameters() || i < 0 {
		return nil
	}
	if i >= len(v.Parameters) {
		return &v.Parameters[len(v.Parameters)-1]
	}
	return &v.Parameters[i]
}

type Parameter struct {
	Type        TypeName `json:"type"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Default     *string  `json:"default,omitempty"`

	parameterType *Type
}

func (p Parameter) ParameterType() *Type {
	return p.parameterType
}

type Position struct {
	Index  int
	Line   int
	Column int
}

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
	// The parent type if any.
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

func (e *Expr) Last() *Expr {
	c := e
	for c.Next != nil {
		c = c.Next
	}
	return c
}

func (e *Expr) Chain() []*Expr {
	chain := make([]*Expr, 0)
	c := e
	for c != nil {
		chain = append(chain, c)
		c = c.Next
	}
	return chain
}

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

type ParseError struct {
	Message   string
	Expr      *Expr
	Parameter *Parameter
	Start     *Position
	End       *Position
}

var _ error = ParseError{}

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

func (e ParseError) Error() string {
	return e.Message
}

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

type System struct {
	types      []*Type
	typeMap    map[string]*Type
	parseOrder []*Type
}

func NewSystem(types []Type) (System, error) {
	sys := System{
		types:      make([]*Type, len(types)),
		typeMap:    make(map[string]*Type),
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
				t.values[strings.ToLower(v.Path)] = v
				if len(v.Aliases) > 0 {
					for _, a := range v.Aliases {
						t.values[strings.ToLower(a)] = v
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
		sys.typeMap[t.Name.ToKey()] = t

		if t.Parse != nil || len(t.Enums) > 0 {
			sys.parseOrder = append(sys.parseOrder, t)
		}
	}

	for _, t := range sys.typeMap {
		for _, v := range t.values {
			v.valueType = sys.Type(v.Type)
			if v.valueType == nil {
				return sys, SystemError{
					Message: fmt.Sprintf("type %s on %s.%s could not be found", v.Type, t.Name, v.Path),
					Value:   v,
				}
			}

			if len(v.Parameters) > 0 {
				for _, p := range v.Parameters {
					p.parameterType = sys.Type(p.Type)
					if p.parameterType == nil {
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

func (s System) Type(name TypeName) *Type {
	return s.typeMap[name.ToKey()]
}

func (s System) Types() []*Type {
	return s.types
}

func (s System) ParseOrder() []*Type {
	return s.parseOrder
}

type Options struct {
	RootType      TypeName
	ExpectedTypes []TypeName
	Expression    string
}

var ErrNoTypes = NewParseError(nil, "undefined types")
var ErrNoExpression = NewParseError(nil, "undefined expression")
var ErrNoRoot = NewParseError(nil, "undefined root type")

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
		case '.', ',':
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
