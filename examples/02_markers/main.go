package main

import (
	"fmt"
	"image/color"

	"deus.solita.fi/Solita/projects/drone_code_camp/repositories/git/ddr.git"
	"gobot.io/x/gobot/platforms/keyboard"
	"gocv.io/x/gocv"
)

func main() {

	track := ddr.NewTrack()
	defer track.Close()

	inflight := false

	window := gocv.NewWindow("Drone")
	//drone := ddr.NewDrone(ddr.DroneReal, "../drone-camera-calibration-400.yaml")
	drone := ddr.NewDrone(ddr.DroneFake, "../camera-calibration.yaml")
	err := drone.Init()

	if err != nil {
		fmt.Printf("error while initializing drone: %v\n", err)
		return
	}

	for {
		frame := <-drone.VideoStream()

		// detect markers in this frame
		markers := track.GetMarkers(&frame)

		// loop through markers and setup tracking for markers in ring 0
		for _, m := range markers {
			if m.RingID() == 0 && !m.IsTracking() {
				m.SetupTracking(&frame)
			}
			if m.RingID() != 0 && m.IsTracking() {
				m.StopTracking()
			}

			if m.Detected {
				m.DrawBorder(&frame, color.RGBA{0, 255, 0, 0})
				//m.DrawCenter(&frame, color.RGBA{0,255,0,0})
			} else {
				m.DrawBorder(&frame, color.RGBA{0, 0, 255, 0})
				//m.DrawCenter(&frame, color.RGBA{0,0,255,0})
			}
		}

		window.IMShow(frame)
		key := window.WaitKey(1)

		switch key {
		case keyboard.Spacebar: // space
			if inflight {
				drone.Land()
			} else {
				drone.TakeOff()
			}
			inflight = !inflight
		}

	}

}
