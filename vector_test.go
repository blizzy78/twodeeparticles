package twodeeparticles

import (
	"math"
	"testing"

	"github.com/matryer/is"
)

func TestVector_Magnitude(t *testing.T) {
	is := is.New(t)
	is.Equal(Vector{17, 23}.Magnitude(), math.Sqrt(17*17+23*23))
}

func TestVector_TryNormalize(t *testing.T) {
	is := is.New(t)

	v := Vector{17, 23}
	m := v.Magnitude()
	n, ok := v.TryNormalize()
	is.Equal(n.X, v.X/m)
	is.Equal(n.Y, v.Y/m)
	is.Equal(n.Magnitude(), 1.0)
	is.True(ok)

	v = Vector{0, 0}
	n, ok = v.TryNormalize()
	is.Equal(v, n)
	is.True(!ok)
}

func TestVector_Add(t *testing.T) {
	is := is.New(t)
	v1 := Vector{17, 23}
	v2 := Vector{5, 7}
	is.Equal(v1.Add(v2), Vector{v1.X + v2.X, v1.Y + v2.Y})
}

func TestVector_Multiply(t *testing.T) {
	is := is.New(t)
	is.Equal(Vector{17, 23}.Multiply(3), Vector{17 * 3, 23 * 3})
}
