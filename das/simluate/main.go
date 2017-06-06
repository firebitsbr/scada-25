package main

import (
	"flag"
	"fmt"
	"time"
	"wj/simluate"
	"wj/udp2op"
)

var (
	buildDate string
)

func version() {
	if buildDate != "" {
		fmt.Println("Build date:", buildDate)
	}
}

var paramBegin = flag.Int("begin", 1024, "The ID of begin.")
var paramEnd = flag.Int("end", 2024, "The ID of end.")
var paramHost = flag.String("host", "127.0.0.1:8200", "The address of server.")

type Point struct {
	ID     int
	Type   int
	BV     float64
	TV     float64
	Value  float64
	IsUp   bool
	Status int16
	Time   int64
	Count  int
}

func main() {
	flag.Parse()
	version()
	points := make([]*Point, 0)
	udp2op.SetHost(*paramHost)
	for i := *paramBegin; i < *paramEnd; i++ {
		point := new(Point)
		point.Type = 0
		point.ID = i
		point.BV = float64((simluate.GetRand()%499 + 1) % (simluate.GetRand()%499 + 1))
		point.TV = point.BV + point.BV*(float64(simluate.GetRand()%3+1)/10)
		point.IsUp = simluate.GetRand()%2 == 0
		point.Value = point.BV + (point.TV-point.BV)/2
		points = append(points, point)
	}
	for {
		t := time.Now().Unix()
		for _, point := range points {
			point.Count--
			if point.Count > 0 {
				continue
			}
			point.Time = t
			point.Value, point.IsUp = simluate.MakeValue(point.BV, point.TV, point.Value, point.IsUp)
			udp2op.PushValue(point.ID, point.Type, point.Value, point.Status, point.Time)
			point.Count = int(simluate.GetRand() % 5)
		}
		time.Sleep(time.Second)
	}
}
