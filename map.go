package main

// import "fmt"
import "math"
import "github.com/sixthgear/noise"
import "github.com/go-gl/gl"

const PRIMITIVE_RESTART = math.MaxUint32

type Vertex struct {
	x, y float32
}

type Line struct {
	a, b Vertex
}

type Map struct {
	width, height int
	gridSize      int
	contourLevel  int
	grid          []int
	contour       []int
	gridLines     []float32
	vl            *VertexList
	vl2           *VertexList
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
	m.contourLevel = 128
	// m.vl = new(VertexList)

	m.grid = make([]int, width*height)

	diag := math.Hypot(float64(m.width/2), float64(m.height/2))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {

			fx := float64(x)
			fy := float64(y)
			// calculate inverse distance from center
			d := 1.0 - math.Hypot(float64(m.width/2)-fx, float64(m.height/2)-fy)/diag
			d = d
			v := noise.OctaveNoise2d(fx, fy, 4, 1, 1.0/128)
			v = (v + 1.0) / 2
			v = v * 256
			m.grid[y*width+x] = int(v)
			m.gridLines = append(m.gridLines, float32(x*m.gridSize), 0, float32(x*m.gridSize), float32((m.height-1)*m.gridSize))
		}
		m.gridLines = append(m.gridLines, 0, float32(y*m.gridSize), float32((m.width-1)*m.gridSize), float32(y*m.gridSize))
	}

	m.vl = m.GenerateIsoband(m.contourLevel, m.contourLevel+64)
	m.vl2 = m.GenerateIsoband(m.contourLevel+64, math.MaxInt32)

	return m
}

func lerp(a, b, t float32) float32 {
	v := (t - a) / (b - a)
	return v
}

func (m *Map) GenerateIsoband(min, max int) *VertexList {

	vl := new(VertexList)
	vl.vertices = make([]float32, 0)
	vl.indices = make([]uint, 0)
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

			var corners [5][2]int = [...][2]int{
				{0, 0},
				{1, 0},
				{1, 1},
				{0, 1},
				{0, 0},
			}

			for i := 0; i < 4; i++ {

				x1, y1 := x+corners[i][0], y+corners[i][1]
				x2, y2 := x+corners[i+1][0], y+corners[i+1][1]
				edge := Vertex{float32(x1 * m.gridSize), float32(y1 * m.gridSize)}
				corner := Vertex{float32(x2 * m.gridSize), float32(y2 * m.gridSize)}

				var lerpTarget *float32

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
					// 0, 8
				}
				// fmt.Printf("%d ", t1*3+t2)
			}
			// fmt.Printf(" - %d\n", len(polygon))

			// Build manual triangle fan
			num := len(polygon)
			if num >= 3 {

				for i := range polygon {
					vl.vertices = append(vl.vertices, polygon[i].x, polygon[i].y)
				}
				// fmt.Println(polygon)
				for i := 2; i < num; i++ {
					vl.indices = append(vl.indices, count)
					vl.indices = append(vl.indices, count+uint(i)-1)
					vl.indices = append(vl.indices, count+uint(i))
				}

				count += uint(num)
			}

			// m.vl.indices = append(m.vl.indices, PRIMITIVE_RESTART)
			// fmt.Println(count)

		}
	}

	return vl

}

func (m *Map) Add(x, y, val int, radius float64) {

	xi := (x - m.gridSize/2) / m.gridSize
	yi := (y - m.gridSize/2) / m.gridSize

	ri := int(math.Ceil(radius / float64(m.gridSize)))

	for i := yi - ri; i <= yi+ri; i++ {
		for j := xi - ri; j <= xi+ri; j++ {
			d := 1.0 - math.Hypot(float64(yi-i), float64(xi-j))/radius
			v := int(d * d * d * float64(val))
			m.grid[i*m.width+j] += v
			m.grid[i*m.width+j] = int(math.Max(0, math.Min(512, float64(m.grid[i*m.width+j]))))
		}
	}

	m.vl = m.GenerateIsoband(m.contourLevel, m.contourLevel+64)
	m.vl2 = m.GenerateIsoband(m.contourLevel+64, math.MaxInt32)
}

func (m *Map) Draw() {
	gl.PushMatrix()
	gl.PushAttrib(gl.CURRENT_BIT | gl.ENABLE_BIT | gl.LIGHTING_BIT | gl.POLYGON_BIT | gl.LINE_BIT)

	// gl.EnableClientState(gl.NORMAL_ARRAY)
	// gl.NormalPointer(gl.FLOAT, 0, m.normals)

	// gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
	// gl.TexCoordPointer(2, gl.FLOAT, 0, m.texcoords)

	// gl.EnableClientState(gl.COLOR_ARRAY)
	// gl.ColorPointer(3, gl.FLOAT, 0, m.colors)

	// draw solids
	// gl.Enable(gl.COLOR_MATERIAL)
	gl.Translatef(float32(m.gridSize/2), float32(m.gridSize/2), 0)

	gl.EnableClientState(gl.VERTEX_ARRAY)

	gl.VertexPointer(2, gl.FLOAT, 0, m.gridLines)
	gl.Color3f(0.2, 0.2, 0.2)
	gl.LineWidth(1)
	gl.DrawArrays(gl.LINES, 0, len(m.gridLines)/2)

	gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	if len(m.vl.vertices) > 0 {
		gl.VertexPointer(2, gl.FLOAT, 0, m.vl.vertices)
		gl.Color3f(1, 1, 1)
		gl.DrawElements(gl.TRIANGLES, len(m.vl.indices), gl.UNSIGNED_INT, m.vl.indices)
	}
	if len(m.vl2.vertices) > 0 {
		gl.VertexPointer(2, gl.FLOAT, 0, m.vl2.vertices)
		gl.Color3f(0.5, 0.5, 0.5)
		gl.DrawElements(gl.TRIANGLES, len(m.vl2.indices), gl.UNSIGNED_INT, m.vl2.indices)
	}

	gl.PopAttrib()
	gl.PopMatrix()
}
