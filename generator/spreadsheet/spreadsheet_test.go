package spreadsheet

import (
	"errors"
	"github.com/maxwellhealth/emissary/data"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	// "time"
)

type testDataSource struct {
	index int
	data  []map[string]interface{}
}

func (t *testDataSource) Next() (data.Getter, error) {
	if t.index >= len(t.data) {
		return &data.Datum{}, errors.New("No data remaining")
	}
	ret := &data.Datum{t.data[t.index]}
	t.index++
	return ret, nil
}

func (t *testDataSource) HasNext() bool {
	if t.index < len(t.data) {
		return true
	}
	return false
}

type sliceWriter struct {
	data []byte
}

func (s *sliceWriter) Write(p []byte) (int, error) {
	s.data = append(s.data, p...)
	return len(p), nil
}

func dataSourceFromSlice(data []map[string]interface{}) data.DataSource {
	return &testDataSource{0, data}
}

func TestSpreadsheet(t *testing.T) {
	Convey("Spreadsheet Generator", t, func() {
		dataSource := dataSourceFromSlice([]map[string]interface{}{
			map[string]interface{}{
				"a": "this is a",
				"b": "this has a , comma",
				"c": 5,
			},
			map[string]interface{}{
				"a": "foo",
				"b": "bar",
				"c": 10,
			},
		})

		writer := &sliceWriter{}

		s := &SpreadsheetGenerator{
			DataSource: dataSource,
			Columns: []Column{
				Column{
					Value:      "{{substring .a 2}}",
					FixedWidth: 5,
				},
				Column{
					Value:      "{{add .c 10}}",
					FixedWidth: 3,
				},
				Column{
					Value:      "{{.b}}",
					FixedWidth: 2,
				},
			},
		}

		Convey("CSV", func() {
			err := s.Generate(writer)
			So(err, ShouldEqual, nil)

			So(string(writer.data), ShouldEqual, "th,15,\"this has a , comma\"\nfo,20,bar\n")
		})

		Convey("TSV", func() {
			s.Format = FORMAT_TSV
			err := s.Generate(writer)
			So(err, ShouldEqual, nil)

			So(string(writer.data), ShouldEqual, "th\t15\tthis has a , comma\nfo\t20\tbar\n")
		})

		Convey("PSV", func() {
			s.Format = FORMAT_PSV
			err := s.Generate(writer)
			So(err, ShouldEqual, nil)

			So(string(writer.data), ShouldEqual, "th|15|this has a , comma\nfo|20|bar\n")
		})

		Convey("Fixed Width", func() {
			s.Format = FORMAT_FIXED_WIDTH
			err := s.Generate(writer)
			So(err, ShouldEqual, nil)
			So(string(writer.data), ShouldEqual, "th   15 th\nfo   20 ba\n")
		})

		Convey("With footer aggregations", func() {
			s.Columns[1].Footer = "{{number .mean 2}} {{number .median 0}}"
			s.ShowColumnFooters = true
			err := s.Generate(writer)
			So(err, ShouldEqual, nil)
			So(string(writer.data), ShouldEqual, "th,15,\"this has a , comma\"\nfo,20,bar\n,17.50 15,\n")
		})

	})
}
