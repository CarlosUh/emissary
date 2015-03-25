package data

import (
	"errors"
	"fmt"
	"github.com/maxwellhealth/go-dotaccess"
	"github.com/soniah/evaler"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var dynamicValue = regexp.MustCompile("\\$([A-Za-z0-9_\\.]+)")
var compoundDef = regexp.MustCompile("\\{([^}]+)\\}")
var conditionBlock = regexp.MustCompile("\\[if(.)+\\[/if\\]")
var conditionPattern = regexp.MustCompile("\\[[^\\]]+\\]")

var DefaultDecimalCount = 8

var TimeFormats = []string{
	"2006-01-02 15:04:05",
	time.RFC3339,
	time.RFC1123,
	time.RFC3339Nano,
}

func parseDate(date string) (time.Time, error) {
	for _, t := range TimeFormats {
		parsed, err := time.Parse(t, date)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, errors.New("Invalid date")
}

type Getter interface {
	Get(key string) (interface{}, error)
}

type Datum struct {
	Source Getter
}

type GettableMap map[string]interface{}

func (g GettableMap) Get(key string) (interface{}, error) {
	return dotaccess.Get(g, key)
}

type ConditionOperator func(args []interface{}) bool

func FormatConditionArguments(args []string) []interface{} {
	formatted := make([]interface{}, len(args))

	for i, a := range args {
		// Number?
		f, err := strconv.ParseFloat(a, 64)
		if err == nil {
			formatted[i] = f
			continue
		}

		// Date?

		parsedDate, err := parseDate(a)
		if err == nil {
			formatted[i] = parsedDate
			continue
		}

		formatted[i] = a
	}

	return formatted
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

func (d *Datum) Get(key string, defaultValue interface{}) string {
	// Check condition blocks first
	conditions := conditionBlock.FindAllString(key, -1)
	// var err error
	for _, c := range conditions {
		key, _ = d.parseCondition(c, key)
	}
	// There may be compound definitions
	matches := compoundDef.FindAllString(key, -1)

	if len(matches) > 0 {
		for _, m := range matches {
			data := strings.TrimPrefix(strings.TrimSuffix(m, "}"), "{")
			key = strings.Replace(key, m, d.getParsedVal(data), 1)
		}
		return key
	}

	return d.getParsedVal(key)
}

func (d *Datum) parseCondition(condition string, key string) (string, error) {
	blocks := conditionPattern.FindAllString(condition, -1)

	if len(blocks) < 2 {
		return "", errors.New("Invalid conditional block")
	}

	if !strings.HasPrefix(blocks[0], "[if ") {
		return "", errors.New("Invalid conditional block. Must open with [if")
	}

	for i, b := range blocks {
		// Skip the last one; it's expected to be an [/if]
		if (i + 1) == len(blocks) {
			break
		}
		// Get the result
		index := strings.Index(condition, b) + len(b)
		nextIndex := strings.Index(condition, blocks[i+1])
		result := condition[index:nextIndex]
		for i := index; i < len(condition); i++ {
			if d.evaluateBlock(b) {
				return strings.Replace(key, condition, result, 1), nil
			}
		}
	}

	return "", nil
}

func (d *Datum) evaluateBlock(block string) bool {
	if block == "[else]" {
		return true
	}

	block = strings.TrimSuffix(block, "]")

	if strings.HasPrefix(block, "[if ") {
		block = strings.TrimPrefix(block, "[if ")
	} else if strings.HasPrefix(block, "[else if ") {
		block = strings.TrimPrefix(block, "[else if ")
	} else {
		panic("Invalid conditional block")
	}

	// Replace dynamic values with quoted values
	dynamicValues := dynamicValue.FindAllString(block, -1)
	for _, v := range dynamicValues {
		block = strings.Replace(block, v, d.getParsedVal(v), 1)
	}

	parsedArgs := []string{}
	currentArg := []byte{}
	reader := strings.NewReader(block)

	inQuote := false
	length := reader.Len()
	for i := 0; i < length; i++ {
		b, _ := reader.ReadByte()
		if b == '"' {
			if !inQuote {
				inQuote = true
			} else {
				parsedArgs = append(parsedArgs, string(currentArg))
				currentArg = []byte{}
				inQuote = false
			}
		} else if b == ' ' {
			if inQuote {
				currentArg = append(currentArg, b)
			} else if len(currentArg) > 0 {
				parsedArgs = append(parsedArgs, string(currentArg))
				currentArg = []byte{}
			}
		} else {
			currentArg = append(currentArg, b)
		}
	}
	if len(currentArg) > 0 {
		parsedArgs = append(parsedArgs, string(currentArg))
	}

	function := parsedArgs[0]

	formattedArgs := FormatConditionArguments(parsedArgs[1:])
	if f, ok := ConditionOperators[function]; ok {
		return f(formattedArgs)
	}
	return false

}

func (d *Datum) getParsedVal(key string) string {
	var err error
	// Split the key by pipes to get formatting arguments, etc
	spl := strings.Split(key, "|")

	val := spl[0]

	// Parse the first argument to get any math, dynamic values, etc

	// Step 1) Replace any values that start with "$" with the values from the map
	matches := dynamicValue.FindAllString(val, -1)

	for _, m := range matches {
		k := strings.TrimPrefix(m, "$")

		// Ignore error because if it's an error we'll just use ""
		v, _ := d.Source.Get(k)

		if v == nil {
			v = ""
		}

		parsedStr := ""

		// If it's a date, spit it out in a readable format
		if timeVal, ok := v.(time.Time); ok {
			parsedStr = timeVal.Format(time.RFC3339Nano)
		} else {
			parsedStr = fmt.Sprint(v)
		}
		val = strings.Replace(val, m, parsedStr, 1)
	}

	// Compile. The Eval will fail if it's not all math-able. BUT, only if it's not a date
	_, err = parseDate(val)
	if err != nil {
		res, err := evaler.Eval(val)
		if err == nil {
			// Formatting will be applied later - get it as big as needed
			val = res.FloatString(DefaultDecimalCount)
		}
	}

	// Now format
	if len(spl) > 1 {
		format := spl[1]

		args := spl[2:]

		if formatter, ok := Formatters[format]; ok {
			val, _ = formatter(val, args)
		}
	}
	return val
}

func getNestedProperty(data map[string]interface{}, key string) interface{} {
	spl := strings.Split(key, ".")

	var v interface{}
	length := len(spl)
	for i, k := range spl {
		v = getProperty(data, k)

		if v == nil {
			return nil
		}

		// If we're not yet at the end and don't have a map, also return nil
		if (i + 1) < length {
			var ok bool

			if data, ok = v.(map[string]interface{}); !ok {
				return nil
			}
		}
	}

	return v
}

func getProperty(data map[string]interface{}, property string) interface{} {
	if v, ok := data[property]; ok {
		return v
	}
	return nil
}
