package texpr

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Run func(root any) (any, error)

const (
	typeInt         = TypeName("int")
	typeText        = TypeName("text")
	typeDate        = TypeName("date")
	typeDuration    = TypeName("duration")
	typeDateTime    = TypeName("dateTime")
	typeTime        = TypeName("time")
	typeDayOfWeek   = TypeName("dayOfWeek")
	typeBool        = TypeName("bool")
	typeUser        = TypeName("user")
	typeContext     = TypeName("context")
	typeTimePackage = TypeName("timePackage")
)

var sys = NewSystemRequired([]Type{{
	Name:        typeDayOfWeek,
	Description: "A day of the week",
	Enums:       []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"},
	As: map[TypeName]string{
		typeText: "text",
	},
	Values: []Value{
		{Path: "text", Type: typeText},
		{Path: "=", Type: typeBool, Parameters: []Parameter{
			{Name: "value", Type: typeDayOfWeek},
		}},
		{Path: "!=", Type: typeBool, Parameters: []Parameter{
			{Name: "value", Type: typeDayOfWeek},
		}},
		{Path: "oneOf", Type: typeBool, Variadic: true, Parameters: []Parameter{
			{Name: "values", Type: typeDayOfWeek},
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
	Name:        typeDuration,
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
	Name:        typeDate,
	Description: "A date without time or zone",
	As: map[TypeName]string{
		typeText: "text",
	},
	Values: []Value{
		{Path: "text", Type: typeText},
		{Path: "year", Type: typeInt},
		{Path: "month", Type: typeInt},
		{Path: "dayOfMonth", Type: typeInt},
		{Path: "dayOfWeek", Type: typeDayOfWeek},
		{Path: "add", Type: typeDate, Parameters: []Parameter{
			{Name: "amount", Type: typeInt},
			{Name: "duration", Type: typeDuration},
		}},
	},
	Parse: func(x string) (any, error) {
		return time.Parse(time.DateOnly, x)
	},
}, {
	Name:        typeDateTime,
	Description: "A date & time without zone",
	As: map[TypeName]string{
		typeText: "text",
	},
	Values: []Value{
		{Path: "text", Type: typeText},
		{Path: "date", Type: typeDate},
		{Path: "time", Type: typeTime},
		{Path: "hour", Type: typeInt},
		{Path: "minute", Type: typeInt, Aliases: []string{"min"}},
		{Path: "second", Type: typeInt},
		{Path: "year", Type: typeInt},
		{Path: "month", Type: typeInt},
		{Path: "dayOfMonth", Type: typeInt},
		{Path: "dayOfWeek", Type: typeDayOfWeek},
		{Path: "add", Type: typeDate, Parameters: []Parameter{
			{Name: "amount", Type: typeInt},
			{Name: "duration", Type: typeDuration},
		}},
	},
	Parse: func(x string) (any, error) {
		return time.Parse(time.DateTime, x)
	},
}, {
	Name:        typeTime,
	Description: "A time without zone",
	As: map[TypeName]string{
		typeText: "text",
	},
	Values: []Value{
		{Path: "text", Type: typeText},
		{Path: "hour", Type: typeInt},
		{Path: "minute", Type: typeInt, Aliases: []string{"min"}},
		{Path: "second", Type: typeInt},
		{Path: "add", Type: typeDate, Parameters: []Parameter{
			{Name: "amount", Type: typeInt},
			{Name: "duration", Type: typeDuration},
		}},
	},
	Parse: func(x string) (any, error) {
		return time.Parse(time.TimeOnly, x)
	},
}, {
	Name:        typeInt,
	Description: "A whole number",
	As: map[TypeName]string{
		typeText: "text",
	},
	Values: []Value{
		{Path: "text", Type: typeText},
		{Path: ">", Type: typeBool, Parameters: []Parameter{
			{Name: "value", Type: typeInt},
		}},
		{Path: ">=", Type: typeBool, Parameters: []Parameter{
			{Name: "value", Type: typeInt},
		}},
		{Path: "<", Type: typeBool, Parameters: []Parameter{
			{Name: "value", Type: typeInt},
		}},
		{Path: "<=", Type: typeBool, Parameters: []Parameter{
			{Name: "value", Type: typeInt},
		}},
		{Path: "=", Type: typeBool, Parameters: []Parameter{
			{Name: "value", Type: typeInt},
		}},
		{Path: "!=", Type: typeBool, Parameters: []Parameter{
			{Name: "value", Type: typeInt},
		}},
		{Path: "+", Type: typeInt, Parameters: []Parameter{
			{Name: "value", Type: typeInt},
		}},
		{Path: "-", Type: typeInt, Parameters: []Parameter{
			{Name: "value", Type: typeInt},
		}},
		{Path: "*", Type: typeInt, Parameters: []Parameter{
			{Name: "value", Type: typeInt},
		}},
		{Path: "/", Type: typeInt, Parameters: []Parameter{
			{Name: "value", Type: typeInt},
		}},
		{Path: "%", Type: typeInt, Parameters: []Parameter{
			{Name: "value", Type: typeInt},
		}},
	},
	Parse: func(x string) (any, error) {
		v, err := strconv.ParseInt(x, 10, 64)
		return int(v), err
	},
}, {
	Name:        typeBool,
	Description: "A true/false value",
	Enums:       []string{"true", "false"},
	As: map[TypeName]string{
		typeText: "text",
	},
	Values: []Value{
		{Path: "text", Type: typeText},
		{Path: "not", Type: typeBool},
		{Path: "and", Type: typeBool, Variadic: true, Parameters: []Parameter{
			{Name: "values", Type: typeBool},
		}},
		{Path: "or", Type: typeBool, Variadic: true, Parameters: []Parameter{
			{Name: "values", Type: typeBool},
		}},
		{Path: "then", Generic: true, Parameters: []Parameter{
			{Name: "trueValue", Generic: true},
			{Name: "falseValue", Generic: true},
		}},
	},
	Parse: func(x string) (any, error) {
		switch strings.ToLower(x) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		}
		return nil, fmt.Errorf("%s is not a valid true/false value", x)
	},
}, {
	Name:        typeUser,
	Description: "A user of our system",
	As: map[TypeName]string{
		typeText: "name",
	},
	Values: []Value{
		{Path: "name", Type: typeText},
		{Path: "createDate", Type: typeDateTime},
	},
}, {
	Name:        typeContext,
	Description: "The context for evaluating expressions.",
	Values: []Value{
		{Path: "time", Type: typeTimePackage},
		{Path: "user", Type: typeUser},
	},
}, {
	Name:        typeTimePackage,
	Description: "The package for date & time related expressions.",
	Values: []Value{
		{Path: "today", Type: typeDate},
		{Path: "yesterday", Type: typeDate},
		{Path: "tomorrow", Type: typeDate},
		{Path: "now", Type: typeDateTime},
		{Path: "current", Type: typeTime},
		{Path: "sunday", Type: typeDayOfWeek, Description: "An unambiguous way to refer to Sunday"},
		{Path: "monday", Type: typeDayOfWeek, Description: "An unambiguous way to refer to Monday"},
		{Path: "tuesday", Type: typeDayOfWeek, Description: "An unambiguous way to refer to Tuesday"},
		{Path: "wednesday", Type: typeDayOfWeek, Description: "An unambiguous way to refer to Wednesday"},
		{Path: "thursday", Type: typeDayOfWeek, Description: "An unambiguous way to refer to Thursday"},
		{Path: "friday", Type: typeDayOfWeek, Description: "An unambiguous way to refer to Friday"},
		{Path: "saturday", Type: typeDayOfWeek, Description: "An unambiguous way to refer to Saturday"},
	},
}, {
	Name: typeText,
	Values: []Value{
		{Path: "length", Type: typeInt, Aliases: []string{"len"}},
		{Path: "lower", Type: typeText},
		{Path: "upper", Type: typeText},
		{Path: "isLower", Type: typeBool},
		{Path: "isUpper", Type: typeBool},
		{Path: "+", Type: typeText, Parameters: []Parameter{
			{Name: "value", Type: typeText},
		}},
		{Path: "=", Type: typeBool, Parameters: []Parameter{
			{Name: "value", Type: typeText},
		}},
		{Path: "!=", Type: typeBool, Parameters: []Parameter{
			{Name: "value", Type: typeText},
		}},
		{Path: "contains", Type: typeBool, Parameters: []Parameter{
			{Name: "value", Type: typeText},
		}},
	},
	ParseOrder: -1,
	Parse: func(x string) (any, error) {
		return x, nil
	},
}})

var compileOptions = CompileOptions[Run]{
	Initial: func(root any) (any, error) {
		return root, nil
	},
	ConstantCompiler: func(e *Expr, root *Type, previous Run, arguments []Run) (Run, error) {
		return func(root any) (any, error) {
			return e.Parsed, nil
		}, nil
	},
	TypeCompilers: TypeCompilers[Run]{
		typeDateTime: ValueCompilers[Run]{
			"hour": runCompiler(func(v time.Time, args []any) (any, error) {
				return v.Hour(), nil
			}),
			"minute": runCompiler(func(v time.Time, args []any) (any, error) {
				return v.Minute(), nil
			}),
		},
		typeInt: ValueCompilers[Run]{
			"text": runCompiler(func(v int, args []any) (any, error) {
				return fmt.Sprintf("%d", v), nil
			}),
			"=": runCompiler(func(v int, args []any) (any, error) {
				return v == args[0].(int), nil
			}),
			">": runCompiler(func(v int, args []any) (any, error) {
				return v > args[0].(int), nil
			}),
			">=": runCompiler(func(v int, args []any) (any, error) {
				return v >= args[0].(int), nil
			}),
		},
		typeBool: ValueCompilers[Run]{
			"text": runCompiler(func(v bool, args []any) (any, error) {
				return strconv.FormatBool(v), nil
			}),
			"and": runCompiler(func(v bool, args []any) (any, error) {
				if !v {
					return false, nil
				}
				for _, a := range args {
					if !a.(bool) {
						return false, nil
					}
				}
				return true, nil
			}),
			"then": runCompilerLazy(func(v bool, args []func() (any, error)) (any, error) {
				if v {
					return args[0]()
				} else {
					return args[1]()
				}
			}),
		},
		typeText: ValueCompilers[Run]{
			"length": runCompiler(func(v string, args []any) (any, error) {
				return len(v), nil
			}),
			"contains": runCompiler(func(v string, args []any) (any, error) {
				return strings.Contains(v, args[0].(string)), nil
			}),
			"lower": runCompiler(func(v string, args []any) (any, error) {
				return strings.ToLower(v), nil
			}),
		},
		typeUser:        mapValueCompiler("name", "createDate"),
		typeContext:     mapValueCompiler("time", "user"),
		typeTimePackage: mapValueCompiler("now", "sunday"),
	},
}

func TestIt(t *testing.T) {
	tests := []struct {
		name           string
		options        Options
		expectedString string
		input          any
		expectedValue  any
		expectedType   TypeName
		expectedError  string
		postParseCheck func(*Expr, *testing.T)
	}{{
		name: "complex",
		options: Options{
			RootType:      typeContext,
			ExpectedTypes: []TypeName{typeBool},
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
			RootType:   typeContext,
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
			RootType:   typeContext,
			Expression: "time.now.hour",
		},
		expectedString: "time.now.hour",
		input: map[string]any{
			"time": map[string]any{
				"now": time.Date(2023, 4, 11, 13, 0, 0, 0, time.Local),
			},
		},
		expectedValue: int(13),
	}, {
		name: "sunday",
		options: Options{
			RootType:   typeContext,
			Expression: "sunday",
		},
		expectedString: "'sunday'",
		input:          map[string]any{},
		expectedValue:  "sunday",
		expectedType:   typeDayOfWeek,
	}, {
		name: "time.sunday",
		options: Options{
			RootType:   typeContext,
			Expression: "time.Sunday",
		},
		expectedString: "time.Sunday",
		input: map[string]any{
			"time": map[string]any{
				"sunday": "sunday",
			},
		},
		expectedValue: "sunday",
		expectedType:  typeDayOfWeek,
	}, {
		name: "time.sunday with tab",
		options: Options{
			RootType:   typeContext,
			Expression: "\n\ttime.\tSunday\n",
		},
		expectedString: "time.Sunday",
		input: map[string]any{
			"time": map[string]any{
				"sunday": "sunday",
			},
		},
		expectedValue: "sunday",
		expectedType:  typeDayOfWeek,
	}, {
		name: "time.sun error",
		options: Options{
			RootType:   typeContext,
			Expression: "time.sun",
		},
		expectedError: "invalid value sun",
	}, {
		name: "time.sun error check",
		options: Options{
			RootType:   typeContext,
			Expression: "time.sun",
		},
		postParseCheck: func(e *Expr, t *testing.T) {
			assert.Equal(t, e.Token, "time")
			assert.NotNil(t, e.Type)
			assert.Equal(t, e.Type.Name, typeTimePackage)
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
			RootType:   typeContext,
			Expression: "time.",
		},
		postParseCheck: func(e *Expr, t *testing.T) {
			assert.Equal(t, e.Token, "time")
			assert.NotNil(t, e.Type)
			assert.Equal(t, e.Type.Name, typeTimePackage)
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
			RootType:   typeContext,
			Expression: "true",
		},
		expectedType:  typeBool,
		expectedValue: true,
	}, {
		name: "bool detect convert to text",
		options: Options{
			RootType:      typeContext,
			Expression:    "true",
			ExpectedTypes: []TypeName{typeText},
		},
		expectedType:  typeText,
		expectedValue: "true",
	}, {
		name: "bool detect expect many stay same",
		options: Options{
			RootType:      typeContext,
			Expression:    "true",
			ExpectedTypes: []TypeName{typeBool, typeText},
		},
		expectedType:  typeBool,
		expectedValue: true,
	}, {
		name: "bool detect expect many convert to text",
		options: Options{
			RootType:      typeContext,
			Expression:    "true",
			ExpectedTypes: []TypeName{typeInt, typeText},
		},
		expectedType:  typeText,
		expectedValue: "true",
	}, {
		name: "variadic",
		options: Options{
			RootType:   typeUser,
			Expression: "name.len=(12).and(createDate.hour>(12), createDate.min=(0))",
		},
		input: map[string]any{
			"name":       "pmd@site.com",
			"createDate": time.Date(2023, 1, 1, 14, 0, 1, 2, time.Local),
		},
		expectedType:  typeBool,
		expectedValue: true,
	}, {
		name: "generic true",
		options: Options{
			RootType:   typeUser,
			Expression: "name.len=(12).then(A, B)",
		},
		input: map[string]any{
			"name": "pmd@site.com",
		},
		expectedType:  typeText,
		expectedValue: "A",
	}, {
		name: "generic false",
		options: Options{
			RootType:   typeUser,
			Expression: "name.len=(13).then(A, B)",
		},
		input: map[string]any{
			"name": "pmd@site.com",
		},
		expectedType:  typeText,
		expectedValue: "B",
	}, {
		name: "generic mixed types",
		options: Options{
			RootType:   typeUser,
			Expression: "name.len=(12).then(A, name.len)",
		},
		input: map[string]any{
			"name": "pmd@site.com",
		},
		expectedType:  typeText,
		expectedValue: "A",
	}, {
		name: "generic mixed types convert",
		options: Options{
			RootType:   typeUser,
			Expression: "name.len=(13).then(A, name.len)",
		},
		input: map[string]any{
			"name": "pmd@site.com",
		},
		expectedType:  typeText,
		expectedValue: "12",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := sys.Parse(test.options)

			if test.postParseCheck != nil {
				test.postParseCheck(expr, t)
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

			compiled, err := Compile(expr, compileOptions)
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

func runCompiler[T any](call func(v T, args []any) (any, error)) Compiler[Run] {
	return func(e *Expr, root *Type, previous Run, arguments []Run) (Run, error) {
		return func(root any) (any, error) {
			prev, err := previous(root)
			if err != nil {
				return nil, err
			}
			args := make([]any, len(arguments))
			for i := range args {
				args[i], err = arguments[i](root)
				if err != nil {
					return nil, err
				}
			}
			if asType, ok := prev.(T); ok {
				return call(asType, args)
			} else {
				return nil, fmt.Errorf("unexpected type: %v, wanted %v", reflect.TypeOf(prev), reflect.TypeOf((*T)(nil)).Elem())
			}
		}, nil
	}
}

func runCompilerLazy[T any](call func(v T, args []func() (any, error)) (any, error)) Compiler[Run] {
	return func(e *Expr, root *Type, previous Run, arguments []Run) (Run, error) {
		return func(root any) (any, error) {
			prev, err := previous(root)
			if err != nil {
				return nil, err
			}
			args := make([]func() (any, error), len(arguments))
			for i := range args {
				k := i
				args[k] = func() (any, error) {
					return arguments[k](root)
				}
			}
			if asType, ok := prev.(T); ok {
				return call(asType, args)
			} else {
				return nil, fmt.Errorf("unexpected type: %v, wanted %v", reflect.TypeOf(prev), reflect.TypeOf((*T)(nil)).Elem())
			}
		}, nil
	}
}

func mapValueCompiler(keys ...string) ValueCompilers[Run] {
	vc := ValueCompilers[Run]{}
	for i := range keys {
		key := keys[i]
		vc[strings.ToLower(key)] = runCompiler(func(v map[string]any, args []any) (any, error) {
			return v[key], nil
		})
	}
	return vc
}
