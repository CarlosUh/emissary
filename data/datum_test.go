package data

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

var now = time.Now()

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
	"h": now,
	"i": "Numeric String 2",
	"j": true,
	"k": false,
	"l": now.Add(10 * time.Minute),
	"m": now,
}

type a struct {
	key         string
	expectation string
}

var assertions = []a{
	a{"{{.e}}", "bar"},
	a{"e", "e"},
	a{"{{.f.g}}", "boop"},
	a{"{{.boof.bloop}}", ""},
	a{"{{.e}} {{.f.g}}", "bar boop"},
	a{"{{(add (mult .a .b.c) .d 3)}}", "30"},
	a{"{{(sub .d 3)}}", "-1"},
	a{"{{(div .a 2)}}", "5"},
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
	a{"{{.h}}", now.String()},
	a{"{{if (gt .h .l)}}h{{else}}l{{end}}", "l"},
	a{"{{if (eq .h .l)}}h{{else}}l{{end}}", "l"},
	a{"{{if (eq .h .m)}}y{{else}}n{{end}}", "y"},
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
}
