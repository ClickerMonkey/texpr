package texpr

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

type TypeName string

func (tn TypeName) ToKey() string {
	return strings.ToLower(string(tn))
}

type Type struct {
	Name        TypeName
	Description string
	Values      []Value
	As          map[TypeName]string
	Enums       []string
	Parse       func(x string) (any, error)
}

type Value struct {
	Path        string
	Description string
	Type        TypeName
	Parameters  []Parameter
	Variadic    bool
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
	Type        TypeName
	Name        string
	Description string
	Default     *string
}

type Expr struct {
	Token     string
	Constant  bool
	Value     *Value
	Type      *Type
	Arguments []*Expr
	Next      *Expr
	Prev      *Expr
	Parent    *Expr
	Parameter *Parameter
}

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

type Options struct {
	RootType     TypeName
	ExpectedType TypeName
	Expression   string
	Types        []Type
}

var ErrNoTypes = errors.New("no types passed to parse")
var ErrNoExpression = errors.New("no expression given")
var ErrNoRootType = errors.New("no root type passed to parse")
var ErrInvalidRootType = errors.New("invalid root type passed to parse")

func Parse(opts Options) (*Expr, error) {
	if len(opts.Types) == 0 {
		return nil, ErrNoTypes
	}
	if len(opts.Expression) == 0 {
		return nil, ErrNoExpression
	}
	if len(opts.RootType) == 0 {
		return nil, ErrNoRootType
	}

	type typeEntry struct {
		Type   *Type
		Values map[string]*Value
	}

	typeMap := make(map[string]typeEntry)
	for i := range opts.Types {
		t := &opts.Types[i]
		k := t.Name.ToKey()
		entry := typeEntry{
			Type:   t,
			Values: make(map[string]*Value),
		}
		for j := range t.Values {
			v := &t.Values[j]
			entry.Values[strings.ToLower(v.Path)] = v
		}
		typeMap[k] = entry
	}

	root := typeMap[opts.RootType.ToKey()]
	if root.Type == nil {
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
			currentValue := parentType.Values[strings.ToLower(current.Token)]

			if currentValue != nil && !current.Constant {
				currentValueType := typeMap[currentValue.Type.ToKey()]
				if currentValueType.Type == nil {
					return fmt.Errorf("type %s on value %s.%s does not exist", currentValue.Type, current.Token, parentType.Type.Name)
				}

				current.Type = currentValueType.Type
				current.Value = currentValue

				args := current.Arguments
				argCount := len(args)
				argMin := currentValue.MinParameters()
				argMax := currentValue.MaxParameters()

				if argCount < argMin {
					return fmt.Errorf("%s.%s expects at least %d parameters", current.Token, parentType.Type.Name, currentValue.MinParameters())
				}
				if argCount > argMax {
					return fmt.Errorf("%s.%s expects no more than %d parameters", current.Token, parentType.Type.Name, currentValue.MaxParameters())
				}

				for i := 0; i < argCount; i++ {
					param := currentValue.Parameter(i)
					paramType := typeMap[param.Type.ToKey()]
					if paramType.Type == nil {
						return fmt.Errorf("parameter %s at %d has type %s which was not defined", param.Name, i, param.Type)
					}
					err := link(current.Arguments[i], paramType.Type)
					if err != nil {
						return err
					}
					current.Arguments[i].Parameter = param
				}

				for i := argCount; i < len(currentValue.Parameters); i++ {
					param := currentValue.Parameter(i)
					paramType := typeMap[param.Type.ToKey()]
					if paramType.Type == nil {
						return fmt.Errorf("parameter %s at %d has type %s which was not defined", param.Name, i, param.Type)
					}
					if param.Default == nil {
						return fmt.Errorf("parameter %s at %d was not given a value or a default value", param.Name, i)
					}
					arg := &Expr{
						Token:     *param.Default,
						Constant:  true,
						Type:      paramType.Type,
						Parameter: param,
						Parent:    current,
					}
					current.Arguments = append(current.Arguments, arg)
				}
			} else if current.Constant || currentValue == nil {
				if current.Next == nil {
					_, err := expectedType.Parse(current.Token)
					if err != nil {
						return fmt.Errorf("constant %s did not match expected type %s", current.Token, expectedType.Name)
					}
					current.Type = expectedType
					current.Constant = true
				} else {
					for _, entry := range typeMap {
						if entry.Type.Parse != nil {
							_, err := entry.Type.Parse(current.Token)
							if err != nil {
								current.Type = entry.Type
								current.Constant = true
								break
							}
						}
					}
				}

				if current.Type == nil {
					return fmt.Errorf("type could not be determined for %s", current.Token)
				}

				if len(current.Type.Enums) > 0 {
					matching := false
					for _, enumText := range current.Type.Enums {
						if strings.EqualFold(current.Token, enumText) {
							current.Token = enumText
							current.Constant = true
							matching = true
							break
						}
					}
					if !matching {
						return fmt.Errorf("%s does not match any expected enum values for %s", current.Token, current.Type.Name)
					}
				}
			} else {
				return fmt.Errorf("unexpected token %s", current.Token)
			}

			parent = current
			parentType = typeMap[current.Type.Name.ToKey()]
			current = current.Next
		}

		if parent != nil && expectedType != nil && parent.Type.Name != expectedType.Name {
			convert := parent.Type.As[expectedType.Name]
			if convert != "" {
				convertType := typeMap[parent.Type.Name.ToKey()]
				convertValue := convertType.Values[strings.ToLower(convert)]
				if convertValue != nil {
					next := &Expr{
						Token: convert,
						Type:  expectedType,
						Value: convertValue,
						Prev:  parent,
					}
					parent.Next = next
					parent = next
				}
			}
		}

		if expectedType != nil && parent != nil && parent.Type.Name != expectedType.Name {
			return fmt.Errorf("expected type %s but was given %s instead", expectedType.Name, parent.Type.Name)
		}

		return nil
	}

	var expectedType *Type
	if opts.ExpectedType != "" {
		expectedType = typeMap[opts.ExpectedType.ToKey()].Type
	}

	err := link(p.first, expectedType)
	if err != nil {
		return nil, err
	}

	return p.first, nil
}

