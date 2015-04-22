package data

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

var dm = map[string]interface{}{
	"a": 10,
	"b": map[string]interface{}{
		"c": 2.5,
	},
	"d": 2,
	"e": "bar",
	"f": map[string]interface{}{
		"g": "boop",
	},
	"h": time.Now(),
	"i": "Numeric String 2",
	"j": true,
	"k": false,
	"l": time.Now().Add(10 * time.Minute),
}

type a struct {
	key         string
	expectation string
}

var now = time.Now()
var assertions = []a{
	a{"{{.e}}", "bar"},
	a{"e", "e"},
	a{"{{.f.g}}", "boop"},
	a{"{{.boof.bloop}}", ""},
	a{"{{.e}} {{.f.g}}", "bar boop"},
	a{"{{(add (mult .a .b.c) .d 3)}}", "30"},
	a{"{{(currency 3000.5000)}}", "$3,000.50"},
	a{"{{(substring \"foobarbaz\" -3)}}", "baz"},
	a{"{{(substring .f.g 2)}} {{(currency .a)}}", "bo $10.00"},
	a{"{{if (eq .a 10)}}yes{{else}}no{{end}}", "yes"},
	a{"{{if (gt .a 9)}}{{(substring .e 2)}}{{else}}no{{end}}", "ba"},
	a{"{{if (gt .a 11)}}{{(substring .e 2)}}{{else if (gt .a 9)}}this one{{else}}last one{{end}}", "this one"},
	a{"{{if (gt .a 11)}}{{(substring .e 2)}}{{else if (gt .a 13)}}this one{{else}}last one{{end}}", "last one"},
	a{"some arbitrary text {{if (gt .a 11)}}{{(substring .e 2)}}{{else if (gt .a 13)}}this one{{else}}last one{{end}}", "some arbitrary text last one"},
	a{"{{date .h \"2006\"}}", now.Format("2006")},
	a{"{{date .h}}", now.Format("02/01/2006")},
	a{"{{date .a}}", "Invalid Date"},
	a{"{{.j}}", "true"},
	a{"{{.k}}", "false"},
	a{"{{if (gt .h .l)}}h{{else}}l{{end}}", "l"},
}

type c struct {
	block       string
	expectation bool
}

var conditionAssertions = []c{
	c{"[if eq 5 5]", true},
	c{"[if eq 5.0 5]", true},
	c{"[if eq foo bar]", false},
	c{"[if eq \"2000-01-01 00:00:00\" \"2000-01-01 00:00:00\"]", true},
	c{"[if eq \"2001-01-01 00:00:00\" \"2000-01-01 00:00:00\"]", false},
	c{"[if ne 5 5]", false},
	c{"[if ne 5 6]", true},
	c{"[if gt 5 4]", true},
	c{"[if gt \"2001-01-01 00:00:00\" \"2000-01-01 00:00:00\"]", true},
	c{"[if gt \"2000-01-01 00:00:00\" \"2000-01-01 00:00:00\"]", false},
	c{"[if gte \"2000-01-01 00:00:00\" \"2000-01-01 00:00:00\"]", true},
	c{"[if gte \"2000-01-01 00:00:00\" \"2001-01-01 00:00:00\"]", false},
	c{"[if gte 5 5.000]", true},
	c{"[if lt \"2000-01-01 00:00:00\" \"2001-01-01 00:00:00\"]", true},
	c{"[if lt \"2000-01-01 00:00:00\" \"2000-01-01 00:00:00\"]", false},
	c{"[if lte \"2000-01-01 00:00:00\" \"2000-01-01 00:00:00\"]", true},
	c{"[if lte \"2001-01-01 00:00:00\" \"2000-01-01 00:00:00\"]", false},
	c{"[if lte 5 5.000]", true},
	c{"[if in 5 6 7 8 9 5]", true},
	c{"[if in a b c d e f]", false},
	c{"[if nin 5 6 7 8 9 5]", false},
	c{"[if nin a b c d e f]", true},
}

func TestDataMap(t *testing.T) {
	datum := &Datum{}
	datum.SetSource(dm, "")

	Convey("DataMap", t, func() {
		for i, assertion := range assertions {
			Convey(fmt.Sprintf("Assertion #%d :: %s => %s", i+1, assertion.key, assertion.expectation), func() {
				So(datum.Get(assertion.key, ""), ShouldEqual, assertion.expectation)
			})
		}

	})

	// Convey("Conditional block evaluation", t, func() {
	// 	for i, assertion := range conditionAssertions {
	// 		Convey(fmt.Sprintf("Assertion #%d :: (%s => %t)", i+1, assertion.block, assertion.expectation), func() {
	// 			So(datum.evaluateBlock(assertion.block), ShouldEqual, assertion.expectation)
	// 		})
	// 	}
	// })
}
