package main

// import "fmt"
import "log"
import "github.com/go-gl/gl"
import "github.com/go-gl/glfw"

var drawing int

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

	initGL()

	i := 10
	m := GenerateMap(800/i, 600/i, i)
	for glfw.WindowParam(glfw.Opened) == 1 {

		if drawing != 0 {
			x, y := glfw.MousePos()
			m.Add(x, y, drawing, 16)
		}

		gl.Clear(gl.COLOR_BUFFER_BIT)
		m.Draw()
		glfw.SwapBuffers()
	}

	glfw.Terminate()

}

func initGL() {
	// gl.Init()
	// gl.DrawBuffer(gl.FRONT_AND_BACK)
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Viewport(0, 0, 800, 600)
	gl.Ortho(0, 800, 600, 0, -1.0, 1.0)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// gl.Enable(gl.LINE_SMOOTH)
	// gl.Enable(gl.POLYGON_SMOOTH)
	// gl.Hint(gl.LINE_SMOOTH_HINT, gl.NICEST)
	// gl.Hint(gl.POLYGON_SMOOTH_HINT, gl.NICEST)

	gl.ShadeModel(gl.SMOOTH)
	gl.ClearColor(0.1, 0.05, 0.0, 1.0)
}

func handleMouseButton(button, state int) {
	switch {
	case button == 0 && state == 1:
		drawing = 2
	case button == 1 && state == 1:
		drawing = -2
	default:
		drawing = 0
	}
}
