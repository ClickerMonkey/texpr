package texpr

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIt(t *testing.T) {
	const (
		TypeInt         = TypeName("int")
		TypeText        = TypeName("text")
		TypeDate        = TypeName("date")
		TypeDuration    = TypeName("duration")
		TypeDateTime    = TypeName("dateTime")
		TypeTime        = TypeName("time")
		TypeDayOfWeek   = TypeName("dayOfWeek")
		TypeBool        = TypeName("bool")
		TypeUser        = TypeName("user")
		TypeContext     = TypeName("context")
		TypeTimePackage = TypeName("timePackage")
	)

	sys, err := NewSystem([]Type{{
		Name:        TypeDayOfWeek,
		Description: "A day of the week",
		Enums:       []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"},
		As: map[TypeName]string{
			TypeText: "text",
		},
		Values: []Value{
			{Path: "text", Type: TypeText},
			{Path: "=", Type: TypeBool, Parameters: []Parameter{
				{Name: "value", Type: TypeDayOfWeek},
			}},
			{Path: "!=", Type: TypeBool, Parameters: []Parameter{
				{Name: "value", Type: TypeDayOfWeek},
			}},
			{Path: "oneOf", Type: TypeBool, Variadic: true, Parameters: []Parameter{
				{Name: "values", Type: TypeDayOfWeek},
			}},
		},
		Parse: func(x string) (any, error) {
			k := strings.ToLower(x)
			switch k {
			case "sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday":
				return k, nil
			}
			return nil, fmt.Errorf("%s invalid day of week", x)
		},
	}, {
		Name:        TypeDuration,
		Description: "A duration of time",
		Enums:       []string{"year", "month", "week", "day", "hour", "minute", "second"},
		Parse: func(x string) (any, error) {
			k := strings.ToLower(x)
			switch k {
			case "year", "month", "week", "day", "hour", "minute", "second":
				return k, nil
			}
			return nil, fmt.Errorf("%s invalid duration", x)
		},
	}, {
		Name:        TypeDate,
		Description: "A date without time or zone",
		As: map[TypeName]string{
			TypeText: "text",
		},
		Values: []Value{
			{Path: "text", Type: TypeText},
			{Path: "year", Type: TypeInt},
			{Path: "month", Type: TypeInt},
			{Path: "dayOfMonth", Type: TypeInt},
			{Path: "dayOfWeek", Type: TypeDayOfWeek},
			{Path: "add", Type: TypeDate, Parameters: []Parameter{
				{Name: "amount", Type: TypeInt},
				{Name: "duration", Type: TypeDuration},
			}},
		},
		Parse: func(x string) (any, error) {
			return time.Parse(time.DateOnly, x)
		},
	}, {
		Name:        TypeDateTime,
		Description: "A date & time without zone",
		As: map[TypeName]string{
			TypeText: "text",
		},
		Values: []Value{
			{Path: "text", Type: TypeText},
			{Path: "date", Type: TypeDate},
			{Path: "time", Type: TypeTime},
			{Path: "hour", Type: TypeInt},
			{Path: "minute", Type: TypeInt},
			{Path: "second", Type: TypeInt},
			{Path: "year", Type: TypeInt},
			{Path: "month", Type: TypeInt},
			{Path: "dayOfMonth", Type: TypeInt},
			{Path: "dayOfWeek", Type: TypeDayOfWeek},
			{Path: "add", Type: TypeDate, Parameters: []Parameter{
				{Name: "amount", Type: TypeInt},
				{Name: "duration", Type: TypeDuration},
			}},
		},
		Parse: func(x string) (any, error) {
			return time.Parse(time.DateTime, x)
		},
	}, {
		Name:        TypeTime,
		Description: "A time without zone",
		As: map[TypeName]string{
			TypeText: "text",
		},
		Values: []Value{
			{Path: "text", Type: TypeText},
			{Path: "hour", Type: TypeInt},
			{Path: "minute", Type: TypeInt},
			{Path: "second", Type: TypeInt},
			{Path: "add", Type: TypeDate, Parameters: []Parameter{
				{Name: "amount", Type: TypeInt},
				{Name: "duration", Type: TypeDuration},
			}},
		},
		Parse: func(x string) (any, error) {
			return time.Parse(time.TimeOnly, x)
		},
	}, {
		Name:        TypeInt,
		Description: "A whole number",
		As: map[TypeName]string{
			TypeText: "text",
		},
		Values: []Value{
			{Path: "text", Type: TypeText},
			{Path: ">", Type: TypeBool, Parameters: []Parameter{
				{Name: "value", Type: TypeInt},
			}},
			{Path: ">=", Type: TypeBool, Parameters: []Parameter{
				{Name: "value", Type: TypeInt},
			}},
			{Path: "<", Type: TypeBool, Parameters: []Parameter{
				{Name: "value", Type: TypeInt},
			}},
			{Path: "<=", Type: TypeBool, Parameters: []Parameter{
				{Name: "value", Type: TypeInt},
			}},
			{Path: "=", Type: TypeBool, Parameters: []Parameter{
				{Name: "value", Type: TypeInt},
			}},
			{Path: "!=", Type: TypeBool, Parameters: []Parameter{
				{Name: "value", Type: TypeInt},
			}},
			{Path: "+", Type: TypeInt, Parameters: []Parameter{
				{Name: "value", Type: TypeInt},
			}},
			{Path: "-", Type: TypeInt, Parameters: []Parameter{
				{Name: "value", Type: TypeInt},
			}},
			{Path: "*", Type: TypeInt, Parameters: []Parameter{
				{Name: "value", Type: TypeInt},
			}},
			{Path: "/", Type: TypeInt, Parameters: []Parameter{
				{Name: "value", Type: TypeInt},
			}},
			{Path: "%", Type: TypeInt, Parameters: []Parameter{
				{Name: "value", Type: TypeInt},
			}},
		},
		Parse: func(x string) (any, error) {
			return strconv.ParseInt(x, 10, 64)
		},
	}, {
		Name:        TypeBool,
		Description: "A true/false value",
		Enums:       []string{"true", "false"},
		As: map[TypeName]string{
			TypeText: "text",
		},
		Values: []Value{
			{Path: "text", Type: TypeText},
			{Path: "not", Type: TypeBool},
			{Path: "and", Type: TypeBool, Variadic: true, Parameters: []Parameter{
				{Name: "values", Type: TypeBool},
			}},
			{Path: "or", Type: TypeBool, Variadic: true, Parameters: []Parameter{
				{Name: "values", Type: TypeBool},
			}},
		},
		Parse: func(x string) (any, error) {
			return strconv.ParseBool(x)
		},
	}, {
		Name:        TypeUser,
		Description: "A user of our system",
		As: map[TypeName]string{
			TypeText: "name",
		},
		Values: []Value{
			{Path: "name", Type: TypeText},
			{Path: "createDate", Type: TypeDateTime},
		},
	}, {
		Name:        TypeContext,
		Description: "The context for evaluating expressions.",
		Values: []Value{
			{Path: "time", Type: TypeTimePackage},
			{Path: "user", Type: TypeUser},
		},
	}, {
		Name:        TypeTimePackage,
		Description: "The package for date & time related expressions.",
		Values: []Value{
			{Path: "today", Type: TypeDate},
			{Path: "yesterday", Type: TypeDate},
			{Path: "tomorrow", Type: TypeDate},
			{Path: "now", Type: TypeDateTime},
			{Path: "current", Type: TypeTime},
			{Path: "sunday", Type: TypeDayOfWeek, Description: "An unambiguous way to refer to Sunday"},
			{Path: "monday", Type: TypeDayOfWeek, Description: "An unambiguous way to refer to Monday"},
			{Path: "tuesday", Type: TypeDayOfWeek, Description: "An unambiguous way to refer to Tuesday"},
			{Path: "wednesday", Type: TypeDayOfWeek, Description: "An unambiguous way to refer to Wednesday"},
			{Path: "thursday", Type: TypeDayOfWeek, Description: "An unambiguous way to refer to Thursday"},
			{Path: "friday", Type: TypeDayOfWeek, Description: "An unambiguous way to refer to Friday"},
			{Path: "saturday", Type: TypeDayOfWeek, Description: "An unambiguous way to refer to Saturday"},
		},
	}, {
		Name: TypeText,
		Values: []Value{
			{Path: "lower", Type: TypeText},
			{Path: "upper", Type: TypeText},
			{Path: "isLower", Type: TypeBool},
			{Path: "isUpper", Type: TypeBool},
			{Path: "+", Type: TypeText, Parameters: []Parameter{
				{Name: "value", Type: TypeText},
			}},
			{Path: "=", Type: TypeBool, Parameters: []Parameter{
				{Name: "value", Type: TypeText},
			}},
			{Path: "!=", Type: TypeBool, Parameters: []Parameter{
				{Name: "value", Type: TypeText},
			}},
			{Path: "contains", Type: TypeBool, Parameters: []Parameter{
				{Name: "value", Type: TypeText},
			}},
		},
		ParseOrder: -1,
		Parse: func(x string) (any, error) {
			return x, nil
		},
	}})

	if err != nil {
		panic(err)
	}

	type Run func(v any, root any) (any, error)
	type RunRoot func(root any) (any, error)
	type Compile func(e *Expr, rootType *Type) (Run, error)
	type Compilers map[TypeName]map[string]Compile

	var compilers Compilers

	compile := func(e *Expr, rootType *Type) (RunRoot, error) {
		current := e
		parent := rootType
		runners := make([]Run, 0)

		for current != nil {
			if current.Value != nil && current.Type != nil {
				cmp := compilers[parent.Name][current.Value.Path]
				if cmp == nil {
					return nil, fmt.Errorf("no compiler found: %s.%s", parent.Name, current.Value.Path)
				}
				run, err := cmp(current, rootType)
				if err != nil || run == nil {
					return nil, err
				}
				runners = append(runners, run)
			} else if current.Constant && current.Type != nil && current.Type.Parse != nil {
				parsed, err := current.Type.ParseInput(current.Token)
				if err != nil {
					return nil, err
				}
				runners = append(runners, func(v, root any) (any, error) {
					return parsed, nil
				})
			} else {
				return nil, fmt.Errorf("unexpected token to compile: %s.%s", parent.Name, current.Token)
			}

			parent = current.Type
			current = current.Next
		}

		return func(root any) (any, error) {
			c := root
			var err error
			for _, run := range runners {
				c, err = run(c, root)
				if err != nil {
					return nil, err
				}
			}
			return c, nil
		}, nil
	}

	mapKeyCompile := func(key string) Compile {
		return func(e *Expr, rootType *Type) (Run, error) {
			return func(v, root any) (any, error) {
				return v.(map[string]any)[key], nil
			}, nil
		}
	}

	compileArguments := func(e *Expr, rootType *Type, withArgs func(args []RunRoot) Run) (Run, error) {
		args := make([]RunRoot, len(e.Arguments))
		var err error
		for i, arg := range e.Arguments {
			args[i], err = compile(arg, rootType)
			if err != nil {
				return nil, err
			}
		}
		return withArgs(args), nil
	}

	compilers = Compilers{
		TypeDateTime: map[string]Compile{
			"hour": func(e *Expr, rootType *Type) (Run, error) {
				return func(v any, root any) (any, error) {
					return int64(v.(time.Time).Hour()), nil
				}, nil
			},
		},
		TypeInt: map[string]Compile{
			">": func(e *Expr, rootType *Type) (Run, error) {
				return compileArguments(e, rootType, func(args []RunRoot) Run {
					return func(v any, root any) (any, error) {
						o, err := args[0](root)
						if err != nil {
							return false, err
						}
						return v.(int64) > o.(int64), nil
					}
				})
			},
		},
		TypeBool: map[string]Compile{
			"and": func(e *Expr, rootType *Type) (Run, error) {
				return compileArguments(e, rootType, func(args []RunRoot) Run {
					return func(v, root any) (any, error) {
						all := true
						for _, condition := range args {
							conditionValue, err := condition(root)
							if err != nil {
								return false, err
							}
							all = all && conditionValue.(bool)
						}
						return all, nil
					}
				})
			},
			"text": func(e *Expr, rootType *Type) (Run, error) {
				return func(v, root any) (any, error) {
					if b, ok := v.(bool); ok && b {
						return "true", nil
					}
					return "false", nil
				}, nil
			},
		},
		TypeText: map[string]Compile{
			"contains": func(e *Expr, rootType *Type) (Run, error) {
				return compileArguments(e, rootType, func(args []RunRoot) Run {
					return func(v, root any) (any, error) {
						o, err := args[0](root)
						if err != nil {
							return false, err
						}
						return strings.Contains(v.(string), o.(string)), nil
					}
				})
			},
			"lower": func(e *Expr, rootType *Type) (Run, error) {
				return func(v, root any) (any, error) {
					return strings.ToLower(v.(string)), nil
				}, nil
			},
		},
		TypeUser: map[string]Compile{
			"name": mapKeyCompile("name"),
		},
		TypeContext: map[string]Compile{
			"time": mapKeyCompile("time"),
			"user": mapKeyCompile("user"),
		},
		TypeTimePackage: map[string]Compile{
			"now":    mapKeyCompile("now"),
			"sunday": mapKeyCompile("sunday"),
		},
	}

	tests := []struct {
		name           string
		options        Options
		expectedString string
		input          any
		expectedValue  any
		expectedType   TypeName
		expectedError  string
		expectedCheck  func(*Expr, *testing.T)
	}{{
		name: "complex",
		options: Options{
			RootType:      TypeContext,
			ExpectedTypes: []TypeName{TypeBool},
			Expression:    "time.now.hour>(12).and(user.name.contains('Ma'))",
		},
		expectedString: "time.now.hour>('12').and(user.name.contains('Ma'))",
		input: map[string]any{
			"time": map[string]any{
				"now": time.Date(2023, 4, 11, 13, 0, 0, 0, time.Local),
			},
			"user": map[string]any{
				"name": "Mason",
			},
		},
		expectedValue: true,
	}, {
		name: "user.name.lower",
		options: Options{
			RootType:   TypeContext,
			Expression: "user.name.lower",
		},
		expectedString: "user.name.lower",
		input: map[string]any{
			"user": map[string]any{
				"name": "Mason",
			},
		},
		expectedValue: "mason",
	}, {
		name: "time.now.hour",
		options: Options{
			RootType:   TypeContext,
			Expression: "time.now.hour",
		},
		expectedString: "time.now.hour",
		input: map[string]any{
			"time": map[string]any{
				"now": time.Date(2023, 4, 11, 13, 0, 0, 0, time.Local),
			},
		},
		expectedValue: int64(13),
	}, {
		name: "sunday",
		options: Options{
			RootType:   TypeContext,
			Expression: "sunday",
		},
		expectedString: "'sunday'",
		input:          map[string]any{},
		expectedValue:  "sunday",
		expectedType:   TypeDayOfWeek,
	}, {
		name: "time.sunday",
		options: Options{
			RootType:   TypeContext,
			Expression: "time.Sunday",
		},
		expectedString: "time.Sunday",
		input: map[string]any{
			"time": map[string]any{
				"sunday": "sunday",
			},
		},
		expectedValue: "sunday",
		expectedType:  TypeDayOfWeek,
	}, {
		name: "time.sunday with tab",
		options: Options{
			RootType:   TypeContext,
			Expression: "\n\ttime.\tSunday\n",
		},
		expectedString: "time.Sunday",
		input: map[string]any{
			"time": map[string]any{
				"sunday": "sunday",
			},
		},
		expectedValue: "sunday",
		expectedType:  TypeDayOfWeek,
	}, {
		name: "time.sun error",
		options: Options{
			RootType:   TypeContext,
			Expression: "time.sun",
		},
		expectedError: "invalid value sun",
	}, {
		name: "time.sun error check",
		options: Options{
			RootType:   TypeContext,
			Expression: "time.sun",
		},
		expectedCheck: func(e *Expr, t *testing.T) {
			assert.Equal(t, e.Token, "time")
			assert.NotNil(t, e.Type)
			assert.Equal(t, e.Type.Name, TypeTimePackage)
			assert.NotNil(t, e.Next)
			assert.Equal(t, e.Next.Token, "sun")
			assert.Nil(t, e.Next.Type)
			assert.Nil(t, e.Next.Value)
			assert.Nil(t, e.Next.Parsed)
		},
		expectedError: "invalid value sun",
	}, {
		name: "time. error check",
		options: Options{
			RootType:   TypeContext,
			Expression: "time.",
		},
		expectedCheck: func(e *Expr, t *testing.T) {
			assert.Equal(t, e.Token, "time")
			assert.NotNil(t, e.Type)
			assert.Equal(t, e.Type.Name, TypeTimePackage)
			assert.NotNil(t, e.Next)
			assert.Equal(t, e.Next.Token, "")
			assert.Nil(t, e.Next.Type)
			assert.Nil(t, e.Next.Value)
			assert.Nil(t, e.Next.Parsed)
		},
		expectedError: "expression expecting a value but found nothing",
	}, {
		name: "bool detect",
		options: Options{
			RootType:   TypeContext,
			Expression: "true",
		},
		expectedType:  TypeBool,
		expectedValue: true,
	}, {
		name: "bool detect convert to text",
		options: Options{
			RootType:      TypeContext,
			Expression:    "true",
			ExpectedTypes: []TypeName{TypeText},
		},
		expectedType:  TypeText,
		expectedValue: "true",
	}, {
		name: "bool detect expect many stay same",
		options: Options{
			RootType:      TypeContext,
			Expression:    "true",
			ExpectedTypes: []TypeName{TypeBool, TypeText},
		},
		expectedType:  TypeBool,
		expectedValue: true,
	}, {
		name: "bool detect expect many convert to text",
		options: Options{
			RootType:      TypeContext,
			Expression:    "true",
			ExpectedTypes: []TypeName{TypeInt, TypeText},
		},
		expectedType:  TypeText,
		expectedValue: "true",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := sys.Parse(test.options)

			if test.expectedCheck != nil {
				test.expectedCheck(expr, t)
			}

			if err != nil {
				if test.expectedError != "" {
					assert.Equal(t, test.expectedError, err.Error())
					return
				} else {
					t.Fatalf("unexpected parse error: %v", err)
				}
			}

			if test.expectedType != "" {
				assert.Equal(t, test.expectedType, expr.Last().Type.Name)
			}

			if test.expectedString != "" {
				assert.Equal(t, test.expectedString, expr.String())
			}

			rootType := sys.Type(test.options.RootType)

			compiled, err := compile(expr, rootType)
			if err != nil {
				if test.expectedError != "" {
					assert.Equal(t, test.expectedError, err.Error())
					return
				} else {
					t.Fatalf("compilation error: %v", err)
				}
			}

			result, err := compiled(test.input)
			if err != nil {
				if test.expectedError != "" {
					assert.Equal(t, test.expectedError, err.Error())
					return
				} else {
					t.Fatalf("execution error: %v", err)
				}
			}

			assert.Equal(t, test.expectedValue, result)

			if test.expectedError != "" {
				t.Fatalf("expected error but none found: %v", test.expectedError)
			}
		})
	}
}
