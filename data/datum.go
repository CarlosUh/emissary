package data

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/fatih/structs"
	"github.com/maxwellhealth/go-dotaccess"
	// "github.com/maxwellhealth/mergo"
	"github.com/dustin/go-humanize"
	"github.com/soniah/evaler"
	"html/template"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var DefaultDecimalCount = 8

var TimeFormats = []string{
	"2006-01-02 15:04:05",
	time.RFC3339,
	time.RFC1123,
	time.RFC3339Nano,
}

func parseFloat(arg interface{}) (float64, error) {
	switch arg.(type) {
	case string:
		return strconv.ParseFloat(arg.(string), 64)
	case int:
		f := arg.(int)
		return float64(f), nil
	case int32:
		f := arg.(int32)
		return float64(f), nil
	case int64:
		f := arg.(int64)
		return float64(f), nil
	case float32:
		f := arg.(float32)
		return float64(f), nil
	case float64:
		f := arg.(float64)
		return f, nil
	default:
		return 0.0, errors.New("Invalid float")

	}
}

var funcMap = template.FuncMap{
	"add": func(args ...interface{}) string {
		total := 0.0
		for _, arg := range args {
			f, _ := parseFloat(arg)
			total += f
		}

		return fmt.Sprint(total)
	},
	"mult": func(args ...interface{}) string {
		total := 1.0
		for _, arg := range args {
			f, _ := parseFloat(arg)
			total *= f
		}

		return fmt.Sprint(total)
	},
	"currency": func(args ...interface{}) string {
		arg := args[0]

		input := fmt.Sprint(arg)

		spl := strings.Split(input, ".")
		intval := spl[0]
		if len(spl) == 1 {
			spl = append(spl, "00")
		}

		// Parse the int
		ival, err := strconv.ParseInt(intval, 10, 64)
		if err != nil {
			return input
		}
		intval = humanize.Comma(ival)

		dec := spl[1]
		parsed, _ := strconv.ParseFloat("."+dec, 64)

		var decval string
		if parsed == 0 {
			decval = ".00"
		} else {
			decval = fmt.Sprint(Round(parsed, .5, 2))[1:]
		}

		if len(decval) == 2 {
			decval = decval + "0"
		}
		return "$" + intval + decval
	},
	"substring": func(args ...interface{}) string {
		// First arg determines how to parse it
		if len(args) < 2 {
			panic("substring formatter requires one numeric argument")
		}

		arg := args[0]
		length := args[1]
		var strval string
		var ok bool
		if strval, ok = arg.(string); !ok {
			return ""
		}

		var intval int
		if intval, ok = length.(int); !ok {
			return ""
		}

		if intval < 0 {
			return strval[(len(strval) + intval):]
		} else {
			return strval[:intval]
		}
	},
	"date": func(args ...interface{}) string {
		if len(args) == 0 {
			return ""
		}

		arg := args[0]

		var dateval time.Time
		var ok bool
		if dateval, ok = arg.(time.Time); !ok {
			return "Invalid Date"
		}

		format := "02/01/2006"
		if len(args) >= 2 {
			if strval, ok := args[1].(string); ok {
				format = strval
			}

		}

		return dateval.Format(format)
	},
}

type Datum struct {
	Source interface{}
}

// Converts the source (map or struct) to a map so that the template engine won't panic when trying to access
// invalid properties.
// If the src is a struct, look at the tagName to see what the map keys should be when converted to a map
func (d *Datum) SetSource(src interface{}, tagName string) {
	value := reflect.ValueOf(src)
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() == reflect.Struct {
		converter := structs.New(src)
		converter.TagName = tagName
		mp := converter.Map()
		d.Source = mp
	} else if value.Kind() == reflect.Map {
		d.Source = src
	} else {
		panic("Invalid type for emissary datum source (" + value.Kind().String() + ")")
	}

}

// A getter is any type that can Get a key, with a default value and return a string
type Getter interface {
	Get(key string, defaultValue string) string
}

var ConditionOperators = map[string]ConditionOperator{
	"eq": func(args []interface{}) bool {
		if len(args) < 2 {
			return false
		}

		return args[0] == args[1]
	},
	"ne": func(args []interface{}) bool {
		if len(args) < 2 {
			return false
		}

		return args[0] != args[1]
	},
	"in": func(args []interface{}) bool {
		if len(args) < 2 {
			return false
		}

		var comp interface{}

		for i, a := range args {
			if i == 0 {
				comp = a
			} else {
				if comp == a {
					return true
				}
			}
		}
		return false
	},
	"nin": func(args []interface{}) bool {
		if len(args) < 2 {
			return false
		}

		var comp interface{}

		for i, a := range args {
			if i == 0 {
				comp = a
			} else {
				if comp == a {
					return false
				}
			}
		}
		return true
	},
	"gt": func(args []interface{}) bool {
		if len(args) < 2 {
			return false
		}
		// Type assertion to float?
		f1, ok1 := args[0].(float64)
		f2, ok2 := args[1].(float64)

		if ok1 && ok2 {
			return f1 > f2
		}

		d1, ok1 := args[0].(time.Time)
		d2, ok2 := args[1].(time.Time)

		if ok1 && ok2 {
			return d1.After(d2)
		}

		// return args[0] > args[1]
		return false
	},
	"gte": func(args []interface{}) bool {
		if len(args) < 2 {
			return false
		}
		// Type assertion to float?
		f1, ok1 := args[0].(float64)
		f2, ok2 := args[1].(float64)

		if ok1 && ok2 {
			return f1 >= f2
		}

		d1, ok1 := args[0].(time.Time)
		d2, ok2 := args[1].(time.Time)

		if ok1 && ok2 {
			return d1.After(d2.Add(-1 * time.Nanosecond))
		}

		// return args[0] > args[1]
		return false
	},
	"lt": func(args []interface{}) bool {
		if len(args) < 2 {
			return false
		}
		// Type assertion to float?
		f1, ok1 := args[0].(float64)
		f2, ok2 := args[1].(float64)

		if ok1 && ok2 {
			return f1 < f2
		}

		d1, ok1 := args[0].(time.Time)
		d2, ok2 := args[1].(time.Time)

		if ok1 && ok2 {
			return d1.Before(d2)
		}

		// return args[0] > args[1]
		return false
	},
	"lte": func(args []interface{}) bool {
		if len(args) < 2 {
			return false
		}
		// Type assertion to float?
		f1, ok1 := args[0].(float64)
		f2, ok2 := args[1].(float64)

		if ok1 && ok2 {
			return f1 <= f2
		}

		d1, ok1 := args[0].(time.Time)
		d2, ok2 := args[1].(time.Time)

		if ok1 && ok2 {
			return d1.Before(d2.Add(1 * time.Nanosecond))
		}

		// return args[0] > args[1]
		return false
	},
}

func (d *Datum) Get(key string, defaultValue string) string {
	tmpl, err := template.New("tmpl").Funcs(funcMap).Parse(key)
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, d.Source)
	if err != nil {
		panic(err)
	}

	return buf.String()

}
