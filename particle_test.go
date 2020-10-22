package twodeeparticles

import (
	"image/color"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestParticle_System(t *testing.T) {
	is := is.New(t)
	s := NewParticleSystem()
	p := newParticle(s)
	is.Equal(p.System(), s)
}

func TestParticle_Update(t *testing.T) {
	is := is.New(t)

	s := NewParticleSystem()

	s.MaxParticles = 1

	s.EmissionRateOverTime = func(d time.Duration, delta time.Duration) float64 {
		return 10.0
	}

	s.LifetimeOverTime = func(d time.Duration, delta time.Duration) time.Duration {
		return 1500 * time.Millisecond
	}

	s.DataOverLifetime = func(old interface{}, t NormalizedDuration, delta time.Duration) interface{} {
		return "data"
	}

	s.EmissionPositionOverTime = func(d time.Duration, delta time.Duration) Vector {
		return Vector{17, 23}
	}

	s.VelocityOverLifetime = func(p *Particle, t NormalizedDuration, delta time.Duration) Vector {
		return Vector{3, 5}
	}

	s.ScaleOverLifetime = func(p *Particle, t NormalizedDuration, delta time.Duration) Vector {
		return Vector{7, 11}
	}

	s.ColorOverLifetime = func(p *Particle, t NormalizedDuration, delta time.Duration) color.Color {
		return color.RGBA{0x12, 0x23, 0x34, 0x45}
	}

	s.RotationOverLifetime = func(p *Particle, t NormalizedDuration, delta time.Duration) float64 {
		return 0.123
	}

	updateCalled := false
	s.UpdateFunc = func(part *Particle, t NormalizedDuration, delta time.Duration) {
		updateCalled = true
	}

	deathCalled := false
	s.DeathFunc = func(p *Particle) {
		deathCalled = true
	}

	now := time.Now()
	s.Update(now)
	now = now.Add(1 * time.Second)
	s.Update(now)

	var p *Particle
	s.ForEachParticle(func(part *Particle, t NormalizedDuration, delta time.Duration) {
		p = part
	}, now)

	is.Equal(p.Data(), "data")
	is.Equal(p.Position(), Vector{17, 23})
	is.Equal(p.Velocity(), Vector{3, 5})
	is.Equal(p.Scale(), Vector{7, 11})
	is.Equal(p.Color(), color.RGBA{0x12, 0x23, 0x34, 0x45})
	is.Equal(p.Angle(), 0.0)
	is.Equal(p.Lifetime(), 1500*time.Millisecond)
	is.True(updateCalled)

	now = now.Add(1 * time.Second)
	s.Update(now)

	is.Equal(p.Position(), Vector{17, 23}.Add(Vector{3, 5}))
	is.Equal(p.Angle(), 0.123)

	now = now.Add(1 * time.Second)
	s.Update(now)

	is.True(deathCalled)
}

func TestParticle_Kill(t *testing.T) {
	is := is.New(t)

	s := NewParticleSystem()

	s.MaxParticles = 1

	spawnMore := true
	s.EmissionRateOverTime = func(d time.Duration, delta time.Duration) float64 {
		if !spawnMore {
			return 0.0
		}
		return 10.0
	}

	s.LifetimeOverTime = func(d time.Duration, delta time.Duration) time.Duration {
		return 10 * time.Second
	}

	now := time.Now()
	s.Update(now)
	now = now.Add(1 * time.Second)
	s.Update(now)

	var p *Particle
	s.ForEachParticle(func(part *Particle, t NormalizedDuration, delta time.Duration) {
		p = part
	}, now)

	p.Kill()

	spawnMore = false
	now = now.Add(1 * time.Second)
	s.Update(now)

	is.Equal(s.NumParticles(), 0)
}
