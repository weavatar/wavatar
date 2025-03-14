package wavatar

import (
	"embed"
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand/v2"
	"path"
)

//go:embed all:parts/*
var parts embed.FS

const (
	AvatarSize = 80
	BgCount    = 4
	FaceCount  = 11
	BrowCount  = 8
	EyeCount   = 13
	PupilCount = 11
	MouthCount = 19
)

// New creates a new Wavatar from a hash (typically an MD5 hash of an email)
func New(hash []byte) image.Image {
	h := fnv.New64a()
	if _, err := h.Write(hash); err != nil {
		panic(err)
	}

	r := rand.New(rand.NewPCG(h.Sum64(), (h.Sum64()>>1)|1))
	face := r.IntN(FaceCount) + 1
	bgColor := r.IntN(240) + 1
	fade := r.IntN(BgCount) + 1
	wavColor := r.IntN(240) + 1
	brow := r.IntN(BrowCount) + 1
	eyes := r.IntN(EyeCount) + 1
	pupil := r.IntN(PupilCount) + 1
	mouth := r.IntN(MouthCount) + 1

	// Create background
	img := image.NewRGBA(image.Rect(0, 0, AvatarSize, AvatarSize))

	// Background color
	bgRGB := hsl(bgColor, 240, 50)
	bgCol := color.RGBA{R: uint8(bgRGB[0]), G: uint8(bgRGB[1]), B: uint8(bgRGB[2]), A: 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bgCol}, image.Point{}, draw.Src)

	// Apply fade pattern
	applyImage(img, "fade", fade)

	// Apply mask
	applyImage(img, "mask", face)

	// Fill with wave color
	wavRGB := hsl(wavColor, 240, 170)
	wavCol := color.RGBA{R: uint8(wavRGB[0]), G: uint8(wavRGB[1]), B: uint8(wavRGB[2]), A: 255}

	centerX, centerY := AvatarSize/2, AvatarSize/2
	floodFill(img, centerX, centerY, wavCol)

	// Apply remaining layers in order
	applyImage(img, "shine", face)
	applyImage(img, "brow", brow)
	applyImage(img, "eyes", eyes)
	applyImage(img, "pupils", pupil)
	applyImage(img, "mouth", mouth)

	return img
}

// applyImage loads and applies a PNG part to the base image
func applyImage(base *image.RGBA, part string, num int) {
	filename := fmt.Sprintf("%s%d.png", part, num)
	file, err := parts.Open(path.Join("parts", filename))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	partImage, err := png.Decode(file)
	if err != nil {
		panic(err)
	}

	draw.Draw(base, base.Bounds(), partImage, image.Point{}, draw.Over)
}

// hsl converts HSL color values to RGB
func hsl(h, s, l int) []int {
	var R, G, B int

	if h > 240 || h < 0 || s > 240 || s < 0 || l > 240 || l < 0 {
		return []int{0, 0, 0}
	}

	if h <= 40 {
		R = 255
		G = h / 40.0 * 256.0
		B = 0
	} else if h <= 80 {
		R = (1.0 - (h-40.0)/40.0) * 256.0
		G = 255
		B = 0
	} else if h <= 120 {
		R = 0
		G = 255
		B = (h - 80.0) / 40.0 * 256.0
	} else if h <= 160 {
		R = 0
		G = (1.0 - (h-120.0)/40.0) * 256.0
		B = 255
	} else if h <= 200 {
		R = (h - 160.0) / 40.0 * 256.0
		G = 0
		B = 255
	} else { // h > 200
		R = 255
		G = 0
		B = (1.0 - (h-200.0)/40.0) * 256.0
	}

	R = R + (240-s)/240*(128-R)
	G = G + (240-s)/240*(128-G)
	B = B + (240-s)/240*(128-B)

	if l < 120 {
		R = (R / 120) * l
		G = (G / 120) * l
		B = (B / 120) * l
	} else {
		R = l*((256-R)/120) + 2*R - 256
		G = l*((256-G)/120) + 2*G - 256
		B = l*((256-B)/120) + 2*B - 256
	}

	return []int{clamp(R), clamp(G), clamp(B)}
}

// clamp ensures a value is between 0 and 255
func clamp(v int) int {
	if v < 0 {
		return 0
	} else if v > 255 {
		return 255
	}
	return v
}

// floodFill performs a flood fill starting at (x,y) with the given color
func floodFill(img *image.RGBA, x, y int, col color.RGBA) {
	type point struct{ x, y int }

	// Get the color at the start point
	startColor := img.RGBAAt(x, y)

	// If the start point is already the target color, do nothing
	if startColor == col {
		return
	}

	// Use a queue for breadth-first traversal
	queue := []point{{x, y}}
	bounds := img.Bounds()

	for len(queue) > 0 {
		// Get the next point from the queue
		p := queue[0]
		queue = queue[1:]

		// If this point is outside the bounds or not the start color, skip it
		if p.x < bounds.Min.X || p.x >= bounds.Max.X ||
			p.y < bounds.Min.Y || p.y >= bounds.Max.Y {
			continue
		}

		currentColor := img.RGBAAt(p.x, p.y)
		if currentColor != startColor {
			continue
		}

		// Set the color at this point
		img.SetRGBA(p.x, p.y, col)

		// Add adjacent points to the queue
		queue = append(queue, point{p.x + 1, p.y})
		queue = append(queue, point{p.x - 1, p.y})
		queue = append(queue, point{p.x, p.y + 1})
		queue = append(queue, point{p.x, p.y - 1})
	}
}
