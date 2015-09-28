package ppu

import "image"

type Renderer interface {
	Render(img image.Image)
}
