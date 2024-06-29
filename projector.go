package main

import "math"

type Projector interface {
	corner(i, j int) (float64, float64, float64)
}

type SinProjector struct{}

func (SinProjector) corner(i, j int) (float64, float64, float64) {
	x, y := corner(i, j)
	r := math.Hypot(x, y) // distance from (0,0)
	z := math.Sin(r) / r
	return x, y, z
}

type EggboxProjector struct{}

func (EggboxProjector) corner(i, j int) (float64, float64, float64) {
	x, y := corner(i, j)
	r := 10.0
	z := (math.Sin(x) + math.Sin(y)) / r
	return x, y, z
}

type MogulsProjector struct{}

func (MogulsProjector) corner(i, j int) (float64, float64, float64) {
	x, y := corner(i, j)
	a := 0.01
	b := 0.01
	q := (2 * math.Pi) / 4.0
	p := (2 * math.Pi) / 10.0
	z := -a*x - b*math.Cos(p*x)*math.Cos(q*y)
	return x, y, z
}

type SaddleProjector struct{}

func (SaddleProjector) corner(i, j int) (float64, float64, float64) {
	x, y := corner(i, j)
	a := 0.1
	b := 0.05
	z := math.Pow(a*x, 2) - math.Pow(b*y, 2)
	return x, y, z
}

// Find point (x,y) at corner of cell (i,j).
func corner(i, j int) (float64, float64) {
	x := xyrange * (float64(i)/cells - 0.5)
	y := xyrange * (float64(j)/cells - 0.5)
	return x, y
}

// Project (x,y,z) isometrically onto 2-D SVG canvas (sx,sy).
func project(x, y, z float64) (float64, float64) {
	sx := width/2 + (x-y)*cos30*xyscale
	sy := height/2 + (x+y)*sin30*xyscale - z*zscale
	return sx, sy
}
