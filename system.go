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

	initOnce        sync.Once
	particles       []*Particle
	pool            sync.Pool
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
type ParticleDataOverNormalizedTimeFunc func(old any, t NormalizedDuration, delta time.Duration) any

// ParticleVisitFunc is a function that is called for p after p's duration t has passed, when looping over all particles
// in the system using ParticleSystem.ForEachParticle. delta is the duration since the last update (for example,
// the duration since the last GPU frame.)
type ParticleVisitFunc func(p *Particle, t NormalizedDuration, delta time.Duration)

// NormalizedDuration is a normalized duration during a longer duration (for example, during a particle's lifetime.)
// The value is always in the range [0.0,1.0], with 0.0 being the start of the longer duration and 1.0 being the end
// of the longer duration.
type NormalizedDuration float64

// NewSystem returns a new particle system.
func NewSystem() *ParticleSystem {
	sys := &ParticleSystem{
		initOnce: sync.Once{},
		pool:     sync.Pool{},
	}

	sys.pool.New = func() any {
		return newParticle(sys)
	}

	return sys
}

// Update updates the system. now should usually be time.Now().
func (sys *ParticleSystem) Update(now time.Time) {
	sys.initOnce.Do(func() {
		sys.init(now)
	})

	defer func() {
		sys.lastUpdateTime = now
	}()

	for {
		sys.removeDeadParticles(now)
		sys.spawnParticles(now)

		if !sys.updateParticles(now) {
			break
		}
	}
}

func (sys *ParticleSystem) init(now time.Time) {
	sys.startTime = now
	sys.lastUpdateTime = now
}

func (sys *ParticleSystem) removeDeadParticles(now time.Time) {
	for idx := len(sys.particles) - 1; idx >= 0; idx-- {
		part := sys.particles[idx]
		if part.alive(now) {
			continue
		}

		sys.particles = append(sys.particles[:idx], sys.particles[idx+1:]...)
		sys.pool.Put(part)

		if sys.DeathFunc != nil {
			sys.DeathFunc(part)
		}
	}
}

func (sys *ParticleSystem) spawnParticles(now time.Time) {
	if sys.EmissionRateOverTime != nil {
		d := sys.Duration(now)
		delta := now.Sub(sys.lastUpdateTime)
		sys.particlesToEmit += sys.EmissionRateOverTime(d, delta) * delta.Seconds()
	}

	for sys.particlesToEmit >= 1 {
		sys.spawnParticle(now)
		sys.particlesToEmit--
	}
}

func (sys *ParticleSystem) spawnParticle(now time.Time) {
	if len(sys.particles) >= sys.MaxParticles {
		return
	}

	part := sys.pool.Get().(*Particle) //nolint:forcetypeassert // we know this is a *Particle

	part.reset()

	dur := sys.Duration(now)
	delta := now.Sub(sys.lastUpdateTime)

	if sys.LifetimeOverTime != nil {
		part.lifetime = sys.LifetimeOverTime(dur, delta)
	} else {
		part.lifetime = 1 * time.Second
	}

	part.birthTime = now
	part.deathTime = now.Add(part.lifetime)
	part.lastUpdateTime = now

	if sys.EmissionPositionOverTime != nil {
		part.position = sys.EmissionPositionOverTime(dur, delta)
	}

	sys.particles = append(sys.particles, part)
}

func (sys *ParticleSystem) updateParticles(now time.Time) bool {
	needsMorePasses := false

	for _, p := range sys.particles {
		p.update(now)

		if !p.alive(now) {
			needsMorePasses = true
		}
	}

	return needsMorePasses
}

// Spawn increases the number of particles to emit on the next Update by num. This can be used
// to instantly spawn a number of particles at any time, regardless of EmissionRateOverTime.
func (sys *ParticleSystem) Spawn(num int) {
	sys.particlesToEmit += float64(num)
}

// ForEachParticle calls fun for each alive particle in the system. now should usually be time.Now().
func (sys *ParticleSystem) ForEachParticle(fun ParticleVisitFunc, now time.Time) {
	delta := now.Sub(sys.lastUpdateTime)

	for _, p := range sys.particles {
		d := p.duration(now)
		t := NormalizedDuration(d.Seconds() / p.lifetime.Seconds())
		fun(p, t, delta)
	}
}

// Duration returns the duration of the system at now, that is, how long the system has been active.
// now should usually be time.Now().
func (sys *ParticleSystem) Duration(now time.Time) time.Duration {
	return now.Sub(sys.startTime)
}

// NumParticles returns the number of alive particles.
func (sys *ParticleSystem) NumParticles() int {
	return len(sys.particles)
}

// Reset kills all alive particles and completely resets the system.
// DeathFunc will be called for all particles that were alive.
func (sys *ParticleSystem) Reset() {
	for _, p := range sys.particles {
		p.Kill()
	}

	sys.removeDeadParticles(time.Now())

	sys.initOnce = sync.Once{}
	sys.particles = nil
	sys.particlesToEmit = 0.0
}

// Duration converts t to a duration with respect to the longer duration m.
// If t is 0, it will return 0, and if t is 1, it will return m.
func (t NormalizedDuration) Duration(m time.Duration) time.Duration {
	return time.Duration(float64(m.Nanoseconds()) * float64(t))
}
