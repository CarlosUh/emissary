// Generates a spreadsheet from a datasource that returns rows of data
//
// Can be in multipel formats (see format constants)
//
// Aggregation is also available. Check out tests for a full run-down
// on what you can do

package spreadsheet

import (
	"encoding/csv"
	"errors"
	"github.com/maxwellhealth/emissary/data"
	"github.com/maxwellhealth/go-floatlist"
	"io"
	"strconv"
	"strings"
)

const (
	FORMAT_CSV         = iota
	FORMAT_TSV         = iota
	FORMAT_PSV         = iota
	FORMAT_FIXED_WIDTH = iota
)

type Column struct {
	Header     string
	Value      string
	Default    string
	Footer     string
	FixedWidth int
}

type SpreadsheetGenerator struct {
	Columns    []Column
	DataSource data.DataSource
	// Configuration options
	ShowColumnHeaders bool
	ShowColumnFooters bool
	Format            int

	csvWriter *csv.Writer
	writer    io.Writer
}

func (s *SpreadsheetGenerator) Generate(writer io.Writer) error {
	var csvWriter *csv.Writer
	if s.Format != FORMAT_FIXED_WIDTH {
		// Make the writer based on the format
		csvWriter = csv.NewWriter(writer)
		switch s.Format {
		case FORMAT_CSV:
			csvWriter.Comma = ','
		case FORMAT_TSV:
			csvWriter.Comma = '\t'
		case FORMAT_PSV:
			csvWriter.Comma = '|'
		}
	}

	s.csvWriter = csvWriter
	s.writer = writer

	if s.ShowColumnHeaders {
		headers := make([]string, len(s.Columns))
		for i, c := range s.Columns {
			headers[i] = c.Header
		}

		err := s.writeRow(headers)
		if err != nil {
			return err
		}
	}

	// Figure out which columns we need to keep track of to get aggregations on the footer. The value spit out by the getter for that column must be parseable as a float for it to be added to the aggregations.
	totalRows := 0

	columnsToTrack := make(map[int][]float64)

	for i, c := range s.Columns {
		// If it has $sum, $mean, $median, $mode, $totalUnempty, or $totalEmpty, we need to track it so we can show it in the footer
		if strings.Contains(c.Footer, ".mean") || strings.Contains(c.Footer, ".median") || strings.Contains(c.Footer, ".sum") || strings.Contains(c.Footer, ".mode") || strings.Contains(c.Footer, ".totalUnempty") || strings.Contains(c.Footer, ".totalEmpty") {
			columnsToTrack[i] = []float64{}
		}
	}

	for s.DataSource.HasNext() {
		next, err := s.DataSource.Next()

		if err != nil {
			return err
		}
		totalRows++

		row := make([]string, len(s.Columns))

		for i, c := range s.Columns {
			val := next.Get(c.Value, c.Default)
			row[i] = val

			// Do we need to keep track of it?
			if colval, ok := columnsToTrack[i]; ok {
				if len(val) > 0 {
					// If it's numerical, add the number. Otherwise add 1.
					f, err := strconv.ParseFloat(val, 64)
					if err != nil {
						columnsToTrack[i] = append(colval, 1.0)
					} else {
						columnsToTrack[i] = append(colval, f)
					}
				} else {
					columnsToTrack[i] = append(colval, 0.0)
				}

			}
		}

		err = s.writeRow(row)
	}

	if s.ShowColumnFooters {
		footer := make([]string, len(s.Columns))
		for i, c := range s.Columns {
			if aggregate, ok := columnsToTrack[i]; ok {
				// Make the different aggregate values
				list := floatlist.Floatlist(aggregate)
				agg := &footerAggregation{
					Sum:          list.Sum(),
					Mean:         list.Mean(),
					Median:       list.Median(),
					Mode:         list.Mode(),
					TotalUnempty: list.GetCountByValue(1.0),
					TotalEmpty:   list.GetCountByValue(0.0),
				}

				datum := &data.Datum{}
				datum.SetSource(agg, "mapTo")
				footer[i] = datum.Get(c.Footer, "")

			} else {
				footer[i] = ""
			}
		}

		s.writeRow(footer)
	}
	return nil
}

type footerAggregation struct {
	Sum          float64 `mapTo:"sum"`
	Mean         float64 `mapTo:"mean"`
	Median       float64 `mapTo:"median"`
	Mode         float64 `mapTo:"mode"`
	TotalEmpty   int     `mapTo:"totalEmpty"`
	TotalUnempty int     `mapTo:"totalUnempty"`
}

func (s *SpreadsheetGenerator) writeRow(row []string) error {
	if len(row) != len(s.Columns) {
		return errors.New("Failed to write row - length of row does not match # of columns")
	}

	var err error
	if s.Format == FORMAT_FIXED_WIDTH {
		for i, c := range s.Columns {
			l := len(row[i])
			val := row[i]
			if l < c.FixedWidth {
				for x := l; x < c.FixedWidth; x++ {
					val = val + " "
				}
			} else if l > c.FixedWidth {
				val = val[:c.FixedWidth]
			}

			row[i] = val
		}

		data := strings.Join(row, "") + "\n"
		_, err = s.writer.Write([]byte(data))
		if err != nil {
			return err
		}
	} else {
		err = s.csvWriter.Write(row)
		if err != nil {
			return err
		}
		s.csvWriter.Flush()
	}

	return nil
}