type parser struct {
	parents []*Expr
	prev    *Expr
	first   *Expr
	e       string
	n       int
	i       int
}

func newParser(e string) parser {
	return parser{
		e: e,
		i: 0,
		n: len(e),
	}
}

func (p *parser) hasData() bool {
	return p.i < p.n
}

// Parses the expression at the current character. If the current character
// is the start of an expression the expression is returned. If the character
// represents a different part of an expression string then the internal state
// of the parser moves forward to parse an expression on the next call.
func (p *parser) parseExpr() (expr *Expr, err error) {
	b := p.e[p.i]
	switch b {
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
		expr, err = p.parseString(b)
	default:
		expr, err = p.parseToken()
	}

	if p.i == p.n && err == nil && len(p.parents) != 0 {
		err = fmt.Errorf("expression missing %d terminating parenthesis", len(p.parents))
	}

	return expr, err
}

func (p *parser) newExpr(e *Expr) *Expr {
	if p.first == nil {
		p.first = e
	}
	e.Prev = p.prev
	if p.prev != nil {
		p.prev.Next = e
	}
	p.prev = e
	if len(p.parents) > 0 && e.Prev == nil {
		parent := p.parents[len(p.parents)-1]
		parent.Arguments = append(parent.Arguments, e)
		e.Parent = parent
	}
	return e
}

func (p *parser) parseToken() (*Expr, error) {
	k := p.i
	out := strings.Builder{}
	b := p.e[k]
	word := wordChars[b]
	for k < p.n {
		b = p.e[k]
		if stopChars[b] || (word && !wordChars[b]) {
			break
		}
		out.WriteByte(b)
		k++
	}
	p.i = k
	return p.newExpr(&Expr{Token: out.String()}), nil
}

func (p *parser) parseString(end byte) (*Expr, error) {
	k := p.i
	out := strings.Builder{}
	escaped := false
	for k < p.n {
		k++
		b := p.e[k]
		if b == '/' && !escaped {
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
			p.i = k + 1
			return p.newExpr(&Expr{Token: out.String(), Constant: true}), nil
		}
		out.WriteByte(b)
		escaped = false
	}

	return nil, fmt.Errorf("string starting at %d did not have a terminating %s", p.i, string([]byte{end}))
}

var stopChars = charsToMap(".,()")

var wordChars = charsToMap("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_")

func charsToMap(x string) map[byte]bool {
	m := make(map[byte]bool, len(x))
	for i := range x {
		m[x[i]] = true
	}
	return m
}
