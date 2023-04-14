package texpr

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

type Int int

func (i Int) Add(other Int) Int {
	return i + other
}

func (i Int) Equals(other Int) Bool {
	return i == other
}

func (i Int) Gt(other Int) Bool {
	return i > other
}

func (i Int) String() String {
	return String(strconv.Itoa(int(i)))
}

type Bool bool

func (b Bool) And(others ...Bool) Bool {
	if !b {
		return false
	}
	for _, o := range others {
		if !o {
			return false
		}
	}
	return true
}

func (b Bool) Or(others ...Bool) Bool {
	if b {
		return true
	}
	for _, o := range others {
		if o {
			return true
		}
	}
	return false
}

func (b Bool) Not() Bool {
	return !b
}

func (b Bool) String() String {
	if b {
		return "true"
	} else {
		return "false"
	}
}

type String string

func (s String) Lower() String {
	return String(strings.ToLower(string(s)))
}

type TimePackage struct {
	Now    time.Time
	Today  time.Time
	Sunday time.Weekday
}

type MessageContext struct {
	Message String
	Time    TimePackage
}

func TestReflect(t *testing.T) {
	r, err := NewReflect(ReflectOptions{
		Conversions: map[reflect.Type]ReflectConversion{
			TypeOf[int](): {
				Type: NameOf[Int](),
				ConvertTo: func(v any) (any, error) {
					return Int(v.(int)), nil
				},
				ConvertFrom: func(v any) (any, error) {
					return int(v.(Int)), nil
				},
			},
			TypeOf[bool](): {
				Type: NameOf[Bool](),
				ConvertTo: func(v any) (any, error) {
					return Bool(v.(bool)), nil
				},
				ConvertFrom: func(v any) (any, error) {
					return bool(v.(Bool)), nil
				},
			},
			TypeOf[string](): {
				Type: NameOf[String](),
				ConvertTo: func(v any) (any, error) {
					return String(v.(string)), nil
				},
				ConvertFrom: func(v any) (any, error) {
					return string(v.(String)), nil
				},
			},
		},
		Types: map[reflect.Type]Type{
			TypeOf[Int]():            {Parse: func(x string) (any, error) { return strconv.Atoi(x) }},
			TypeOf[Bool]():           {Parse: func(x string) (any, error) { return strconv.ParseBool(x) }},
			TypeOf[String]():         {ParseOrder: -1, Parse: func(x string) (any, error) { return x, nil }},
			TypeOf[TimePackage]():    {},
			TypeOf[MessageContext](): {},
			TypeOf[time.Time]():      {Parse: func(x string) (any, error) { return time.Parse(time.DateTime, x) }},
			TypeOf[time.Weekday](): {
				Enums: []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"},
				Parse: func(x string) (any, error) {
					options := map[string]time.Weekday{
						"sunday":    time.Sunday,
						"monday":    time.Monday,
						"tuesday":   time.Tuesday,
						"wednesday": time.Wednesday,
						"thursday":  time.Thursday,
						"friday":    time.Friday,
						"saturday":  time.Saturday,
					}
					found, exists := options[strings.ToLower(x)]
					if !exists {
						return nil, fmt.Errorf("%s is not a valid weekday", x)
					}
					return found, nil
				},
			},
		},
	})

	t.Run("reflect", func(t *testing.T) {

		if err != nil {
			panic(err)
		}

		e, err := r.Parse(Options{
			RootType:   NameOf[MessageContext](),
			Expression: "time.now.hour.add(1).equals(8)",
		})

		if err != nil {
			panic(err)
		}

		eval := r.Compile(e)

		v, err := eval(MessageContext{
			Message: "Hello World!",
			Time: TimePackage{
				Now:    time.Now(),
				Today:  time.Now().Local(),
				Sunday: time.Sunday,
			},
		})

		if err != nil {
			panic(err)
		}

		fmt.Printf("Reflect expression result: %v", v)
	})
}
