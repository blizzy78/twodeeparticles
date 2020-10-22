package twodeeparticles

import (
	"image/color"
	"sync"
	"time"
)

// A ParticleSystem simulates a number of particles. Various functions are called to customize the behavior of the particles.
//
// The position of a particle is always relative to its system's origin. In other words, a particle system maintains its
// own frame of reference. Particles are not simulated in "world space." However, when particles are actually drawn on screen,
// the origin of the particle system can be moved freely, thus emulating a simulation in world space.
type ParticleSystem struct {
	// MaxParticles limits the total number of particles being alive at a time. When particles die, new particles may be
	// spawned according to EmissionRateOverTime.
	MaxParticles int

	// DataOverLifetime returns arbitrary data for a particle, over its lifetime. This allows to attach data to the particle
	// and act on it later on. The data returned is not used by the system itself.
	DataOverLifetime ParticleDataOverNormalizedTimeFunc

	// DeathFunc is called when a particle has died. This can be used to clean up the data returned by DataOverLifetime
	// (for example, to return the data back into a pool.)
	DeathFunc ParticleDeathFunc

	// UpdateFunc is called to update a particle during its lifetime. This can be used to Particle.Kill it when certain
	// conditions are met.
	UpdateFunc ParticleVisitFunc

	// EmissionRateOverTime returns the emission rate of the system, in particles/second, over the duration of the system.
	//
	// If EmissionRateOverTime is nil, no particles will spawn.
	EmissionRateOverTime ValueOverTimeFunc

	// EmissionPositionOverTime returns the initial position of a particle that is being spawned, over the duration
	// of the system. The position is measured in arbitrary units (for example, in pixels), and is relative to the
	// system's origin.
	//
	// If EmissionPositionOverTime is nil, particles will spawn at the origin.
	EmissionPositionOverTime VectorOverTimeFunc

	// LifetimeOverTime returns the lifetime of a particle that is being spawned, over the duration of the system.
	// After the duration has passed, the particle will die automatically.
	//
	// If LifetimeOverTime is nil, particles will die after 1 second.
	LifetimeOverTime DurationOverTimeFunc

	// VelocityOverLifetime returns a particle's velocity (direction times speed), in arbitrary units per second,
	// over its lifetime.
	//
	// If VelocityOverLifetime is nil, particles will not move.
	VelocityOverLifetime ParticleVectorOverNormalizedTimeFunc

	// ScaleOverLifetime returns a particle's scale (size multiplier), over its lifetime.
	//
	// If ScaleOverLifetime is nil, particles will use (1.0,1.0).
	ScaleOverLifetime ParticleVectorOverNormalizedTimeFunc

	// ColorOverLifetime returns a particle's color, over its lifetime.
	//
	// If ColorOverLifetime is nil, particles will use color.White.
	ColorOverLifetime ParticleColorOverNormalizedTimeFunc

	// RotationOverLifetime returns a particle's angular velocity, in radians, over its lifetime.
	//
	// If RotationOverLifetime is nil, particles will not rotate.
	RotationOverLifetime ParticleValueOverNormalizedTimeFunc

	initOnce        *sync.Once
	particles       []*Particle
	pool            *sync.Pool
	startTime       time.Time
	lastUpdateTime  time.Time
	particlesToEmit float64
}

// ParticleDeathFunc is a function that is called when p has died.
type ParticleDeathFunc func(p *Particle)

// ValueOverTimeFunc is a function that returns a value after duration d has passed.
// delta is the duration since the last update (for example, the duration since the last GPU frame.)
type ValueOverTimeFunc func(d time.Duration, delta time.Duration) float64

// VectorOverTimeFunc is a function that returns a vector after duration d has passed.
// delta is the duration since the last update (for example, the duration since the last GPU frame.)
type VectorOverTimeFunc func(d time.Duration, delta time.Duration) Vector

// DurationOverTimeFunc is a function that returns a duration after duration d has passed.
// delta is the duration since the last update (for example, the duration since the last GPU frame.)
type DurationOverTimeFunc func(d time.Duration, delta time.Duration) time.Duration

// ParticleValueOverNormalizedTimeFunc is a function that returns a value for p after p's duration t has passed.
// delta is the duration since the last update (for example, the duration since the last GPU frame.)
type ParticleValueOverNormalizedTimeFunc func(p *Particle, t NormalizedDuration, delta time.Duration) float64

// ParticleVectorOverNormalizedTimeFunc is a function that returns a vector for p after p's duration t has passed.
// delta is the duration since the last update (for example, the duration since the last GPU frame.)
type ParticleVectorOverNormalizedTimeFunc func(p *Particle, t NormalizedDuration, delta time.Duration) Vector

// ParticleColorOverNormalizedTimeFunc is a function that returns a color for p after p's duration t has passed.
// delta is the duration since the last update (for example, the duration since the last GPU frame.)
type ParticleColorOverNormalizedTimeFunc func(p *Particle, t NormalizedDuration, delta time.Duration) color.Color

