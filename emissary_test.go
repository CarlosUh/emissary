package emissary

import (
	"github.com/maxwellhealth/emissary/delivery"
	"github.com/maxwellhealth/emissary/generator"
	"github.com/maxwellhealth/emissary/middleware"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestEmissary(t *testing.T) {
	Convey("Emissary", t, func() {

		del := &delivery.Mock{}
		gen := &generator.Mock{
			Data: "This is a test",
		}

		mod := &Emissary{
			DeliveryModule: del,
			Middleware:     []middleware.Module{&middleware.Reverse{}},
			Generator:      gen,
			Schedules:      []string{"* * * * *"},
		}

		Convey("Run", func() {
			err := mod.Run()
			So(err, ShouldEqual, nil)

			So(string(del.Data), ShouldEqual, "tset a si sihT")
		})

		Convey("ShouldRun", func() {
			shouldRun, err := mod.ShouldRun(time.Now())
			So(err, ShouldEqual, nil)
			So(shouldRun, ShouldEqual, true)
		})

	})
}
