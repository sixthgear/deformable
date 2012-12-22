package main

// import "fmt"
import "math"
import "log"
import "github.com/go-gl/gl"
import "github.com/go-gl/glfw"

var drawing int

var (
	running          bool = true
	m                *Map
	brushValue       int        = 188
	currentBrushSize int        = 0
	brushSizes       []int      = []int{24, 32, 48, 64, 128}
	cursorVerts      []float64  = buildCursor(brushSizes[currentBrushSize])
	camera           [2]float32 = [2]float32{0, 0}
)

func main() {

	if err := glfw.Init(); err != nil {
		log.Fatal(err.Error())
	}

	if err := glfw.OpenWindow(800, 600, 8, 8, 8, 8, 8, 0, glfw.Windowed); err != nil {
		glfw.Terminate()
		log.Fatal(err.Error())
	}

	glfw.SetWindowTitle("Deformable")
	glfw.SetSwapInterval(1)
	glfw.SetMouseButtonCallback(handleMouseButton)
	glfw.SetKeyCallback(handleKeyDown)
	glfw.Disable(glfw.MouseCursor)

	initGL()

	i := 16
	m = GenerateMap(1600/i, 1200/i, i)
	for running && glfw.WindowParam(glfw.Opened) == 1 {

		x, y := glfw.MousePos()

		if drawing != 0 {
			m.Add(x+int(camera[0]), y+int(camera[1]), drawing, brushSizes[currentBrushSize])
		}

		switch 1 {
		case glfw.Key(glfw.KeyLeft), glfw.Key('A'):
			camera[0] -= 8
		case glfw.Key(glfw.KeyRight), glfw.Key('D'):
			camera[0] += 8
		}
		switch 1 {
		case glfw.Key(glfw.KeyUp), glfw.Key('W'):
			camera[1] -= 8
		case glfw.Key(glfw.KeyDown), glfw.Key('S'):
			camera[1] += 8
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

		glfw.SwapBuffers()
	}

	glfw.Terminate()

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
func handleKeyDown(key, state int) {

	if state == 0 {
		return
	}

	switch key {
	case '1':
		brushValue = 130
	case '2':
		brushValue = 155
	case '3':
		brushValue = 175
	case '4':
		brushValue = 205
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
	case glfw.KeyEsc:
		running = false
	case glfw.KeyTab:
		m.renderMode = (m.renderMode + 1) % 2
	case '\\':
		m.renderSmooth = !m.renderSmooth
	}
}

func handleMouseButton(button, state int) {
	switch {
	case button == 0 && state == 1:
		drawing = brushValue
	case button == 1 && state == 1:
		drawing = 40
	default:
		drawing = 0
	}
}
