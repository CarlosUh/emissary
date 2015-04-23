package data

import (
	"bytes"
	"github.com/fatih/structs"
	"html/template"
	"reflect"
	"time"
)

var DefaultDecimalCount = 8

var TimeFormats = []string{
	"2006-01-02 15:04:05",
	time.RFC3339,
	time.RFC1123,
	time.RFC3339Nano,
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
