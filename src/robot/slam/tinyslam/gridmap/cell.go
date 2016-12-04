package gridmap

import (
	"image/color"
)

type Cell interface {
	Gray() color.Gray
}

type SimpleCell uint16

func (c SimpleCell) Gray() color.Gray {
	return color.Gray{uint8(c >> 8)}
}