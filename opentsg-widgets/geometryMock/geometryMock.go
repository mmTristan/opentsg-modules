// Package geometrymock is used by widget test suites for when they use geometry
package geometrymock

import (
	"fmt"
	"image"
	"math/rand"

	"github.com/mrmxf/opentsg-modules/opentsg-core/gridgen"
)

func Mockgeom(randSrc *rand.Rand, w, h int, nameLength int) []gridgen.Segmenter {

	points := Randomspread(randSrc, w, h)
	ns := linearneighbour(points)
	mocksegments := make([]gridgen.Segmenter, len(ns))

	for i, n := range ns {

		trueNeighs := make([]string, len(n.neighbours))
		for j, neigh := range n.neighbours {
			trueNeighs[j] = fmt.Sprintf("%04x%04x", neigh, neigh)
		}

		name := ""
		for j := 0; j < nameLength; j += 4 {
			name += fmt.Sprintf("%04x", i)
		}
		mocksegments[i] = gridgen.Segmenter{ID: name, Shape: n.area, Neighbours: trueNeighs}
	}

	return mocksegments
}

// Randomspread generates random sized rectangles to fill the height and width
func Randomspread(randSrc *rand.Rand, height, width int) []image.Rectangle {
	/*
		go down at random height intervals then width to make squares with several borders
	*/

	direction := 0
	start := image.Point{0, 0}
	var recs []image.Rectangle

	canv := image.Rect(0, 0, width, height)

	for start.In(canv) {
		vectorstart := start

		/*
			if direction mod 2 == 0 then go down
			else go along loop thorugh

		*/

		if direction%2 == 0 {
			width := 20 + randSrc.Intn(50)
			for vectorstart.In(canv) {
				height := 20 + randSrc.Intn(50)
				recs = append(recs, image.Rectangle{vectorstart, image.Point{vectorstart.X + width, vectorstart.Y + height}})
				vectorstart = image.Point{vectorstart.X, vectorstart.Y + height}
			}
			start = image.Point{start.X + width, start.Y}
		} else {
			height := 20 + randSrc.Intn(50)
			for vectorstart.In(canv) {
				width := 20 + randSrc.Intn(50)
				recs = append(recs, image.Rectangle{vectorstart, image.Point{vectorstart.X + width, vectorstart.Y + height}})
				vectorstart = image.Point{vectorstart.X + width, vectorstart.Y}
			}
			start = image.Point{start.X, start.Y + height}
		}

		// recs = append(recs, image.Rectangle{vectorstart, image.Point{vectorstart.X + width, vectorstart.Y + height}})
		direction++
	}

	return recs
}

type nodal struct {
	neighbours []int
	area       image.Rectangle
	// update to have masks as the future goes on for more wild shapes
}

func linearneighbour(tiles []image.Rectangle) []nodal {
	nds := make([]nodal, len(tiles))
	max := 0
	j := 0
	order := []int{0}

	// loop through every neighbour checking bounding boxes for overlaps
	for len(tiles) > j {
		t := tiles[j]
		adjaceny := []int{}
		// 	adjaceny[j] = make([]int, len(tiles))
		for k, neighbour := range tiles {

			if k == j { // skip its self

				continue
			}

			/*  edge detection ideas insert better algorthim possibly*/
			box1 := image.Rectangle{t.Min.Add(image.Point{1, -1}), t.Max.Add(image.Point{-1, 1})}
			box2 := image.Rectangle{t.Min.Add(image.Point{-1, 1}), t.Max.Add(image.Point{1, -1})}

			if (neighbour.Min.X < box1.Max.X && neighbour.Max.X > box1.Min.X &&
				neighbour.Min.Y < box1.Max.Y && neighbour.Max.Y > box1.Min.Y) || (neighbour.Min.X < box2.Max.X && neighbour.Max.X > box2.Min.X &&
				neighbour.Min.Y < box2.Max.Y && neighbour.Max.Y > box2.Min.Y) {
				adjaceny = append(adjaceny, k)

				if ok, _ := contains(order, k); !ok {
					order = append(order, k)
				}

			}

		}
		nds[j] = nodal{neighbours: adjaceny, area: t}
		j++
		if len(adjaceny) > max {
			max = len(adjaceny)
		}
	}

	return nds
}

func contains[T string | int](s []T, str T) (bool, int) {
	for i, v := range s {
		if v == str {
			return true, i
		}
	}

	return false, 0
}
