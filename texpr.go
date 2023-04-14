package texpr

import (
	"errors"
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
		return nil, ErrNoParse
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
					return sys, fmt.Errorf("%s as %s using value %s could not be found", t.Name, typeName, valuePath)
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
				return sys, fmt.Errorf("type %s on %s.%s could not be found", v.Type, t.Name, v.Path)
			}

			if len(v.Parameters) > 0 {
				for _, p := range v.Parameters {
					p.parameterType = sys.Type(p.Type)
					if p.parameterType == nil {
						return sys, fmt.Errorf("type %s on %s.%s (parameter %s) could not be found", v.Type, t.Name, v.Path, p.Name)
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
	RootType     TypeName
	ExpectedType TypeName
	Expression   string
}

var ErrNoParse = errors.New("no parser exists for type")
var ErrNoTypes = errors.New("no types passed to parse")
var ErrNoExpression = errors.New("no expression given")
var ErrNoRootType = errors.New("no root type passed to parse")
var ErrInvalidRootType = errors.New("invalid root type passed to parse")

func (sys System) Parse(opts Options) (*Expr, error) {
	if len(sys.Types()) == 0 {
		return nil, ErrNoTypes
	}
	if len(opts.Expression) == 0 {
		return nil, ErrNoExpression
	}
	if len(opts.RootType) == 0 {
		return nil, ErrNoRootType
	}

	root := sys.Type(opts.RootType)
	if root == nil {
		return nil, ErrInvalidRootType
	}

	p := newParser(opts.Expression)
	for p.hasData() {
		_, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
	}

	var link func(e *Expr, expectedType *Type) error

	link = func(e *Expr, expectedType *Type) error {
		current := e
		parentType := root
		var parent *Expr

		for current != nil {
			currentValue := parentType.Value(current.Token)

			current.ParentType = parentType

			if currentValue != nil && !current.Constant {
				current.Type = currentValue.ValueType()
				current.Value = currentValue

				args := current.Arguments
				argCount := len(args)
				argMin := currentValue.MinParameters()
				argMax := currentValue.MaxParameters()

				if argCount < argMin {
					return fmt.Errorf("%s.%s expects at least %d parameters", current.Token, parentType.Name, currentValue.MinParameters())
				}
				if argCount > argMax {
					return fmt.Errorf("%s.%s expects no more than %d parameters", current.Token, parentType.Name, currentValue.MaxParameters())
				}

				for i := 0; i < argCount; i++ {
					param := currentValue.Parameter(i)
					err := link(current.Arguments[i], param.parameterType)
					if err != nil {
						return err
					}
					current.Arguments[i].Parameter = param
				}

				for i := argCount; i < len(currentValue.Parameters); i++ {
					param := currentValue.Parameter(i)
					if param.Default == nil {
						return fmt.Errorf("parameter %s at %d was not given a value or a default value", param.Name, i)
					}
					arg := &Expr{
						Token:     *param.Default,
						Constant:  true,
						Type:      param.parameterType,
						Parameter: param,
						Parent:    current,
					}
					current.Arguments = append(current.Arguments, arg)
				}
			} else if current.Constant || currentValue == nil {
				if current.Next == nil && expectedType != nil {
					parsed, err := expectedType.Parse(current.Token)
					if err != nil {
						return fmt.Errorf("constant %s did not match expected type %s", current.Token, expectedType.Name)
					}
					current.Type = expectedType
					current.Constant = true
					current.Parsed = parsed
				} else {
					for _, parser := range sys.ParseOrder() {
						parsed, err := parser.Parse(current.Token)
						if err == nil {
							current.Type = parser
							current.Constant = true
							current.Parsed = parsed
							break
						}
					}
				}

				if current.Type == nil {
					return fmt.Errorf("type could not be determined for %s", current.Token)
				}
			} else {
				return fmt.Errorf("unexpected token %s", current.Token)
			}

			parent = current
			parentType = current.Type
			current = current.Next
		}

		if parent != nil && expectedType != nil && parent.Type.Name != expectedType.Name {
			convert := parent.Type.AsValue(expectedType.Name)
			if convert != nil {
				next := &Expr{
					Token:      convert.Path,
					Type:       expectedType,
					Value:      convert,
					Prev:       parent,
					ParentType: parent.Type,
				}
				parent.Next = next
				parent = next
			}
		}

		if expectedType != nil && parent != nil && parent.Type.Name != expectedType.Name {
			return fmt.Errorf("expected type %s but was given %s instead", expectedType.Name, parent.Type.Name)
		}

		return nil
	}

	var expectedType *Type
	if opts.ExpectedType != "" {
		expectedType = sys.Type(opts.ExpectedType)
	}

	err := link(p.first, expectedType)
	if err != nil {
		return nil, err
	}

	return p.first, nil
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
	// the current column in the line
	lineReset int
	// the current line
	line int
}

// Creates a new parser for the given expression.
func newParser(e string) parser {
	return parser{
		e: e,
		i: 0,
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
		err = fmt.Errorf("expression missing %d terminating parenthesis", len(p.parents))
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

	return nil, fmt.Errorf("quoted constant starting at %d did not have a terminating %s", p.i, string([]byte{end}))
}

// Any chars that end a token.
var stopChars = charsToMap(".,()")

// Any chars that are valid ".name" values.
var wordChars = charsToMap("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_")

func charsToMap(x string) map[byte]bool {
	m := make(map[byte]bool, len(x))
	for i := range x {
		m[x[i]] = true
	}
	return m
}
