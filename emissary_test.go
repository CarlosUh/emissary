package emissary

import (
	"github.com/maxwellhealth/emissary/delivery"
	"github.com/maxwellhealth/emissary/generator"
	"github.com/maxwellhealth/emissary/security"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestEmissary(t *testing.T) {
	Convey("Emissary", t, func() {
		mod := &Emissary{
			DeliveryModule: delivery.Noop,
			SecurityModule: security.Noop,
			Generator:      generator.Noop,
			Schedules:      []string{"* * * * *"},
		}

		Convey("Run", func() {
			err := mod.Run()
			So(err, ShouldEqual, nil)
		})

		Convey("ShouldRun", func() {
			shouldRun, err := mod.ShouldRun(time.Now())
			So(err, ShouldEqual, nil)
			So(shouldRun, ShouldEqual, true)
		})

	})
}
