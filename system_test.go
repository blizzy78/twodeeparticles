package twodeeparticles

import (
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestParticleSystem_Reset(t *testing.T) {
	is := is.New(t)

	sys := NewSystem()

	sys.MaxParticles = 1

	sys.LifetimeOverTime = func(d time.Duration, delta time.Duration) time.Duration {
		return 10 * time.Second
	}

	sys.Spawn(1)

	now := time.Now()
	sys.Update(now)

	sys.Reset()

	is.Equal(sys.NumParticles(), 0)
}

func TestParticleSystem_Update_SpawnMoreAfterKill(t *testing.T) {
	is := is.New(t)

	sys := NewSystem()

	sys.MaxParticles = 1

	sys.EmissionRateOverTime = func(d time.Duration, delta time.Duration) float64 {
		return 1.0
	}

	sys.LifetimeOverTime = func(d time.Duration, delta time.Duration) time.Duration {
		return 10 * time.Second
	}

	sys.Spawn(1)

	now := time.Now()
	sys.Update(now)

	killCalled := false
	sys.UpdateFunc = func(p *Particle, t NormalizedDuration, delta time.Duration) {
		if t > 0 {
			killCalled = true

			p.Kill()
		}
	}

	now = now.Add(1 * time.Second)
	sys.Update(now)

	is.Equal(sys.NumParticles(), 1)
	is.True(killCalled)
}

func TestParticleSystem_Spawn(t *testing.T) {
	is := is.New(t)

	sys := NewSystem()

	sys.MaxParticles = 1

	sys.Spawn(1)

	now := time.Now()
	sys.Update(now)

	is.Equal(sys.NumParticles(), 1)
}

func TestNormalizedDuration_Duration(t *testing.T) {
	is := is.New(t)
	is.Equal(NormalizedDuration(0.2).Duration(5000*time.Millisecond), 1000*time.Millisecond)
}
