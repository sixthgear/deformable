package main

import (
	"log"
	"math"

	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
)

var drawing int

var (
	running           bool = true
	m                 *Map
	currentBrushValue int        = 0
	currentBrushSize  int        = 0
	brushValues       []int      = []int{145, 170, 195, 245}
	brushSizes        []int      = []int{24, 32, 48, 64, 128}
	cursorVerts       []float64  = buildCursor(brushSizes[currentBrushSize])
	camera            [2]float32 = [2]float32{0, 0}
)

func main() {

	if !glfw.Init() {
		log.Fatal("glfw failed to initialize")
	}
	defer glfw.Terminate()

	window, err := glfw.CreateWindow(640, 480, "Deformable", nil, nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	window.MakeContextCurrent()
	glfw.SwapInterval(1)
	window.SetMouseButtonCallback(handleMouseButton)
	window.SetKeyCallback(handleKeyDown)
	window.SetInputMode(glfw.Cursor, glfw.CursorHidden)

	gl.Init()
	initGL()

	i := 16
	m = GenerateMap(1600/i, 1200/i, i)
	for running && !window.ShouldClose() {

		x, y := window.GetCursorPosition()

		if drawing != 0 {
			m.Add(int(x)+int(camera[0]), int(y)+int(camera[1]), drawing, brushSizes[currentBrushSize])
		}

		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.LoadIdentity()

		gl.PushMatrix()
		gl.PushAttrib(gl.CURRENT_BIT | gl.ENABLE_BIT | gl.LIGHTING_BIT | gl.POLYGON_BIT | gl.LINE_BIT)
		gl.Translatef(-camera[0], -camera[1], 0)
		m.Draw()
		gl.PopAttrib()
		gl.PopMatrix()

		gl.PushAttrib(gl.COLOR_BUFFER_BIT)
		gl.LineWidth(2)
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.ONE_MINUS_DST_COLOR, gl.ZERO)
		// gl.Enable(gl.LINE_SMOOTH)
		// gl.Hint(gl.LINE_SMOOTH_HINT, gl.NICEST)

		gl.Translatef(float32(x), float32(y), 0)

		gl.EnableClientState(gl.VERTEX_ARRAY)
		gl.VertexPointer(2, gl.DOUBLE, 0, cursorVerts)
		gl.DrawArrays(gl.LINE_LOOP, 0, 24)
		gl.PopAttrib()

		window.SwapBuffers()
		glfw.PollEvents()
	}

}

func initGL() {
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Viewport(0, 0, 800, 600)
	gl.Ortho(0, 800, 600, 0, -1.0, 1.0)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	gl.ClearColor(0.1, 0.05, 0.0, 1.0)
}

func buildCursor(size int) []float64 {
	vl := make([]float64, 0)
	for i := 0; i < 24; i++ {
		a := float64(i) / 24 * (math.Pi * 2)
		v := float64(size)
		vl = append(vl, math.Cos(a)*v, math.Sin(a)*v)
	}
	return vl
}

func handleKeyDown(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

	if action == glfw.Release {
		return
	}

	switch key {
	case glfw.KeyLeft, glfw.KeyA:
		camera[0] -= 8
	case glfw.KeyRight, glfw.KeyD:
		camera[0] += 8
	case glfw.KeyUp, glfw.KeyW:
		camera[1] -= 8
	case glfw.KeyDown, glfw.KeyS:
		camera[1] += 8
	case '1':
		currentBrushValue = 0
	case '2':
		currentBrushValue = 1
	case '3':
		currentBrushValue = 2
	case '4':
		currentBrushValue = 3
	case '[':
		currentBrushSize--
		if currentBrushSize < 0 {
			currentBrushSize = 0
		}
		cursorVerts = buildCursor(brushSizes[currentBrushSize])
	case ']':
		currentBrushSize++
		if currentBrushSize > len(brushSizes)-1 {
			currentBrushSize = len(brushSizes) - 1
		}
		cursorVerts = buildCursor(brushSizes[currentBrushSize])
	case glfw.KeyEscape:
		running = false
	case glfw.KeyTab:
		m.renderMode = (m.renderMode + 1) % 2
	case '\\':
		m.renderSmooth = !m.renderSmooth
	}
}

func handleMouseButton(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	switch {
	case button == 0 && action == glfw.Press:
		drawing = brushValues[currentBrushValue]
	case button == 1 && action == glfw.Press:
		drawing = 1
	default:
		drawing = 0
	}
}
