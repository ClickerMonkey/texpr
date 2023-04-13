package texpr

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"
)

func TypeOf[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}
func NameOf[T any]() TypeName {
	return TypeName(TypeOf[T]().Name())
}

type ReflectConversion struct {
	Type        TypeName
	ConvertTo   func(v any) (any, error)
	ConvertFrom func(v any) (any, error)
}

type ReflectOptions struct {
	Conversions map[reflect.Type]ReflectConversion
	Types       map[reflect.Type]Type
}

type Reflect struct {
	options ReflectOptions
	system  System
}

func NewReflect(options ReflectOptions) (r *Reflect, err error) {
	r = &Reflect{
		options: options,
	}

	if options.Conversions == nil {
		options.Conversions = make(map[reflect.Type]ReflectConversion)
	}

	supportedTypes := make(map[reflect.Type]TypeName, len(options.Types)+len(options.Conversions))
	for rt, t := range options.Types {
		if t.Name == "" {
			t.Name = TypeName(rt.Name())
		}
		supportedTypes[rt] = t.Name
		options.Types[rt] = t
	}
	for rt, c := range options.Conversions {
		supportedTypes[rt] = c.Type
	}

	systemTypes := make([]Type, 0, len(options.Types))

	for rt, t := range options.Types {
		if t.Parse == nil && reflect.PointerTo(rt).Implements(TypeOf[encoding.TextUnmarshaler]()) {
			t.Parse = func(x string) (any, error) {
				y, ok := reflect.New(rt).Interface().(encoding.TextUnmarshaler)
				if ok {
					err := y.UnmarshalText([]byte(x))
					return y, err
				}
				return nil, fmt.Errorf("unmarshalling of %v is not supported for %s", rt, x)
			}
		}

		if rt.Kind() == reflect.Struct {
			fields := getFields(rt)
			for path, field := range fields {
				if supportedTypes[field.Type] == "" {
					continue
				}

				value, valueIndex := findValue(path, t)
				if value == nil {
					t.Values = append(t.Values, Value{})
					value = &t.Values[len(t.Values)-1]
				}
				if value.Path == "" {
					value.Path = path
				}
				if value.Type == "" {
					value.Type = supportedTypes[field.Type]
				}
				if valueIndex != -1 {
					t.Values[valueIndex] = *value
				}
			}
		}

		methods := rt.NumMethod()
		for i := 0; i < methods; i++ {
			m := rt.Method(i)
			mOut := m.Type.NumOut()
			if mOut < 0 || mOut > 2 || (mOut == 2 && !m.Type.Out(1).Implements(TypeOf[error]())) || supportedTypes[m.Type.Out(0)] == "" {
				continue
			}
			mIn := m.Type.NumIn()
			skip := false
			for k := 1; k < mIn && !skip; k++ {
				if m.Type.IsVariadic() && k == mIn-1 {
					if supportedTypes[m.Type.In(k).Elem()] == "" {
						skip = true
					}
				} else if supportedTypes[m.Type.In(k)] == "" {
					skip = true
				}
			}
			if skip {
				continue
			}

			value, valueIndex := findValue(m.Name, t)
			if value == nil {
				t.Values = append(t.Values, Value{})
				value = &t.Values[len(t.Values)-1]
			}
			if value.Path == "" {
				value.Path = m.Name
			}
			if value.Type == "" {
				value.Type = supportedTypes[m.Type.Out(0)]
			}

			if m.Type.IsVariadic() {
				value.Variadic = true
			}

			for k := 1; k < mIn; k++ {
				in := m.Type.In(k)
				param := Parameter{}
				if m.Type.IsVariadic() && k == mIn-1 {
					param.Type = supportedTypes[in.Elem()]
				} else {
					param.Type = supportedTypes[in]
				}
				value.Parameters = append(value.Parameters, param)
			}
			if valueIndex != -1 {
				t.Values[valueIndex] = *value
			}
		}

		systemTypes = append(systemTypes, t)
		options.Types[rt] = t
	}

	r.system, err = NewSystem(systemTypes)

	return
}

func findValue(token string, t Type) (*Value, int) {
	if len(t.Values) == 0 {
		return nil, -1
	}
	for i := range t.Values {
		v := &t.Values[i]
		if strings.EqualFold(v.Path, token) {
			return v, i
		}
		if len(v.Aliases) > 0 {
			for _, a := range v.Aliases {
				if strings.EqualFold(a, token) {
					return v, i
				}
			}
		}
	}
	return nil, -1
}

func getFields(rt reflect.Type) map[string]reflect.StructField {
	m := make(map[string]reflect.StructField)
	fields := rt.NumField()
	for i := 0; i < fields; i++ {
		field := rt.Field(i)
		if field.Anonymous {
			anon := getFields(field.Type)
			for k, v := range anon {
				m[k] = v
			}
		} else {
			m[strings.ToLower(field.Name)] = field
		}
	}
	return m
}

type ReflectRun func(root any) (any, error)

func (r Reflect) Parse(opts Options) (*Expr, error) {
	return r.system.Parse(opts)
}

func (r Reflect) Compile(e *Expr) (ReflectRun, error) {
	return func(root any) (any, error) {
		return root, nil
	}, nil
}
