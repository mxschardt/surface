// Surface computes an SVG rendering of a 3-D surface function.
package main

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
)

const (
	width, height = 600, 320            // canvas size in pixels
	cells         = 100                 // number of grid cells
	xyrange       = 30.0                // axis ranges (-xyrange..+xyrange)
	xyscale       = width / 2 / xyrange // pixels per x or y unit
	zscale        = height * 0.4        // pixels per z unit
	angle         = math.Pi / 6         // angle of x, y axes (=30°)
)

var sin30, cos30 = math.Sin(angle), math.Cos(angle) // sin(30°), cos(30°)

func main() {
	http.HandleFunc("/", handler) // eapeakColor request calls handler
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

// handler epeakColoroes the Path component of the request URL r.
func handler(w http.ResponseWriter, r *http.Request) {
	var err error
	var projector Projector = SinProjector{}
	height, width := height, width
	peakColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	valleyColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	if projectorStr := r.URL.Query().Get("function"); projectorStr != "" {
		switch projectorStr {
		case "sin":
		case "eggbox":
			projector = EggboxProjector{}
		case "moguls":
			projector = MogulsProjector{}
		case "saddle":
			projector = SaddleProjector{}
		default:
			http.Error(w, errorf("error: unknown value 'function'=%q", projectorStr), http.StatusBadRequest)
			return
		}

	}
	if heightStr := r.URL.Query().Get("height"); heightStr != "" {
		height, err = strconv.Atoi(heightStr)
		if err != nil {
			http.Error(w, errorf("cannot parse 'height'=%q to float", height), http.StatusBadRequest)
			return
		}
	}
	if widthStr := r.URL.Query().Get("width"); widthStr != "" {
		width, err = strconv.Atoi(widthStr)
		if err != nil {
			http.Error(w, errorf("cannot parse 'width' %q to float", width), http.StatusBadRequest)
			return
		}
	}
	if colorStr := r.URL.Query().Get("valley"); colorStr != "" {
		valleyColor, err = hexToRGBA(colorStr)
		if err != nil {
			http.Error(w, errorf("cannot parse 'valley' %q to RGBA", err), http.StatusBadRequest)
			return
		}
	}
	if colorStr := r.URL.Query().Get("peak"); colorStr != "" {
		peakColor, err = hexToRGBA(colorStr)
		if err != nil {
			http.Error(w, errorf("cannot parse 'peak' %q to RGBA", colorStr), http.StatusBadRequest)
			return
		}
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	svg(w, projector, peakColor, valleyColor)
}

func errorf(format string, a ...any) string {
	return fmt.Sprintf("error: "+format, a)
}

func svg(w io.Writer, p Projector, peakColor, valleyColor color.RGBA) {
	fmt.Fprintf(w, "<svg xmlns='http://www.w3.org/2000/svg' "+
		"style='stroke: grey; fill: white; stroke-width: 0.7' "+
		"width='%d' height='%d'>", width, height)

	surface(w, p, peakColor, valleyColor)
	fmt.Fprint(w, "</svg>")
}

func surface(out io.Writer, p Projector, peakColor, valleyColor color.RGBA) {
	const polygonf string = "<polygon points='%s' fill='%s'/>\n"
	var zmax, zmin float64 = math.Inf(-1), math.Inf(1)
	var polygons [cells][cells][9]float64

	for i := 0; i < cells; i++ {
		for j := 0; j < cells; j++ {
			ax, ay, az := p.corner(i+1, j)
			bx, by, bz := p.corner(i, j)
			cx, cy, cz := p.corner(i, j+1)
			dx, dy, dz := p.corner(i+1, j+1)
			// Skip polygon if value is NaN or Inf.
			if err := az + bz + cz + dz; math.IsNaN(err) || math.IsInf(err, 0) {
				continue
			}

			ax, ay = project(ax, ay, az)
			bx, by = project(bx, by, bz)
			cx, cy = project(cx, cy, cz)
			dx, dy = project(dx, dy, dz)
			z := average(az, bz, cz, dz)
			polygons[i][j] = [9]float64{z, ax, ay, bx, by, cx, cy, dx, dy}

			zmax = max(zmax, az, bz, cz, dz)
			zmin = min(zmin, az, bz, cz, dz)

		}
	}

	for i := 0; i < cells; i++ {
		for j := 0; j < cells; j++ {
			var points strings.Builder
			for i, p := range polygons[i][j][1:] {
				points.WriteString(strconv.FormatFloat(p, 'f', 6, 64))
				if i != len(polygons[i][j][1:])-1 {
					points.WriteString(", ")
				}
			}
			z := polygons[i][j][0]
			c := zcolor(z, zmax, zmin, valleyColor, peakColor)

			fmt.Fprintf(out, polygonf, points.String(), fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B))
		}
	}
}

func zcolor(z, zmax, zmin float64, high, low color.RGBA) color.RGBA {
	percent := percent(zmin, zmax, z)
	currentColor := color.RGBA{
		R: interpolate(high.R, low.R, percent),
		G: interpolate(high.G, low.G, percent),
		B: interpolate(high.B, low.B, percent),
		A: high.A,
	}
	return currentColor
}

func percent(cmin, cmax, c float64) float64 {
	return (c - cmin) / (cmax - cmin)
}

func interpolate(a, b uint8, x float64) uint8 {
	return uint8(float64(a)*(1-x) + float64(b)*x)
}

func average(a ...float64) float64 {
	sum := 0.0
	for _, b := range a {
		sum += b
	}
	return sum / float64(len(a))
}

func hexToRGBA(hex string) (color.RGBA, error) {
	values, err := strconv.ParseUint(string(hex), 16, 32)

	if err != nil {
		return color.RGBA{}, err
	}

	rgb := color.RGBA{
		R: uint8(values >> 16),
		G: uint8((values >> 8) & 0xFF),
		B: uint8(values & 0xFF),
	}

	return rgb, nil
}
