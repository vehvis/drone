package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl32"

	"deus.solita.fi/Solita/projects/drone_code_camp/repositories/git/ddr.git"
	"gobot.io/x/gobot/platforms/keyboard"
	"gocv.io/x/gocv"
)

func main() {

	track := ddr.NewTrack()
	defer track.Close()

	inflight := false

	window := gocv.NewWindow("Drone")
	drone := ddr.NewDrone(ddr.DroneReal, "../drone-camera-calibration-720.yaml")
	//drone := ddr.NewDrone(ddr.DroneFake, "../camera-calibration.yaml")
	err := drone.Init()

	if err != nil {
		fmt.Printf("error while initializing drone: %v\n", err)
		return
	}

	targetVec := mgl32.Vec3{0, 0, 0}
	
	flightStatus := ""

	var tvArr []mgl32.Vec3

	var center []bool

	currentRing := 0

	for {
		frame := <-drone.VideoStream()

		gocv.PutText(&frame, flightStatus,
			image.Pt(50, 50),
			gocv.FontHersheySimplex, 0.4, color.RGBA{255, 255, 0, 0}, 2)

		// detect markers in this frame
		markers := track.GetMarkers(&frame)

		rings := track.ExtractRings(markers)
		for id, ring := range rings {
			if id == currentRing {
				pose := ring.EstimatePose(drone)
				ring.Draw(&frame, pose, drone)
				_ = id
			}
		}

		going := int64(0)

		if ring, ok := rings[currentRing]; ok {

			pose := ring.EstimatePose(drone)

			targetVec = drone.CameraToDroneMatrix().Mul3x1(pose.Position)
			ringFace := pose.Rotation.Mul3x1(mgl32.Vec3{0.0, 0.0, 1.0})

			tvArr = append(tvArr, targetVec)
			if len(tvArr) > 3 {
				tvArr = tvArr[1:]
			}
			foo := avg(tvArr)

			x := foo[0]
			y := foo[1]
			z := foo[2]

			xTrans := translate(x, 0)
			yTrans := translate(y, 0.10)
			zTrans := translate(z, 0.8)

			xRot := rotate(ringFace[0])

			xOk := xTrans == 0
			yOk := yTrans == 0
			zOk := zTrans == 0
			rotOk := xRot == 0

			isOk := false

			if going == 0 {
				drone.Right(xTrans)
				drone.Down(yTrans)
				drone.Forward(zTrans)

				if xRot == 0 {
					drone.CeaseRotation()
				} else {
					drone.CounterClockwise(xRot)
				}

				if xOk && yOk && zOk && rotOk {
					drone.Hover()
					drone.CeaseRotation()
				}

				center = append(center, xOk && yOk && zOk && rotOk)
				if len(center) > 3 {
					center = center[1:]
				}

				isOk = true
				for _, c := range center {
					if !c {
						isOk = false
					}
				}

				if isOk {
					going = makeTimestamp()
					fmt.Printf("Going for %d", currentRing)
					drone.Hover()
				}
			} else {
				if makeTimestamp()-going < 3000 {
					drone.Forward(60)
				} else {
					currentRing = currentRing + 1
					going = 0
					fmt.Printf("Next ring %d", currentRing)
				}
			}

			flightStatus = fmt.Sprintf("X=%f Y=%f Z=%f, OK=%t, xt=%d yt=%d, xr=%f, GOING=%t", foo[0], foo[1], foo[2], isOk, xTrans, yTrans, ringFace[0], going > 0)

		} else if going == 0 {
			drone.Hover()
			drone.CeaseRotation()
		}
		ddr.DrawHorizon(drone, &frame)
		ddr.DrawControls(drone, &frame)
		window.IMShow(frame)
		key := window.WaitKey(1)

		switch key {
		case keyboard.Spacebar: // space
			if inflight {
				drone.Land()
				fmt.Print("Land!\n")
			} else {
				drone.TakeOff()
				fmt.Print("Takeoff!\n")
			}
			inflight = !inflight
		case keyboard.T:
			if inflight {
				fmt.Printf("Turbo: %d", currentRing)
				drone.Forward(70)
				time.Sleep(1 * time.Second)
				currentRing++
				fmt.Printf("Next ring: %d", currentRing)
			}
		}
	}
}

func translate(val float32, target float32) int {
	ret := math.Min(float64((val-target)*50), 15)

	if math.Abs(ret) < 7 {
		return 0
	}
	return int(ret)
}

func rotate(val float32) int {
	ret := math.Min(float64(val*50), 30)

	if math.Abs(ret) < 10 {
		return 0
	}
	return int(ret)
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func avg(arr []mgl32.Vec3) mgl32.Vec3 {
	sum := mgl32.Vec3{0, 0, 0}

	for _, vec := range arr {
		sum = sum.Add(vec)
		//fmt.Printf("%f, %f, %f", sum[0], sum[1], sum[2])
	}

	sum = sum.Mul(1.0 / float32(len(arr)))
	return sum
}
