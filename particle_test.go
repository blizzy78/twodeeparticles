package twodeeparticles

import (
	"image/color"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestParticle_System(t *testing.T) {
	is := is.New(t)
	s := NewSystem()
	p := newParticle(s)
	is.Equal(p.System(), s)
}

func TestParticle_Update(t *testing.T) {
	is := is.New(t)

	sys := NewSystem()

	sys.MaxParticles = 1

	sys.LifetimeOverTime = func(d time.Duration, delta time.Duration) time.Duration {
		return 1500 * time.Millisecond
	}

	sys.DataOverLifetime = func(old any, t NormalizedDuration, delta time.Duration) any {
		return "data"
	}

	sys.EmissionPositionOverTime = func(d time.Duration, delta time.Duration) Vector {
		return Vector{17, 23}
	}

	sys.VelocityOverLifetime = func(p *Particle, t NormalizedDuration, delta time.Duration) Vector {
		return Vector{3, 5}
	}

	sys.ScaleOverLifetime = func(p *Particle, t NormalizedDuration, delta time.Duration) Vector {
		return Vector{7, 11}
	}

	sys.ColorOverLifetime = func(p *Particle, t NormalizedDuration, delta time.Duration) color.Color {
		return color.RGBA{0x12, 0x23, 0x34, 0x45}
	}

	sys.RotationOverLifetime = func(p *Particle, t NormalizedDuration, delta time.Duration) float64 {
		return 0.123
	}

	updateCalled := false
	sys.UpdateFunc = func(part *Particle, t NormalizedDuration, delta time.Duration) {
		updateCalled = true
	}

	deathCalled := false
	sys.DeathFunc = func(p *Particle) {
		deathCalled = true
	}

	sys.Spawn(1)

	now := time.Now()
	sys.Update(now)

	var part *Particle

	sys.ForEachParticle(func(p *Particle, t NormalizedDuration, delta time.Duration) {
		part = p
	}, now)

	is.Equal(part.Data(), "data")
	is.Equal(part.Position(), Vector{17, 23})
	is.Equal(part.Velocity(), Vector{3, 5})
	is.Equal(part.Scale(), Vector{7, 11})
	is.Equal(part.Color(), color.RGBA{0x12, 0x23, 0x34, 0x45})
	is.Equal(part.Angle(), 0.0)
	is.Equal(part.Lifetime(), 1500*time.Millisecond)
	is.True(updateCalled)

	now = now.Add(1 * time.Second)
	sys.Update(now)

	is.Equal(part.Position(), Vector{17, 23}.Add(Vector{3, 5}))
	is.Equal(part.Angle(), 0.123)

	now = now.Add(1 * time.Second)
	sys.Update(now)

	is.True(deathCalled)
}

func TestParticle_Kill(t *testing.T) {
	is := is.New(t)

	sys := NewSystem()

	sys.MaxParticles = 1

	sys.LifetimeOverTime = func(d time.Duration, delta time.Duration) time.Duration {
		return 10 * time.Second
	}

	sys.Spawn(1)

	now := time.Now()
	sys.Update(now)

	var part *Particle

	sys.ForEachParticle(func(p *Particle, t NormalizedDuration, delta time.Duration) {
		part = p
	}, now)

	part.Kill()

	now = now.Add(1 * time.Second)
	sys.Update(now)

	is.Equal(sys.NumParticles(), 0)
}
