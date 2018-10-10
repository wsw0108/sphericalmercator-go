package sphericalmercator // import "github.com/wsw0108/sphericalmercator-go"

import (
	"math"
	"sync"
)

const (
	r2d       = 180.0 / math.Pi
	d2r       = math.Pi / 180.0
	a         = 6378137.0
	maxExtent = 20037508.342789244
)

var mu sync.Mutex
var cache = map[int]*cacheEntry{}

type SphericalMercator struct {
	size int
	tms  bool
	bc   []float64
	cc   []float64
	zc   []float64
	ac   []float64
}

type OptSetter func(m *SphericalMercator)

func TileSize(size int) OptSetter {
	return func(m *SphericalMercator) {
		m.size = size
	}
}

func TmsStyle() OptSetter {
	return func(m *SphericalMercator) {
		m.tms = true
	}
}

func New(setters ...OptSetter) *SphericalMercator {
	m := &SphericalMercator{size: 256}
	for _, s := range setters {
		s(m)
	}
	m.init()
	return m
}

func (m *SphericalMercator) init() {
	mu.Lock()
	defer mu.Unlock()
	c := cache[m.size]
	if c == nil {
		c = newCacheEntry()
		cache[m.size] = c
		size := m.size
		for d := 0; d < 30; d++ {
			c.bc = append(c.bc, float64(size)/360.0)
			c.cc = append(c.cc, float64(size)/(2*math.Pi))
			c.zc = append(c.zc, float64(size)/2)
			c.ac = append(c.ac, float64(size))
			size *= 2
		}
	}
	m.bc = c.bc
	m.cc = c.cc
	m.zc = c.zc
	m.ac = c.ac
}

func (m *SphericalMercator) ToPixel(ll []float64, zoom int) []float64 {
	d := m.zc[zoom]
	f := math.Min(math.Max(math.Sin(d2r*ll[1]), -0.9999), 0.9999)
	x := math.Round(d + ll[0]*m.bc[zoom])
	y := math.Round(d + 0.5*math.Log((1+f)/(1-f))*(-m.cc[zoom]))
	if x > m.ac[zoom] {
		x = m.ac[zoom]
	}
	if y > m.ac[zoom] {
		y = m.ac[zoom]
	}
	return []float64{x, y}
}

func (m *SphericalMercator) ToLonLat(px []float64, zoom int) []float64 {
	g := (px[1] - m.zc[zoom]) / (-m.cc[zoom])
	lon := (px[0] - m.zc[zoom]) / m.bc[zoom]
	lat := r2d * (2*math.Atan(math.Exp(g)) - 0.5*math.Pi)
	return []float64{lon, lat}
}

func (m *SphericalMercator) BBOX(x, y, zoom int) []float64 {
	if m.tms {
		y = int(math.Pow(float64(2), float64(zoom))-1) - y
	}
	ll := []float64{float64(x * m.size), float64((y + 1) * m.size)}
	ur := []float64{float64((x + 1) * m.size), float64(y * m.size)}
	p0 := m.ToLonLat(ll, zoom)
	p1 := m.ToLonLat(ur, zoom)
	bbox := []float64{p0[0], p0[1], p1[0], p1[1]}
	return bbox
}

func (m *SphericalMercator) XYZ(bbox []float64, zoom int) []int {
	ll := bbox[0:2]
	ur := bbox[2:4]
	llPx := m.ToPixel(ll, zoom)
	urPx := m.ToPixel(ur, zoom)
	size := float64(m.size)
	x0 := math.Floor(llPx[0] / size)
	x1 := math.Floor((urPx[0] - 1) / size)
	y0 := math.Floor(urPx[1] / size)
	y1 := math.Floor((llPx[1] - 1) / size)
	minX := int(math.Min(x0, x1))
	if minX < 0 {
		minX = 0
	}
	minY := int(math.Min(y0, y1))
	if minY < 0 {
		minY = 0
	}
	maxX := int(math.Max(x0, x1))
	maxY := int(math.Max(y0, y1))
	bounds := []int{minX, minY, maxX, maxY}
	if m.tms {
		minY = int(math.Pow(float64(2), float64(zoom))-1) - bounds[3]
		maxY = int(math.Pow(float64(2), float64(zoom))-1) - bounds[1]
		bounds[1] = minY
		bounds[3] = maxY
	}
	return bounds
}

func clamp(n, min, max float64) float64 {
	if n > max {
		n = max
	}
	if n < min {
		n = min
	}
	return n
}

func (m *SphericalMercator) Forward(ll []float64) []float64 {
	x := a * ll[0] * d2r
	y := a * math.Log(math.Tan((math.Pi*0.25)+(0.5*ll[1]*d2r)))
	x = clamp(x, -maxExtent, maxExtent)
	y = clamp(y, -maxExtent, maxExtent)
	return []float64{x, y}
}

func (m *SphericalMercator) Inverse(xy []float64) []float64 {
	x := xy[0] * r2d / a
	y := ((math.Pi * 0.5) - 2.0*math.Atan(math.Exp(-xy[1]/a))) * r2d
	return []float64{x, y}
}

type cacheEntry struct {
	bc []float64
	cc []float64
	zc []float64
	ac []float64
}

func newCacheEntry() *cacheEntry {
	return &cacheEntry{}
}
