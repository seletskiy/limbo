package engine

import (
	"image"
	"image/color"
	"runtime"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/kovetskiy/lorg"
	"github.com/reconquest/karma-go"
)

func init() {
	// Required for OpenGL to work.
	runtime.LockOSThread()
}

type Engine struct {
	vertices struct {
		points []float32

		buffers struct {
			triangles  uint32
			attributes uint32
		}
	}

	shaders struct {
		program  uint32
		vertex   uint32
		fragment uint32
	}

	contexts struct {
		last    uint32
		handles map[uint32]*Context
	}

	font struct {
		image   *image.RGBA
		texture uint32
	}

	log lorg.Logger
}

func New(log lorg.Logger) (*Engine, error) {
	err := gl.Init()
	if err != nil {
		return nil, karma.Format(
			err,
			"{gl} unable to init",
		)
	}

	err = glfw.Init()
	if err != nil {
		return nil, karma.Format(
			err,
			"{glfw} unable to init",
		)
	}

	engine := &Engine{}

	engine.log = log
	engine.contexts.handles = map[uint32]*Context{}

	return engine, nil
}

func (engine *Engine) CreateWindow(
	width int,
	height int,
	title string,
) (uint32, error) {
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	var parent *glfw.Window
	for _, context := range engine.contexts.handles {
		parent = context.window
		break
	}

	window, err := glfw.CreateWindow(width, height, title, nil, parent)
	if err != nil {
		return 0, karma.Format(
			err,
			"{glfw} unable to creaate window",
		)
	}

	window.SetFramebufferSizeCallback(
		func(window *glfw.Window, width int, height int) {
			gl.Viewport(0, 0, int32(width), int32(height))
		},
	)

	window.MakeContextCurrent()

	gl.Enable(gl.DEBUG_OUTPUT)
	gl.DebugMessageCallback(engine.debug, nil)

	context := &Context{
		window: window,
	}

	gl.GenVertexArrays(1, &context.vao)

	engine.contexts.last++
	engine.contexts.handles[engine.contexts.last] = context

	return engine.contexts.last, nil
}

func (engine *Engine) Render() error {
	for _, context := range engine.contexts.handles {
		err := engine.render(context)
		if err != nil {
			return err
		}
	}

	return nil
}

func (engine *Engine) SetFont(img image.Image) {
	engine.font.image = image.NewRGBA(img.Bounds())

	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := img.At(x, y)
			r, b, g, _ := pixel.RGBA()

			if r == 0 && g == 0 && b == 0 {
				engine.font.image.Set(x, y, color.White)
			}
		}
	}
}

func (engine *Engine) Empty() bool {
	for handle, context := range engine.contexts.handles {
		if context.window.ShouldClose() {
			context.window.Destroy()
			delete(engine.contexts.handles, handle)
		}
	}

	if len(engine.contexts.handles) == 0 {
		return true
	}

	return false
}

func (engine *Engine) Stop() {
	glfw.Terminate()
}

func (engine *Engine) render(context *Context) error {
	context.window.MakeContextCurrent()

	err := engine.initShaders()
	if err != nil {
		return err
	}

	err = engine.initVertices()
	if err != nil {
		return err
	}

	engine.initTextures()

	gl.BindVertexArray(context.vao)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.ClearColor(0, 0, 0, 1)

	glyphWidth, glyphHeight := 8, 16

	width, height := context.window.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))

	gl.Uniform2i(0, int32(width), int32(height))
	gl.Uniform2i(1, int32(glyphWidth), int32(glyphHeight))

	glyphs := int(width/glyphWidth) * int(height/glyphHeight)

	textures := make([]int32, glyphs*2)

	textures[0] = 5
	textures[1] = int32(height % glyphHeight)

	gl.BindBuffer(gl.ARRAY_BUFFER, engine.vertices.buffers.triangles)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 2*4, gl.PtrOffset(0))

	gl.BindBuffer(gl.ARRAY_BUFFER, engine.vertices.buffers.attributes)
	gl.BufferData(
		gl.ARRAY_BUFFER,
		4*len(textures),
		gl.Ptr(textures),
		gl.DYNAMIC_DRAW,
	)

	gl.VertexAttribIPointer(1, 2, gl.INT, 2*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribDivisor(1, 1)

	gl.DrawArraysInstanced(gl.TRIANGLE_STRIP, 0, 6, int32(glyphs))

	glfw.PollEvents()
	context.window.SwapBuffers()

	return nil
}

