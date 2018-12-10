package main

import (
	"fmt"
	"image"
	"image/color"
	"time"

	ddr "deus.solita.fi/Solita/projects/drone_code_camp/repositories/git/ddr.git"
	"gobot.io/x/gobot/platforms/keyboard"
	"gocv.io/x/gocv"
)

func main() {

	inflight := false

	window := gocv.NewWindow("Drone")
	drone := ddr.NewDrone(ddr.DroneReal, "../drone-camera-calibration-720.yaml")
	//drone := ddr.NewDrone(ddr.DroneFake, "../camera-calibration.yaml")
	err := drone.Init()

	if err != nil {
		fmt.Printf("error while initializing drone: %v\n", err)
		return
	}

	flightStatus := "Initialized"

	for {
		frame := <-drone.VideoStream()

		gocv.PutText(&frame, flightStatus,
			image.Pt(50, 50),
			gocv.FontHersheySimplex, 0.8, color.RGBA{0, 0, 0, 0}, 2)

		window.IMShow(frame)
		key := window.WaitKey(1)

		switch key {
		case keyboard.Spacebar: // space
			if inflight {
				drone.Land()
				flightStatus = "Taken off"
			} else {
				drone.TakeOff()
				flightStatus = "Landed"
			}
			inflight = !inflight
		case keyboard.F:
			if inflight {
				flightStatus = "In flight"
				go fly(drone)
			}
		}

	}

}

func fly(drone ddr.Drone) {
	drone.Forward(10)
	time.Sleep(5 * time.Second)
	drone.Hover()
}
