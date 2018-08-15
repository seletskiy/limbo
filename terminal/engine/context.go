package engine

import (
	"github.com/go-gl/glfw/v3.2/glfw"
)

type Context struct {
	window *glfw.Window
	vao    uint32
}
