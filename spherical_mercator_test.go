package sphericalmercator

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

var (
	maxExtentMerc  = []float64{-20037508.342789244, -20037508.342789244, 20037508.342789244, 20037508.342789244}
	maxExtentWgs84 = []float64{-180, -85.0511287798066, 180, 85.0511287798066}
)

func TestToPixel(t *testing.T) {
	type args struct {
		ll   []float64
		zoom int
	}
	tests := []struct {
		name string
		m    *SphericalMercator
		args args
		want []float64
	}{
		{
			"ToPixel",
			New(TileSize(256)),
			args{[]float64{-179, 85}, 9},
			[]float64{364, 215},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.ToPixel(tt.args.ll, tt.args.zoom); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SphericalMercator.ToPixel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToLonLat(t *testing.T) {
	type args struct {
		px   []float64
		zoom int
	}
	tests := []struct {
		name string
		m    *SphericalMercator
		args args
		want []float64
	}{
		{
			"ToLonLat",
			New(TileSize(256)),
			args{[]float64{200, 200}, 9},
			[]float64{-179.45068359375, 85.00351401304403},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.ToLonLat(tt.args.px, tt.args.zoom); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SphericalMercator.ToLonLat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBBOX(t *testing.T) {
	type args struct {
		x    int
		y    int
		zoom int
	}
	tests := []struct {
		name string
		m    *SphericalMercator
		args args
		want []float64
	}{
		{
			"[0,0,0]",
			New(TmsStyle()),
			args{0, 0, 0},
			[]float64{-180, -85.05112877980659, 180, 85.0511287798066},
		},
		{
			"[0,0,1]",
			New(TmsStyle()),
			args{0, 0, 1},
			[]float64{-180, -85.05112877980659, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.BBOX(tt.args.x, tt.args.y, tt.args.zoom); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SphericalMercator.BBOX() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestXYZ(t *testing.T) {
	type args struct {
		bbox []float64
		zoom int
	}
	tests := []struct {
		name string
		m    *SphericalMercator
		args args
		want []int
	}{
		{
			"World",
			New(TmsStyle()),
			args{[]float64{-180, -85.05112877980659, 180, 85.0511287798066}, 0},
			[]int{0, 0, 0, 0},
		},
		{
			"SW",
			New(TmsStyle()),
			args{[]float64{-180, -85.05112877980659, 0, 0}, 1},
			[]int{0, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.XYZ(tt.args.bbox, tt.args.zoom); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SphericalMercator.XYZ() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestXYZBroken(t *testing.T) {
	m := New(TmsStyle())
	extent := []float64{-0.087891, 40.95703, 0.087891, 41.044916}
	xyz := m.XYZ(extent, 3)
	if !(xyz[0] <= xyz[2]) {
		t.Fail()
	}
	if !(xyz[1] <= xyz[3]) {
		t.Fail()
	}
}

func TestXYZNegative(t *testing.T) {
	m := New()
	extent := []float64{-112.5, 85.0511, -112.5, 85.0511}
	xyz := m.XYZ(extent, 0)
	if !(xyz[0] == 0) {
		t.Fail()
	}
}

func TestXYZFuzz(t *testing.T) {
	m := New(TmsStyle())
	for i := 0; i < 1000; i++ {
		x1 := -180 + 360*rand.Float64()
		x2 := -180 + 360*rand.Float64()
		y1 := -85 + 170*rand.Float64()
		y2 := -85 + 170*rand.Float64()
		zoom := int(math.Floor(22 * rand.Float64()))
		extent := []float64{
			math.Min(x1, x2),
			math.Min(y1, y2),
			math.Max(x1, x2),
			math.Max(y1, y2),
		}
		xyz := m.XYZ(extent, zoom)
		if xyz[0] > xyz[2] {
			t.Fail()
		}
		if xyz[1] > xyz[3] {
			t.Fail()
		}
	}
}

func TestForward(t *testing.T) {
	type args struct {
		ll []float64
	}
	tests := []struct {
		name string
		m    *SphericalMercator
		args args
		want []float64
	}{
		{
			"SW",
			New(),
			args{maxExtentWgs84[:2]},
			maxExtentMerc[:2],
		},
		{
			"NE",
			New(),
			args{maxExtentWgs84[2:4]},
			maxExtentMerc[2:4],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Forward(tt.args.ll); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SphericalMercator.Forward() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInverse(t *testing.T) {
	type args struct {
		xy []float64
	}
	tests := []struct {
		name string
		m    *SphericalMercator
		args args
		want []float64
	}{
		{
			"SW",
			New(),
			args{maxExtentMerc[:2]},
			maxExtentWgs84[:2],
		},
		{
			"NE",
			New(),
			args{maxExtentMerc[2:4]},
			maxExtentWgs84[2:4],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Inverse(tt.args.xy); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SphericalMercator.Inverse() = %v, want %v", got, tt.want)
			}
		})
	}
}
