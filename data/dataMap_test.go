package data

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

var dm = DataMap(map[string]interface{}{
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
})

type a struct {
	key         string
	expectation string
}

var now = time.Now()
var assertions = []a{
	a{"$e", "bar"},
	a{"e", "e"},
	a{"$f.g", "boop"},
	a{"$boof.bloop", ""},
	a{"$e $f.g", "bar boop"},
	a{"$a * $b.c + $d + 3", "30.00000000"},
	a{"3000.5000|currency", "$3,000.50"},
	a{"foobarbaz|substring|-3", "baz"},
	a{"foobarbaz|substring|3", "foo"},
	a{"{$f.g|substring|2} {$a|currency}", "bo $10.00"},
	a{"[if eq $a 10]yes[else]no[/if]", "yes"},
	a{"[if gt $a 9]{$e|substring|2}[else]no[/if]", "ba"},
	a{"[if gt $a 11]{$e|substring|2}[else if gt $a 9]this one[else]last one[/if]", "this one"},
	a{"[if gt $a 11]{$e|substring|2}[else if gt $a 13]this one[else]last one[/if]", "last one"},
	a{"some arbitrary text [if gt $a 11]{$e|substring|2}[else if gt $a 13]this one[else]last one[/if]", "some arbitrary text last one"},
	a{"2001-01-01 00:00:00|date|2006", "2001"},
	a{"$h|date|2006", now.Format("2006")},
	a{"$h|date", now.Format("02/01/2006")},
	a{"a|date", "Invalid Date"},
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

	Convey("DataMap", t, func() {
		for i, assertion := range assertions {
			Convey(fmt.Sprintf("Assertion #%d :: (%s => %s)", i+1, assertion.key, assertion.expectation), func() {
				So(dm.Get(assertion.key, ""), ShouldEqual, assertion.expectation)
			})
		}

	})

	Convey("Conditional block evaluation", t, func() {
		for i, assertion := range conditionAssertions {
			Convey(fmt.Sprintf("Assertion #%d :: (%s => %t)", i+1, assertion.block, assertion.expectation), func() {
				So(dm.evaluateBlock(assertion.block), ShouldEqual, assertion.expectation)
			})
		}
	})
}
