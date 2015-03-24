package emissary

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/gorhill/cronexpr"
	"github.com/maxwellhealth/emissary/delivery"
	"github.com/maxwellhealth/emissary/generator"
	"github.com/maxwellhealth/emissary/security"
	"os"
	"time"
)

type Emissary struct {
	Name           string
	DeliveryModule delivery.Module
	SecurityModule security.Module
	FileName       string
	Schedules      []string
	Generator      generator.FileGenerator
}

func (e *Emissary) Run() error {
	// Get a random identifier for the temporary file
	buf := make([]byte, 32)
	i, err := rand.Read(buf)
	if i != 32 {
		return errors.New("Failed to generate temporary file")
	}
	if err != nil {
		return err
	}

	fileName := hex.EncodeToString(buf)

	file, err := os.OpenFile("/tmp/"+fileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	if err != nil {
		return err
	}

	defer func() {
		file.Close()
		os.Remove("/tmp/" + fileName)
	}()

	// Generate the file
	err = e.Generator.Generate(file)
	if err != nil {
		return err
	}

	// Secure
	securedFile, err := os.OpenFile("/tmp/"+fileName+"-s", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	if err != nil {
		return err
	}

	defer func() {
		securedFile.Close()
		os.Remove("/tmp/" + fileName + "-s")
	}()

	err = e.SecurityModule.Secure(file, securedFile)
	if err != nil {
		return err
	}

	err = e.DeliveryModule.Deliver(securedFile)
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
