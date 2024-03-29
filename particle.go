package twodeeparticles

import (
	"image/color"
	"math"
	"time"
)

// A Particle is a part of a particle system.
type Particle struct {
	system         *ParticleSystem
	lifetime       time.Duration
	birthTime      time.Time
	deathTime      time.Time
	lastUpdateTime time.Time

	isAlive  bool
	data     any
	position Vector
	velocity Vector
	scale    Vector
	angle    float64
	color    color.Color
}

func newParticle(sys *ParticleSystem) *Particle {
	return &Particle{
		system: sys,
		color:  color.White,
	}
}

// System returns the particle system that p is a part of.
func (p *Particle) System() *ParticleSystem {
	return p.system
}

// Data returns the arbitrary data that has been assigned to p (see ParticleSystem.DataOverLifetime.)
func (p *Particle) Data() any {
	return p.data
}

// Position returns p's current position, in arbitrary units (for example, in pixels), relative to its
// system's origin.
func (p *Particle) Position() Vector {
	return p.position
}

// Velocity returns p's current velocity (direction times speed), in arbitrary units (for example, in pixels)
// per second.
func (p *Particle) Velocity() Vector {
	return p.velocity
}

// Scale returns p's current scale (size multiplier).
func (p *Particle) Scale() Vector {
	return p.scale
}

// Angle returns p's current rotation angle, in radians.
func (p *Particle) Angle() float64 {
	return p.angle
}

// Color returns p's current color.
func (p *Particle) Color() color.Color {
	return p.color
}

// Lifetime returns p's maximum lifetime.
func (p *Particle) Lifetime() time.Duration {
	return p.lifetime
}

// Kill kills p, even if p's lifetime has not yet been exceeded.
func (p *Particle) Kill() {
	p.isAlive = false
}

func (p *Particle) duration(now time.Time) time.Duration {
	return now.Sub(p.birthTime)
}

func (p *Particle) alive(now time.Time) bool {
	return p.isAlive && p.deathTime.After(now)
}

func (p *Particle) reset() {
	p.isAlive = true
	p.data = nil
	p.position = ZeroVector
	p.velocity = ZeroVector
	p.scale = OneVector
	p.color = color.White
}

func (p *Particle) update(now time.Time) {
	defer func() {
		p.lastUpdateTime = now
	}()

	d := p.duration(now)
	delta := now.Sub(p.lastUpdateTime)
	t := NormalizedDuration(d.Seconds() / p.lifetime.Seconds())

	if p.system.UpdateFunc != nil {
		p.system.UpdateFunc(p, t, delta)
	}

	if p.system.DataOverLifetime != nil {
		p.data = p.system.DataOverLifetime(p.data, t, delta)
	}

	if p.system.VelocityOverLifetime != nil {
		p.velocity = p.system.VelocityOverLifetime(p, t, delta)
	}

	sec := delta.Seconds()
	p.position = p.position.Add(p.velocity.Multiply(sec))

	if p.system.ScaleOverLifetime != nil {
		p.scale = p.system.ScaleOverLifetime(p, t, delta)
	}

	if p.system.RotationOverLifetime != nil {
		p.angle += p.system.RotationOverLifetime(p, t, delta) * delta.Seconds()
		if p.angle > 2.0*math.Pi {
			p.angle -= 2.0 * math.Pi
		} else if p.angle < 0 {
			p.angle += 2.0 * math.Pi
		}
	}

	if p.system.ColorOverLifetime != nil {
		p.color = p.system.ColorOverLifetime(p, t, delta)
	}
}
