package data

import (
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"math"
	"strconv"
	"strings"
)

type Formatter func(input string, arguments []string) (string, error)

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

var Formatters = map[string]Formatter{
	"currency": func(input string, arguments []string) (string, error) {

		spl := strings.Split(input, ".")
		intval := spl[0]

		// Parse the int
		ival, err := strconv.ParseInt(intval, 10, 64)
		if err != nil {
			return input, err
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
		return "$" + intval + decval, nil
	},
	"substring": func(input string, arguments []string) (string, error) {
		// First arg determines how to parse it
		if len(arguments) == 0 {
			return input, errors.New("substring formatter requires one numeric argument")
		}

		arg := arguments[0]

		parsed, err := strconv.ParseInt(arg, 10, 64)

		if err != nil {
			return input, errors.New("substring formatter requires one numeric argument")
		}

		if parsed < 0 {
			return input[(int64(len(input)) + parsed):], nil
		} else {
			return input[:parsed], nil
		}
	},
	"date": func(input string, arguments []string) (string, error) {
		// Parse the date...
		parsedDate, err := parseDate(input)
		if err != nil {
			return "Invalid Date", err
		}

		var format string
		if len(arguments) >= 1 {
			format = arguments[0]
		} else {
			format = "02/01/2006"
		}

		return parsedDate.Format(format), nil
	},
}
