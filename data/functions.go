package data

import (
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"html/template"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type kind int

var (
	errBadComparisonType = errors.New("invalid type for comparison")
	errBadComparison     = errors.New("incompatible types for comparison")
	errNoComparison      = errors.New("missing argument for comparison")
)

const (
	invalidKind kind = iota
	boolKind
	complexKind
	intKind
	floatKind
	integerKind
	stringKind
	uintKind
)

func Round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func basicKind(v reflect.Value) (kind, error) {
	switch v.Kind() {
	case reflect.Bool:
		return boolKind, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intKind, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintKind, nil
	case reflect.Float32, reflect.Float64:
		return floatKind, nil
	case reflect.Complex64, reflect.Complex128:
		return complexKind, nil
	case reflect.String:
		return stringKind, nil
	}
	return invalidKind, errBadComparisonType
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

// MATH FUNCTIONS

func add(args ...interface{}) string {
	total := 0.0
	for _, arg := range args {
		f, _ := parseFloat(arg)
		total += f
	}

	return fmt.Sprint(total)
}

func sub(args ...interface{}) string {
	total := 0.0
	for i, arg := range args {
		f, _ := parseFloat(arg)
		if i == 0 {
			total = f
		} else {
			total -= f
		}

	}

	return fmt.Sprint(total)
}

func mult(args ...interface{}) string {
	total := 1.0
	for _, arg := range args {
		f, _ := parseFloat(arg)
		total *= f
	}

	return fmt.Sprint(total)
}

func div(args ...interface{}) string {
	total := 0.0
	for i, arg := range args {
		f, _ := parseFloat(arg)

		if i == 0 {
			total = f
		} else {
			total = total / f
		}

	}

	return fmt.Sprint(total)
}

// FORMATTERS
func currency(args ...interface{}) string {
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
}

func substring(args ...interface{}) (string, error) {
	// First arg determines how to parse it
	if len(args) < 2 {
		return "", errors.New("substring formatter requires one numeric argument")
	}

	arg := args[0]
	length := args[1]
	var strval string
	var ok bool
	if strval, ok = arg.(string); !ok {
		return "", nil
	}

	var intval int
	if intval, ok = length.(int); !ok {
		return "", nil
	}

	if intval < 0 {
		return strval[(len(strval) + intval):], nil
	} else {
		return strval[:intval], nil
	}
}

func date(args ...interface{}) string {
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
}

func number(args ...interface{}) (string, error) {
	if len(args) == 0 {
		return "", errors.New("Not enough arguments for number")
	}

	arg1, err := parseFloat(args[0])
	if err != nil {
		return "", err
	}

	// Round to decimal places
	var round int
	if len(args) < 1 {
		round = 0
	} else {
		arg2, err := parseFloat(args[1])
		if err != nil {
			return "", err
		}

		round = int(arg2)
	}

	parsed := Round(arg1, .5, round)

	format := fmt.Sprintf("%d", round)
	return fmt.Sprintf("%."+format+"f", parsed), nil
}

// COMPARATORS

func eq(arg1 interface{}, arg2 ...interface{}) (bool, error) {
	// Addition - if arg1 and arg2 are both time.Time
	var t1 time.Time
	var t2 time.Time
	foundT1 := false
	foundT2 := false
	var err error
	if t, ok := arg1.(time.Time); ok {
		t1 = t
		foundT1 = true
	} else if t, ok := arg1.(GetTime); ok {
		t1, err = t.GetTime()
		if err != nil {
			return false, err
		}
		foundT1 = true
	}
	if t, ok := arg2[0].(time.Time); ok {
		t2 = t
		foundT2 = true
	} else if t, ok := arg2[0].(GetTime); ok {
		t2, err = t.GetTime()
		if err != nil {
			return false, err
		}
		foundT2 = true
	}

	if foundT1 && foundT2 {
		return t1 == t2, nil
	} else if foundT1 || foundT2 {
		return false, errors.New("Cannot compare date with non-date")
	}

	v1 := reflect.ValueOf(arg1)
	k1, err := basicKind(v1)
	if err != nil {
		return false, err
	}
	if len(arg2) == 0 {
		return false, errNoComparison
	}
	for _, arg := range arg2 {
		v2 := reflect.ValueOf(arg)
		k2, err := basicKind(v2)
		if err != nil {
			return false, err
		}
		truth := false
		if k1 != k2 {
			// Special case: Can compare integer values regardless of type's sign.
			switch {
			case k1 == intKind && k2 == uintKind:
				truth = v1.Int() >= 0 && uint64(v1.Int()) == v2.Uint()
			case k1 == uintKind && k2 == intKind:
				truth = v2.Int() >= 0 && v1.Uint() == uint64(v2.Int())
			default:
				return false, errors.New("Invalid comparison")
			}
		} else {
			switch k1 {
			case boolKind:
				truth = v1.Bool() == v2.Bool()
			case complexKind:
				truth = v1.Complex() == v2.Complex()
			case floatKind:
				truth = v1.Float() == v2.Float()
			case intKind:
				truth = v1.Int() == v2.Int()
			case stringKind:
				truth = v1.String() == v2.String()
			case uintKind:
				truth = v1.Uint() == v2.Uint()
			default:
				panic("invalid kind")
			}
		}
		if truth {
			return true, nil
		}
	}
	return false, nil
}

func neq(arg1, arg2 interface{}) (bool, error) {
	// != is the inverse of ==.
	equal, err := eq(arg1, arg2)
	return !equal, err
}

type GetTime interface {
	GetTime() (time.Time, error)
}

func lt(arg1 interface{}, arg2 interface{}) (bool, error) {
	// Addition - if arg1 and arg2 are both time.Time
	var t1 time.Time
	var t2 time.Time
	foundT1 := false
	foundT2 := false
	var err error
	if t, ok := arg1.(time.Time); ok {
		t1 = t
		foundT1 = true
	} else if t, ok := arg1.(GetTime); ok {
		t1, err = t.GetTime()
		if err != nil {
			return false, err
		}
		foundT1 = true
	}
	if t, ok := arg2.(time.Time); ok {
		t2 = t
		foundT2 = true
	} else if t, ok := arg2.(GetTime); ok {
		t2, err = t.GetTime()
		if err != nil {
			return false, err
		}
		foundT2 = true
	}

	if foundT1 && foundT2 {
		return t2.After(t1), nil
	} else if foundT1 || foundT2 {
		return false, errors.New("Cannot compare date with non-date")
	}

	// Normal stuff
	v1 := reflect.ValueOf(arg1)
	k1, err := basicKind(v1)
	if err != nil {
		return false, err
	}
	v2 := reflect.ValueOf(arg2)
	k2, err := basicKind(v2)
	if err != nil {
		return false, err
	}
	truth := false
	if k1 != k2 {
		// Special case: Can compare integer values regardless of type's sign.
		switch {
		case k1 == intKind && k2 == uintKind:
			truth = v1.Int() < 0 || uint64(v1.Int()) < v2.Uint()
		case k1 == uintKind && k2 == intKind:
			truth = v2.Int() >= 0 && v1.Uint() < uint64(v2.Int())
		default:
			return false, errBadComparison
		}
	} else {
		switch k1 {
		case boolKind, complexKind:
			return false, errBadComparisonType
		case floatKind:
			truth = v1.Float() < v2.Float()
		case intKind:
			truth = v1.Int() < v2.Int()
		case stringKind:
			truth = v1.String() < v2.String()
		case uintKind:
			truth = v1.Uint() < v2.Uint()
		default:
			return false, errors.New("invalid kind")
		}
	}
	return truth, nil
}

// le evaluates the comparison <= b.
func lte(arg1, arg2 interface{}) (bool, error) {
	// <= is < or ==.
	lessThan, err := lt(arg1, arg2)
	if lessThan || err != nil {
		return lessThan, err
	}
	return eq(arg1, arg2)
}

func gt(arg1, arg2 interface{}) (bool, error) {
	// > is the inverse of <=.
	lessOrEqual, err := lte(arg1, arg2)
	if err != nil {
		return false, err
	}
	return !lessOrEqual, nil
}

func gte(arg1, arg2 interface{}) (bool, error) {
	// >= is the inverse of <.
	lessThan, err := lt(arg1, arg2)
	if err != nil {
		return false, err
	}
	return !lessThan, nil
}

var funcMap = template.FuncMap{
	"add":       add,
	"sub":       sub,
	"mult":      mult,
	"div":       div,
	"lt":        lt,
	"gt":        gt,
	"lte":       lte,
	"gte":       gte,
	"eq":        eq,
	"neq":       neq,
	"number":    number,
	"currency":  currency,
	"substring": substring,
	"date":      date,
}
