package fourcolour

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"sync"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	errhandle "github.com/mrmxf/opentsg-modules/opentsg-core/errHandle"
	"github.com/mrmxf/opentsg-modules/opentsg-core/gridgen"
	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"github.com/mrmxf/opentsg-modules/opentsg-core/widgethandler"
)

const (
	WidgetType = "builtin.fourcolor"
)

func FourColourGenerator(canvasChan chan draw.Image, debug bool, c *context.Context, wg, wgc *sync.WaitGroup, logs *errhandle.Logger) {
	defer wg.Done()
	opts := []any{c}
	conf := widgethandler.GenConf[Config]{Debug: debug, Schema: Schema, WidgetType: WidgetType, ExtraOpt: opts}
	widgethandler.WidgetRunner(canvasChan, conf, c, logs, wgc) // Update this to pass an error which is then formatted afterwards
}

var getGeometry = gridgen.GetGridGeometry

func (f Config) Handle(resp tsg.Response, req *tsg.Request) {
	if len(f.Colourpallette) < 4 {
		resp.Write(tsg.WidgetError, fmt.Sprintf("invalid number of colours chosen for the fourcolour pallette, need at least 4 got %v", len(f.Colourpallette)))
		return
	}
	pallette := make([]color.Color, len(f.Colourpallette))

	for i, c := range f.Colourpallette {
		pallette[i] = c.ToColour(f.ColourSpace)
	}

	flats := req.PatchProperties.Geometry

	namelocations := make(map[string]int)
	nodes := make([]nodal, len(flats))

	// TODO Manual neighbour finding or ensure they must always suggest having a neighbour
	//	fmt.Println(len(nodes))
	for i, flat := range flats {
		neighs := []int{}

		for _, ftag := range flat.Tags {
			if len(ftag) < 10 {
				continue
			}

			if ftag[:10] == "neighbour:" {
				// do some maths about incrementing the start point as more neighbours are found
				neigh := ftag[10:]
				neighpos, ok := namelocations[neigh]
				if !ok {
					for j, f := range flats {
						if f.Name == neigh {
							neighpos = j
							namelocations[neigh] = j
							ok = true
						}
					}
				}
				if ok {
					neighs = append(neighs, neighpos)
				}
			}
		}
		nodes[i] = nodal{neighbours: neighs, area: flat.Shape}
		//	fmt.Println(len(neighs), neighs)
	}

	// extract the colour here
	_, filled := bruteColourArea(nodes, len(pallette)+1)
	// Break if there's an error etc
	for _, node := range filled {
		setcolour := node.color
		// fmt.Println(node.area, canvas.Bounds(), setcolour)
		colour.Draw(resp.BaseImage(), node.area, &image.Uniform{pallette[setcolour-1]}, image.Point{}, draw.Src)

	}
}

// amend so that the number of colours is based off of the input, can be upgraded to 5 or 6 for performance
func (f Config) Generate(canvas draw.Image, opt ...any) error {

	if len(f.Colourpallette) < 4 {
		return fmt.Errorf("invalid number of colours chosen for the fourcolour pallette")
	}
	pallette := make([]color.Color, len(f.Colourpallette))

	for i, c := range f.Colourpallette {
		pallette[i] = c.ToColour(f.ColourSpace)
	}

	var c *context.Context
	if len(opt) != 0 {
		var ok bool
		c, ok = opt[0].(*context.Context)
		if !ok {
			return fmt.Errorf("0DEV configuration error when assiging fourcolour context")
		}
	} else {
		return fmt.Errorf("0DEV configuration error when assiging fourcolour context")
	}

	flats, err := getGeometry(c, f.GridLoc.Location)
	if err != nil {
		return err
	}

	namelocations := make(map[string]int)
	nodes := make([]nodal, len(flats))

	// TODO Manual neighbour finding or ensure they must always suggest having a neighbour
	//	fmt.Println(len(nodes))
	for i, flat := range flats {
		neighs := []int{}

		for _, ftag := range flat.Tags {
			if len(ftag) < 10 {
				continue
			}

			if ftag[:10] == "neighbour:" {
				// do some maths about incrementing the start point as more neighbours are found
				neigh := ftag[10:]
				neighpos, ok := namelocations[neigh]
				if !ok {
					for j, f := range flats {
						if f.Name == neigh {
							neighpos = j
							namelocations[neigh] = j
							ok = true
						}
					}
				}
				if err == nil && ok {
					neighs = append(neighs, neighpos)
				}
			}
		}
		nodes[i] = nodal{neighbours: neighs, area: flat.Shape}
		//	fmt.Println(len(neighs), neighs)
	}

	// extract the colour here
	_, filled := bruteColourArea(nodes, len(pallette)+1)
	// Break if there's an error etc
	for _, node := range filled {
		setcolour := node.color
		// fmt.Println(node.area, canvas.Bounds(), setcolour)
		colour.Draw(canvas, node.area, &image.Uniform{pallette[setcolour-1]}, image.Point{}, draw.Src)

	}
	// TODO add timeout feature with user and other input

	return nil
}

