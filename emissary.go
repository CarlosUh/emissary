package emissary

import (
	"bytes"
	"github.com/gorhill/cronexpr"
	"github.com/maxwellhealth/emissary/delivery"
	"github.com/maxwellhealth/emissary/generator"
	"github.com/maxwellhealth/emissary/middleware"
	"time"
)

type Emissary struct {
	Name           string
	DeliveryModule delivery.Module
	Middleware     []middleware.Module
	FileName       string
	Schedules      []string
	Generator      generator.FileGenerator
}

func (e *Emissary) Run() error {
	generated := new(bytes.Buffer)

	// Generate the file
	err := e.Generator.Generate(generated)
	if err != nil {
		return err
	}

	// Middleware?
	for _, m := range e.Middleware {
		buf := new(bytes.Buffer)
		err := m.Passthru(generated, buf)
		if err != nil {
			return err
		}

		generated = buf

	}

	err = e.DeliveryModule.Deliver(generated)
	if err != nil {
		return err
	}

	return nil

}

func (e *Emissary) ShouldRun(t time.Time) (bool, error) {
	t = t.Truncate(time.Minute)
	compare := t.Add(-1 * time.Nanosecond)

	for _, s := range e.Schedules {
		parsed, err := cronexpr.Parse(s)
		if err != nil {
			return false, err
		}
		next := parsed.Next(compare)
		if next == t {
			return true, nil
		}
	}
	return false, nil
}
