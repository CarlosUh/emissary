Emissary is a framework that facilitates flexible and fully configurable **electronic data interchange** (EDI) in a modular way, allowing non-engineers to implement complex file generation, security, and delivery strategies. 

We use this at Maxwell Health to translate and format data from our system into the various files that our carrier and integration partners can read. But Emissary is abstract enough that it can be used for any sort of "us-to-them" data delivery use case.

Emissary is not a standalone application. Rather, it should be part of a larger application and wrapped to fit your use case.

# Core Concepts
## The "Emissary"
A single Emissary (`*emissary.Emissary`) consists of configurations and modules that are used for file generation, middleware (handling, formatting, security, etc), and delivery. It has but two methods: `Run() error` will generate the data using the generator, pass the data through any middleware modules, and deliver the data using the delivery module. `ShouldRun(time.Time) (bool, error)` parses the emissary's `Schedules` (cron syntax) to decide if it should be run at the provided time.

## Generator
A generator is any type that implements the `generator.FileGenerator` interface, which has one method: `Generate(io.Writer) error`. Any generator can make use of Emissary's `DataSource` to retrieve individual `DataMap`s, which implement a highly flexible syntax for data retrieval from arbitrary `map[string]interface{}`s. (see below)

## Middleware
A middleware module takes an `io.Reader`, which reads from the file generated by the `Generator`, and writes back to an `io.Writer`. You can use this to, for example, encrypt the file (PGP?) before passing it to the delivery module, or maybe store it somewhere on your file system in addition to delivering it somewhere. Check out the "reverse" middleware for a (stupid) example.

## Delivery Module
A delivery module takes an `io.Reader`, which reads from the generated file, and then performs an abstracted task (SFTP file drop, email, etc).

# EDL (Emissary Data Language)
This is a wrapper for go's `html/template`, with some added and modified functions for formatting, math, and comparisons. Specifically meant for non-engineers to format and/or manipulate the result when retrieving a value from the `DataSource`. The `Get` method of the `Datum` takes an EDL **key**, which is interpreted and then used to retrieve and format a value or values from the map. How you use the EDL is up to you (we let members of our services team configure each column of a spreadsheet report with EDL, which is passed verbatim to `Datum.Get()`).

For a full list of examples, please see the test file (`/data/datum_test.go`).

Note that the source of a datum can be only a map or a struct (or pointer to a struct). If it's a struct, it is first converted to a map before being run through the template engine, because if it's a struct then missing fields will panic instead of being left blank. You can specify the struct tag used to determine the translated map keys as the second argument in `datum.SetSource`.

## Dynamic Values
To get the value of an element with dot notation or do any sort of dynamic calculations or formatting, you need to wrap the key with double curly braces, a la Handlebars. To access a property of the data source, put a dot in front of the property name. For example, if you have a map that looks like this:

```go
datum := emissary.Datum{}
datum.SetSource(map[string]interface{}{
	"foo":"bar",
	"baz":map[string]interface{}{
		"bing":"boop",
	},
}, "")
```

You can retrieve the value of `baz.bing` with `datum.Get("{{.baz.bing}}")`

Note that even if the retrieved value is not a string, the result will be the string-ified version of that value. Dates, for example, will end up as their RFC3339 format unless formatted otherwise.

## Functions
You can use functions in the markup to do comparisons, formatting, and mathematical operations. The syntax is very similar to lisp - see below for examples.

### Formatters
To use the built-in currency formatter:

```go
datum := emissary.Datum{}
datum.SetSource(map[string]interface{}{
	"cost":1000,
}, "")

formattedCost := datum.Get("{{currency .cost}}") 
// "$1,000.00"
```

Some formatters take arguments:

```go
datum := emissary.Datum{}
datum.SetSource(map[string]interface{}{
	"date":time.Now,
}, "")

formattedDate := datum.Get("{{(date .date "2006-01-02")}}") 
// "2015-03-21", at the time of this writing
```

#### Full list of Formatters
* currency - adds commas, dollar sign prefix, and two decimal places w/ rounding
* substring - negative argument takes `n` characters from end of string, positive from beginning of string
* date - formats a date with first argument. See `time.Format`
* number - rounds number to n (second argument, after the number itself) decimal places

### Comparators
You can use the comparators in conjunction with if/else statements:

```go
datum := emissary.Datum{}
datum.SetSource(map[string]interface{}{
	"cost":1000,
})

result := datum.Get("{{if (gt .cost 900)}}big{{else}}small{{end}}") 
// "big"
```

Note that dates can also be compared, unlike in go's `html/template`.

#### Full list of Comparators
* eq
* neq
* gt
* gte
* lt
* lte

### Mathematical Operators
Mathematical operators can be run on numbers (floats or ints):

```go
datum := emissary.Datum{}
datum.SetSource(map[string]interface{}{
	"cost":"1000",
}, "")

result := datum.Get("{{add .cost 500}}") 
// "1500"
```

You can also use them in combination:

```go
datum := emissary.Datum{}
datum.SetSource(map[string]interface{}{
	"cost":"1000",
}, "")

result := datum.Get("{{add .cost (mult 500 2)}}") 
// "1500"
```

#### Full list of Mathematical Operators
* add
* sub
* mult
* div