type nodal struct {
	neighbours []int
	color      int
	area       image.Rectangle
	// update to have masks as the future goes on for more wild shapes
}

// add a colour count to make colouring in easier where everything is based off of the number of colours given by a user
func bruteColourArea(colourNodes []nodal, colourLength int) (bool, []nodal) {
	// loop through every colour and check its nieghbours if there are no clashes move onto the next node.
	// there is no thought in why it just recurses until a solution is reached

	nodePos := 0
	max := len(colourNodes)
	zeroes := true
	for nodePos < max {

		colournode := colourNodes[nodePos]
		if colournode.color == 0 { // 0 isn't a colour
			colournode.color = 1
		}

		for c := colournode.color; c < colourLength; c++ {

			//	fmt.Println(node, c)
			setcolour := c
			//	check the chosen colour does not clash with one of the neighbours
			match := false
			for _, k := range colournode.neighbours {
				if colourNodes[k].color == setcolour {
					match = true
					// This colour doesn't work as it matches with a neighbour

					break
				}

			}

			colournode.color = setcolour
			colourNodes[nodePos] = colournode

			var nextNode bool
			switch {
			case !match && (nodePos != max-1): // if there are no neighbour matches go onto the next node
				nodePos++ // moe to the next colour wheel after breaking

				nextNode = true

			case !match && (nodePos == max-1):
				nodePos++

				nextNode = true
				//	return true, colourNodes
				/*else if nodePos == max-1 {
					colourNodes[nodePos].color = 0
					nodePos++

					break
				}*/
			case c == colourLength-1 && match:
				colourNodes[nodePos].color = 0
				nodePos++
				// flag that a zero has emerged
				zeroes = false

			}

			if nextNode {
				break
			}
		}

	}

	// loop through these bits

	C := 0
	for !zeroes {
		zeroes, colourNodes = empty(colourNodes, colourLength)
		C++
		//	fmt.Println("RUN", C)
		// if c==100 pull the plug for everbodies sake
	}

	return true, colourNodes
}

// empty searches for empty nodes
// and recursivley fills them if possible, if not a false value is returned.
func empty(colourNodes []nodal, colourLength int) (bool, []nodal) {
	fail := true

	for i := range colourNodes {

		pos := colourNodes[i]
		if pos.color == 0 {
			// fmt.Println("TRIGGER")

			neighbours := []int{i}
			neighbours = append(neighbours, pos.neighbours...)
			moreN := len(neighbours)

			for _, n := range neighbours[1:] {
				neighbours = append(neighbours, colourNodes[n].neighbours...)
			}
			moreNN := len(neighbours)
			for _, n := range neighbours[1:moreN] {
				neighbours = append(neighbours, colourNodes[n].neighbours...)
			}

			for _, n := range neighbours[moreN:moreNN] {
				neighbours = append(neighbours, colourNodes[n].neighbours...)
			}

			// reset all the neighbours to 0
			for _, n := range neighbours {
				colourNodes[n].color = 0
			}
			segPos := 0

			max := len(neighbours)

			// recursively search the neighbours
			for segPos < max {

				nodePos := neighbours[segPos]
				colournode := colourNodes[nodePos]

				if colournode.color == 0 { // 0 isn't a colour
					colournode.color = 1
				}

				for c := colournode.color; c < colourLength; c++ {

					setcolour := c
					//	check the chosen colour does not clash with one of the neighbours
					match := false
					for _, k := range colournode.neighbours {
						if colourNodes[k].color == setcolour {
							match = true
							// This colour doesn't work as it matches with a neighbour

							break
						}

					}
					colournode.color = setcolour
					colourNodes[nodePos] = colournode
					if !match { // if there are no neighbour matches go onto the next node
						segPos++ // moe to the next colour wheel after breaking

						break

					} else if c == colourLength-1 && match {

						not4 := false
						// go back until a node that is not four is found
						for !not4 {
							nodePos := neighbours[segPos]
							if colourNodes[nodePos].color < colourLength-1 && colourNodes[nodePos].color > 0 {

								colourNodes[nodePos].color++
								not4 = true

							} else {

								colourNodes[nodePos].color = 0
								segPos--
								if segPos == -1 { // if we move all the way back to the start then something is wrong

									segPos = len(neighbours) // move segpos to the end so it quits the loop
									fail = false

									break
								}
							}

						}

					}

				}
			}
		}
	}

	return fail, colourNodes
}
