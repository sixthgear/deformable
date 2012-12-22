package main

// import "fmt"
import "math"
import "github.com/sixthgear/noise"
import "github.com/go-gl/gl"

const PRIMITIVE_RESTART = math.MaxUint32

type Vertex struct {
	x, y float32
}

type Map struct {
	width, height int
	renderMode    int
	renderSmooth  bool
	gridSize      int
	grid          []int
	contour       []int
	gridLines     []float32
	vl            []*VertexList
}

type VertexList struct {
	vertices []float32
	indices  []uint
	colors   []float32
}

func GenerateMap(width, height int, gridSize int) *Map {

	m := new(Map)

	m.width = width
	m.height = height
	m.gridSize = gridSize
	m.grid = make([]int, width*height)

	diag := math.Hypot(float64(m.width/2), float64(m.height/2))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {

			fx := float64(x)
			fy := float64(y)
			// calculate inverse distance from center
			d := 1.0 - math.Hypot(float64(m.width/2)-fx, float64(m.height/2)-fy)/diag
			d = d
			v := noise.OctaveNoise2d(fx, fy, 4, 0.5, 1.0/64)
			v = (v + 1.0) / 2
			v = v * 256
			m.grid[y*width+x] = int(v)
			m.gridLines = append(m.gridLines, float32(x*m.gridSize), 0, float32(x*m.gridSize), float32((m.height-1)*m.gridSize))
		}
		m.gridLines = append(m.gridLines, 0, float32(y*m.gridSize), float32((m.width-1)*m.gridSize), float32(y*m.gridSize))
	}

	m.RebuildVertices()
	return m
}

func lerp(a, b, t float32) float32 {
	v := (t - a) / (b - a)
	return v
}

func (m *Map) RebuildVertices() {
	m.vl = make([]*VertexList, 4)
	m.vl[0] = m.GenerateIsoband(125, 150, [3]float32{1.0, 1.0, 1.0})
	m.vl[1] = m.GenerateIsoband(150, 175, [3]float32{0.8, 0.8, 0.8})
	m.vl[2] = m.GenerateIsoband(175, 200, [3]float32{0.625, 0.625, 0.625})
	m.vl[3] = m.GenerateIsoband(200, math.MaxInt32, [3]float32{0.5, 0.5, 0.5})
}

func (m *Map) GenerateIsoband(min, max int, color [3]float32) *VertexList {

	vl := new(VertexList)

	vl.vertices = make([]float32, 0)
	vl.indices = make([]uint, 0)
	vl.colors = append(vl.colors, color[0], color[1], color[2])
	threshold := make([]int, len(m.grid))
	count := uint(0)

	for i, v := range m.grid {
		switch {
		case v < min:
			threshold[i] = 0
		case v > max:
			threshold[i] = 2
		default:
			threshold[i] = 1
		}
	}

	for y := 0; y < m.height-1; y++ {
		for x := 0; x < m.width-1; x++ {

			polygon := make([]Vertex, 0)
			corners := [5][2]int{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}

			for i := 0; i < 4; i++ {

				var lerpTarget *float32
				x1, y1 := x+corners[i][0], y+corners[i][1]
				x2, y2 := x+corners[i+1][0], y+corners[i+1][1]
				edge := Vertex{float32(x1 * m.gridSize), float32(y1 * m.gridSize)}
				corner := Vertex{float32(x2 * m.gridSize), float32(y2 * m.gridSize)}

				switch i % 2 {
				case 0:
					lerpTarget = &edge.x
				case 1:
					lerpTarget = &edge.y
				}

				t1 := threshold[y1*m.width+x1]
				t2 := threshold[y2*m.width+x2]
				a := float32(m.grid[y1*m.width+x1])
				b := float32(m.grid[y2*m.width+x2])
				factor := float32(m.gridSize) * float32(-2*corners[i][1]+1)

				switch t1*3 + t2 {
				case 1:
					// corner in min
					*lerpTarget += lerp(a, b, float32(min)) * factor
					polygon = append(polygon, edge)   // add lerp edge
					polygon = append(polygon, corner) // add corner
				case 7:
					// corner in max
					*lerpTarget += lerp(a, b, float32(max)) * factor
					polygon = append(polygon, edge)   // add lerp edge
					polygon = append(polygon, corner) // add corner					
				case 3:
					// corner out min
					*lerpTarget += lerp(a, b, float32(min)) * factor
					polygon = append(polygon, edge) // // add lerp edge					
				case 5:
					// corner out max			
					*lerpTarget += lerp(a, b, float32(max)) * factor
					polygon = append(polygon, edge) // // add lerp edge										
				case 2:
					// double edge min -> max
					old := *lerpTarget
					*lerpTarget += lerp(a, b, float32(min)) * factor
					polygon = append(polygon, edge) // // add lerp edge
					*lerpTarget = old
					*lerpTarget += lerp(a, b, float32(max)) * factor
					polygon = append(polygon, edge) // // add lerp edge										
				case 6:
					// double edge max -> min
					old := *lerpTarget
					*lerpTarget += lerp(a, b, float32(max)) * factor
					polygon = append(polygon, edge) // add lerp edge
					*lerpTarget = old
					*lerpTarget += lerp(a, b, float32(min)) * factor
					polygon = append(polygon, edge) // add edge					
				case 4:
					// solid
					polygon = append(polygon, corner) // add corner					
				default:
					// blank, do nothing
				}
			}

			// Build manual triangle fan
			num := len(polygon)
			if num >= 3 {

				for i := range polygon {
					vl.vertices = append(vl.vertices, polygon[i].x, polygon[i].y)
				}

				for i := 2; i < num; i++ {
					vl.indices = append(vl.indices, count)
					vl.indices = append(vl.indices, count+uint(i)-1)
					vl.indices = append(vl.indices, count+uint(i))
				}

				count += uint(num)
			}
		}
	}

	return vl

}

