package bench

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/mmTristan/opentsg-modules/opentsg-core/aces"
)

func TestWhat(t *testing.T) {
	// Maps just don't add up! :(
	match := 0
	for i := 0; i < 100000; i++ {
		m := mapTest()
		if m {
			match++
		}
	}

	fmt.Println(match)
}

// go test ./bench/ -bench=. -benchtime=40s

func BenchmarkNRGBA64(b *testing.B) {
	// decode to get the colour values

	img := image.NewNRGBA64(image.Rect(0, 0, 1920, 1080))

	colors := make(map[int]color.Color)
	colors[0] = color.NRGBA64{R: 0xffff, A: 0xffff}
	colors[1] = color.NRGBA64{G: 0xffff, A: 0xffff}
	colors[2] = color.NRGBA64{B: 0xffff, A: 0xffff}
	colors[3] = color.NRGBA64{R: 0xffff, G: 0xffff, B: 0xffff, A: 0xffff}
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		tester(img, colors)
	}
}

func BenchmarkNRGBA64ACESColour(b *testing.B) {
	// decode to get the colour values

	img := image.NewNRGBA64(image.Rect(0, 0, 1920, 1080))

	colors := make(map[int]color.Color)
	colors[0] = aces.RGBA128{R: 0xffff, A: 0xffff}
	colors[1] = aces.RGBA128{G: 0xffff, A: 0xffff}
	colors[2] = aces.RGBA128{B: 0xffff, A: 0xffff}
	colors[3] = aces.RGBA128{R: 0xffff, G: 0xffff, B: 0xffff, A: 0xffff}
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		tester(img, colors)
	}
}

func BenchmarkACES(b *testing.B) {
	// decode to get the colour values

	img := aces.NewARGBA(image.Rect(0, 0, 1920, 1080))

	colors := make(map[int]color.Color)
	colors[0] = aces.RGBA128{R: 0xffff, A: 0xffff}
	colors[1] = aces.RGBA128{G: 0xffff, A: 0xffff}
	colors[2] = aces.RGBA128{B: 0xffff, A: 0xffff}
	colors[3] = aces.RGBA128{R: 0xffff, G: 0xffff, B: 0xffff, A: 0xffff}
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		tester(img, colors)
	}
}

func BenchmarkACESNRGBAColour(b *testing.B) {
	// decode to get the colour values

	img := aces.NewARGBA(image.Rect(0, 0, 1920, 1080))

	colors := make(map[int]color.Color)
	colors[0] = color.NRGBA64{R: 0xffff, A: 0xffff}
	colors[1] = color.NRGBA64{G: 0xffff, A: 0xffff}
	colors[2] = color.NRGBA64{B: 0xffff, A: 0xffff}
	colors[3] = color.NRGBA64{R: 0xffff, G: 0xffff, B: 0xffff, A: 0xffff}
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		tester(img, colors)
	}
}
