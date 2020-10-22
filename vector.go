package twodeeparticles

import (
	"errors"
	"math"
)

var (
	// ZeroVector is a vector with a magnitude of zero.
	ZeroVector = Vector{0.0, 0.0}

	// OneVector is a vector whose components are all one.
	OneVector = Vector{1.0, 1.0}
)

var errNormalizeZeroVector = errors.New("normalize zero vector")

// A Vector is a geometric entity that has a direction and a length.
type Vector struct {
	X float64
	Y float64
}

// Magnitude returns the length of v.
func (v Vector) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// Normalize returns a vector that has the same direction as v, but whose length is one.
// In other words, it returns a unit vector with the same direction as v.
// If v has a length of zero, it will panic.
func (v Vector) Normalize() Vector {
	n, ok := v.TryNormalize()
	if !ok {
		panic(errNormalizeZeroVector)
	}
	return n
}

// TryNormalize returns a vector that has the same direction as v, but whose length is one.
// In other words, it returns a unit vector with the same direction as v.
// If v has a length of zero, it will return v and false, else the described result and true.
func (v Vector) TryNormalize() (Vector, bool) {
	m := v.Magnitude()
	if m == 0 {
		return v, false
	}
	return Vector{v.X / m, v.Y / m}, true
}

// Add returns a vector whose components are component-wise additions of v and v2.
func (v Vector) Add(v2 Vector) Vector {
	return Vector{v.X + v2.X, v.Y + v2.Y}
}

// Multiply returns a vector whose components are v's components multiplied by d.
func (v Vector) Multiply(d float64) Vector {
	return Vector{v.X * d, v.Y * d}
}