// ParticleDataOverNormalizedTimeFunc is a function that returns arbitrary data for p after p's duration t has passed.
// The data from previous updates is passed as old and may be modified and returned. For the first update, nil is
// passed as the old data. delta is the duration since the last update (for example, the duration since the last
// GPU frame.)
type ParticleDataOverNormalizedTimeFunc func(old interface{}, t NormalizedDuration, delta time.Duration) interface{}

// ParticleVisitFunc is a function that is called for p after p's duration t has passed, when looping over all particles
// in the system using ParticleSystem.ForEachParticle. delta is the duration since the last update (for example,
// the duration since the last GPU frame.)
type ParticleVisitFunc func(p *Particle, t NormalizedDuration, delta time.Duration)

// NormalizedDuration is a normalized duration during a longer duration (for example, during a particle's lifetime.)
// The value is always in the range [0.0,1.0], with 0.0 being the start of the longer duration and 1.0 being the end
// of the longer duration.
type NormalizedDuration float64

// NewParticleSystem returns a new particle system.
func NewParticleSystem() *ParticleSystem {
	s := &ParticleSystem{
		initOnce: &sync.Once{},
		pool:     &sync.Pool{},
	}

	s.pool.New = func() interface{} {
		return newParticle(s)
	}

	return s
}

// Update updates the system. now should usually be time.Now().
func (s *ParticleSystem) Update(now time.Time) {
	s.initOnce.Do(func() {
		s.init(now)
	})

	defer func() {
		s.lastUpdateTime = now
	}()

	for {
		s.removeDeadParticles(now)
		s.spawnParticles(now)
		if !s.updateParticles(now) {
			break
		}
	}
}

func (s *ParticleSystem) init(now time.Time) {
	s.startTime = now
	s.lastUpdateTime = now
}

func (s *ParticleSystem) removeDeadParticles(now time.Time) {
	for i := len(s.particles) - 1; i >= 0; i-- {
		p := s.particles[i]
		if p.alive(now) {
			continue
		}

		s.particles = append(s.particles[:i], s.particles[i+1:]...)
		s.pool.Put(p)

		if s.DeathFunc != nil {
			s.DeathFunc(p)
		}
	}
}

func (s *ParticleSystem) spawnParticles(now time.Time) {
	if s.EmissionRateOverTime == nil {
		return
	}

	d := s.Duration(now)
	delta := now.Sub(s.lastUpdateTime)
	if delta <= 0 {
		return
	}

	s.particlesToEmit += s.EmissionRateOverTime(d, delta) * delta.Seconds()
	for s.particlesToEmit >= 1 {
		s.spawnParticle(now)
		s.particlesToEmit--
	}
}

func (s *ParticleSystem) spawnParticle(now time.Time) {
	if len(s.particles) >= s.MaxParticles {
		return
	}

	p := s.pool.Get().(*Particle)

	p.reset()

	d := s.Duration(now)
	delta := now.Sub(s.lastUpdateTime)
	if s.LifetimeOverTime != nil {
		p.lifetime = s.LifetimeOverTime(d, delta)
	} else {
		p.lifetime = 1 * time.Second
	}
	p.birthTime = now
	p.deathTime = now.Add(p.lifetime)
	p.lastUpdateTime = now

	if s.EmissionPositionOverTime != nil {
		p.position = s.EmissionPositionOverTime(d, delta)
	}

	s.particles = append(s.particles, p)
}

func (s *ParticleSystem) updateParticles(now time.Time) bool {
	needsMorePasses := false
	for _, p := range s.particles {
		p.update(now)

		if !p.alive(now) {
			needsMorePasses = true
		}
	}
	return needsMorePasses
}

// ForEachParticle calls f for each alive particle in the system. now should usually be time.Now().
func (s *ParticleSystem) ForEachParticle(f ParticleVisitFunc, now time.Time) {
	delta := now.Sub(s.lastUpdateTime)
	for _, p := range s.particles {
		d := p.duration(now)
		t := NormalizedDuration(d.Seconds() / p.lifetime.Seconds())
		f(p, t, delta)
	}
}

// Duration returns the duration of the system at now, that is, how long the system has been active.
// now should usually be time.Now().
func (s *ParticleSystem) Duration(now time.Time) time.Duration {
	return now.Sub(s.startTime)
}

// NumParticles returns the number of alive particles.
func (s *ParticleSystem) NumParticles() int {
	return len(s.particles)
}

// Reset kills all alive particles and completely resets the system.
// DeathFunc will be called for all particles that were alive.
func (s *ParticleSystem) Reset() {
	for _, p := range s.particles {
		p.Kill()
	}
	s.removeDeadParticles(time.Now())

	s.initOnce = &sync.Once{}
	s.particles = nil
	s.particlesToEmit = 0.0
}

// Duration converts t to a duration with respect to the longer duration m.
// If t is 0, it will return 0, and if t is 1, it will return m.
func (t NormalizedDuration) Duration(m time.Duration) time.Duration {
	return time.Duration(float64(m.Nanoseconds()) * float64(t))
}
