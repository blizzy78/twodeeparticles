package twodeeparticles

import (
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestParticleSystem_Reset(t *testing.T) {
	is := is.New(t)

	s := NewSystem()

	s.MaxParticles = 1

	s.LifetimeOverTime = func(d time.Duration, delta time.Duration) time.Duration {
		return 10 * time.Second
	}

	s.Spawn(1)
	now := time.Now()
	s.Update(now)

	s.Reset()

	is.Equal(s.NumParticles(), 0)
}

func TestParticleSystem_Update_SpawnMoreAfterKill(t *testing.T) {
	is := is.New(t)

	s := NewSystem()

	s.MaxParticles = 1

	s.EmissionRateOverTime = func(d time.Duration, delta time.Duration) float64 {
		return 1.0
	}

	s.LifetimeOverTime = func(d time.Duration, delta time.Duration) time.Duration {
		return 10 * time.Second
	}

	s.Spawn(1)
	now := time.Now()
	s.Update(now)

	killCalled := false
	s.UpdateFunc = func(p *Particle, t NormalizedDuration, delta time.Duration) {
		if t > 0 {
			killCalled = true
			p.Kill()
		}
	}

	now = now.Add(1 * time.Second)
	s.Update(now)

	is.Equal(s.NumParticles(), 1)
	is.True(killCalled)
}

func TestParticleSystem_Spawn(t *testing.T) {
	is := is.New(t)

	s := NewSystem()

	s.MaxParticles = 1

	s.Spawn(1)
	now := time.Now()
	s.Update(now)

	is.Equal(s.NumParticles(), 1)
}

func TestNormalizedDuration_Duration(t *testing.T) {
	is := is.New(t)
	is.Equal(NormalizedDuration(0.2).Duration(5000*time.Millisecond), 1000*time.Millisecond)
}