func (m *Map) Add(x, y, val, radius int) {

	xi := (x - m.gridSize/2) / m.gridSize
	yi := (y - m.gridSize/2) / m.gridSize
	ri := radius / m.gridSize

	for y := -ri; y <= ri; y++ {
		for x := -ri; x <= ri; x++ {

			i := (yi+y)*m.width + (xi + x)
			if xi+x < 0 || xi+x >= m.width || yi+y < 0 || yi+y >= m.height {
				continue
			}

			a := float64(m.grid[i])
			b := float64(val)
			d := 1.0 - math.Hypot(float64(x), float64(y))/(float64(ri)*math.Sqrt2)
			d *= 0.25

			m.grid[i] = int(a + (b-a)*d)
			m.grid[i] = int(math.Max(0, float64(m.grid[i])))
			m.grid[i] = int(math.Min(512, float64(m.grid[i])))
		}
	}

	m.RebuildVertices()
}

func (m *Map) Draw() {

	// gl.Enable(gl.PRIMITIVE_RESTART)
	// gl.PrimitiveRestartIndex(PRIMITIVE_RESTART)
	gl.EnableClientState(gl.VERTEX_ARRAY)
	gl.Translatef(float32(m.gridSize/2), float32(m.gridSize/2), 0)

	if m.renderSmooth {
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		gl.Enable(gl.POLYGON_SMOOTH)
		gl.Hint(gl.POLYGON_SMOOTH_HINT, gl.NICEST)
	}

	if m.renderMode == 1 {
		gl.LineWidth(1)
		gl.VertexPointer(2, gl.FLOAT, 0, m.gridLines)
		gl.Color3f(0.2, 0.2, 0.2)
		gl.DrawArrays(gl.LINES, 0, len(m.gridLines)/2)
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	}

	for _, vl := range m.vl {
		if len(vl.vertices) > 0 {
			gl.VertexPointer(2, gl.FLOAT, 0, vl.vertices)
			gl.Color3f(vl.colors[0], vl.colors[1], vl.colors[2])
			gl.DrawElements(gl.TRIANGLES, len(vl.indices), gl.UNSIGNED_INT, vl.indices)
		}
	}

}

// gl.EnableClientState(gl.NORMAL_ARRAY)
// gl.NormalPointer(gl.FLOAT, 0, m.normals)
// gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
// gl.TexCoordPointer(2, gl.FLOAT, 0, m.texcoords)
// gl.EnableClientState(gl.COLOR_ARRAY)
// gl.ColorPointer(3, gl.FLOAT, 0, m.colors)
// gl.Enable(gl.COLOR_MATERIAL)
