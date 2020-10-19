package twodeeparticles

import (
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

	isAlive   bool
	data      interface{}
	x         float64
	y         float64
	xVelocity float64
	yVelocity float64
	xScale    float64
	yScale    float64
	angle     float64
}

func newParticle(s *ParticleSystem) *Particle {
	return &Particle{
		system: s,
	}
}

// System returns the particle system that p is a part of.
func (p *Particle) System() *ParticleSystem {
	return p.system
}

// Data returns the arbitrary data that has been assigned to p (see ParticleSystem.DataOverLifetime.)
func (p *Particle) Data() interface{} {
	return p.data
}

// Position returns p's current position, in arbitrary units (for example, in pixels), relative to its
// system's origin.
func (p *Particle) Position() (float64, float64) {
	return p.x, p.y
}

// Velocity returns p's current velocity (direction times speed), in arbitrary units (for example, in pixels)
// per second.
func (p *Particle) Velocity() (float64, float64) {
	return p.xVelocity, p.yVelocity
}

// Scale returns p's current scale (size multiplier).
func (p *Particle) Scale() (float64, float64) {
	return p.xScale, p.yScale
}

// Angle returns p's current rotation angle, in radians.
func (p *Particle) Angle() float64 {
	return p.angle
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
	p.x, p.y = 0, 0
	p.xVelocity, p.yVelocity = 0.0, 0.0
	p.xScale, p.yScale = 1.0, 1.0
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
		p.xVelocity, p.yVelocity = p.system.VelocityOverLifetime(p, t, delta)
	}

	sec := delta.Seconds()
	p.x += p.xVelocity * sec
	p.y += p.yVelocity * sec

	if p.system.ScaleOverLifetime != nil {
		p.xScale, p.yScale = p.system.ScaleOverLifetime(p, t, delta)
	}

	if p.system.RotationOverLifetime != nil {
		p.angle += p.system.RotationOverLifetime(p, t, delta) * delta.Seconds()
		if p.angle > 2.0*math.Pi {
			p.angle -= 2.0 * math.Pi
		} else if p.angle < 0 {
			p.angle += 2.0 * math.Pi
		}
	}
}
