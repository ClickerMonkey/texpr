package texpr

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestIt(t *testing.T) {
	const (
		TypeInt       = TypeName("int")
		TypeText      = TypeName("text")
		TypeDate      = TypeName("date")
		TypeDuration  = TypeName("duration")
		TypeDateTime  = TypeName("dateTime")
		TypeDayOfWeek = TypeName("dayOfWeek")
		TypeBool      = TypeName("bool")
		TypeUser      = TypeName("user")
		TypeContext   = TypeName("context")
	)

	types := []Type{{
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
			{Path: "today", Type: TypeDate},
			{Path: "now", Type: TypeDateTime},
			{Path: "user", Type: TypeUser},
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
		Parse: func(x string) (any, error) {
			return x, nil
		},
	}}

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
				parsed, err := current.Type.Parse(current.Token)
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
				other, err := compile(e.Arguments[0], rootType)
				if err != nil {
					return nil, err
				}

				return func(v any, root any) (any, error) {
					o, err := other(root)
					if err != nil {
						return false, err
					}
					return v.(int64) > o.(int64), nil
				}, nil
			},
		},
		TypeBool: map[string]Compile{
			"and": func(e *Expr, rootType *Type) (Run, error) {
				conditions := make([]RunRoot, len(e.Arguments))
				var err error
				for i := range e.Arguments {
					conditions[i], err = compile(e.Arguments[i], rootType)
					if err != nil {
						return nil, err
					}
				}

				return func(v, root any) (any, error) {
					all := true
					for _, condition := range conditions {
						conditionValue, err := condition(root)
						if err != nil {
							return false, err
						}
						all = all && conditionValue.(bool)
					}
					return all, nil
				}, nil
			},
		},
		TypeText: map[string]Compile{
			"contains": func(e *Expr, rootType *Type) (Run, error) {
				other, err := compile(e.Arguments[0], rootType)
				if err != nil {
					return nil, err
				}

				return func(v, root any) (any, error) {
					o, err := other(root)
					if err != nil {
						return false, err
					}
					return strings.Contains(v.(string), o.(string)), nil
				}, nil
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
			"now":  mapKeyCompile("now"),
			"user": mapKeyCompile("user"),
		},
	}

	tests := []struct {
		name           string
		options        Options
		expectedString string
		input          any
		expectedValue  any
	}{{
		name: "complex",
		options: Options{
			RootType:     TypeContext,
			ExpectedType: TypeBool,
			Types:        types,
			Expression:   "now.hour>(12).and(user.name.contains('Ma'))",
		},
		expectedString: "now.hour>('12').and(user.name.contains('Ma'))",
		input: map[string]any{
			"now": time.Date(2023, 4, 11, 13, 0, 0, 0, time.Local),
			"user": map[string]any{
				"name": "Mason",
			},
		},
		expectedValue: true,
	}, {
		name: "user.name.lower",
		options: Options{
			RootType:   TypeContext,
			Types:      types,
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
		name: "now.hour",
		options: Options{
			RootType:   TypeContext,
			Types:      types,
			Expression: "now.hour",
		},
		expectedString: "now.hour",
		input: map[string]any{
			"now": time.Date(2023, 4, 11, 13, 0, 0, 0, time.Local),
		},
		expectedValue: 13,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := Parse(test.options)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
				return
			}

			serial := expr.String()

			if serial != test.expectedString {
				t.Fatalf("serialized expression did not match expected: %s", serial)
			}

			var rootType *Type
			for _, t := range test.options.Types {
				if t.Name.ToKey() == test.options.RootType.ToKey() {
					rootType = &t
					break
				}
			}

			compiled, err := compile(expr, rootType)
			if err != nil {
				t.Fatalf("compilation error: %v", err)
			}

			result, err := compiled(test.input)
			if err != nil {
				t.Fatalf("execution error: %v", err)
			}

			actual := fmt.Sprintf("%+v", result)
			expected := fmt.Sprintf("%+v", test.expectedValue)

			if actual != expected {
				t.Fatalf("Fail, actual: %s, expected: %s", actual, expected)
			}
		})
	}
}