func (engine *Engine) initTextures() error {
	if engine.font.texture > 0 {
		gl.BindTexture(gl.TEXTURE_2D, engine.font.texture)

		return nil
	}

	gl.GenTextures(1, &engine.font.texture)
	gl.BindTexture(gl.TEXTURE_2D, engine.font.texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_BORDER)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(engine.font.image.Bounds().Size().X),
		int32(engine.font.image.Bounds().Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(engine.font.image.Pix),
	)

	return nil
}

func (engine *Engine) initShaders() error {
	if engine.shaders.program > 0 {
		gl.UseProgram(engine.shaders.program)

		return nil
	}

	var err error

	engine.shaders.vertex, err = engine.compileShader(
		gl.VERTEX_SHADER,
		vertexShader,
	)
	if err != nil {
		return karma.Format(
			err,
			"{gl} unable to compile vertex shader",
		)
	}

	engine.shaders.fragment, err = engine.compileShader(
		gl.FRAGMENT_SHADER,
		fragmentShader,
	)
	if err != nil {
		return karma.Format(
			err,
			"{gl} unable to compile fragment shader",
		)
	}

	engine.shaders.program = gl.CreateProgram()

	gl.AttachShader(engine.shaders.program, engine.shaders.vertex)
	gl.AttachShader(engine.shaders.program, engine.shaders.fragment)
	gl.LinkProgram(engine.shaders.program)
	gl.UseProgram(engine.shaders.program)

	return nil
}

func (engine *Engine) initVertices() error {
	if engine.vertices.points != nil {
		return nil
	}

	//         ->
	// (0; 1) X--X (1; 1)
	//      ^ |\ | |
	//      | | \| V
	// (0; 0) X--X (1; 0)
	//         <-
	engine.vertices.points = []float32{
		0, 0,
		0, 1,
		1, 0,

		1, 0,
		0, 1,
		1, 1,
	}

	// First buffer for vertices, which form triangles, which form cells.
	gl.GenBuffers(1, &engine.vertices.buffers.triangles)

	gl.BindBuffer(gl.ARRAY_BUFFER, engine.vertices.buffers.triangles)

	gl.BufferData(
		gl.ARRAY_BUFFER,
		len(engine.vertices.points)*4,
		gl.Ptr(engine.vertices.points),
		gl.STATIC_DRAW,
	)

	// Second buffer for vertices data, in this case it's glyph coordinates
	// in font.
	gl.GenBuffers(1, &engine.vertices.buffers.attributes)

	return nil
}

func (engine *Engine) compileShader(
	kind uint32,
	source string,
) (uint32, error) {
	handle := gl.CreateShader(kind)

	buffer, free := gl.Strs(source + "\x00")
	defer free()

	gl.ShaderSource(handle, 1, buffer, nil)
	gl.CompileShader(handle)

	var result int32
	gl.GetShaderiv(handle, gl.COMPILE_STATUS, &result)
	if result == gl.FALSE {
		var length int32
		gl.GetShaderiv(handle, gl.INFO_LOG_LENGTH, &length)

		err := strings.Repeat("\x00", int(length+1))
		gl.GetShaderInfoLog(handle, length, nil, gl.Str(err))

		return 0, karma.Describe("source", source).Format(
			err,
			"{shader} compilation error",
		)
	}

	return handle, nil
}

func (engine *Engine) debug(
	source uint32,
	kind uint32,
	id uint32,
	severity uint32,
	length int32,
	message string,
	userParam unsafe.Pointer,
) {
	logger := engine.log.Warningf

	switch kind {
	case gl.DEBUG_TYPE_ERROR:
		logger = engine.log.Errorf
	case gl.DEBUG_TYPE_OTHER:
		logger = engine.log.Debugf
	}

	logger(
		"{gl} %s | type=0x%x severity=0x%x source=0x%x",
		message,
		kind,
		severity,
		source,
	)
}
