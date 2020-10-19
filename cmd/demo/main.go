// +build demo

package main

import (
	"fmt"
	_ "image/png"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/blizzy78/twodeeparticles"
	"github.com/fogleman/ease"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// values for Bubbles
const (
	maxParticles = 300

	emissionRate         = 80.0
	emissionRateVariance = 30.0

	moveTime         = 2.0
	moveTimeVariance = 2.0
	fadeOutTime      = 0.15

	startPositionMaxDistance = 20.0

	startSpeed         = 150.0
	startSpeedVariance = 50.0

	startScale       = 0.2
	endScale         = 0.65
	endScaleVariance = 0.3

	minRotationAngle = 100.0
	maxRotationAngle = minRotationAngle * 2.0

	minAlpha = 0.5
)

type game struct {
	dot       *ebiten.Image
	rand      *rand.Rand
	particles *twodeeparticles.ParticleSystem
	drawOpts  *ebiten.DrawImageOptions
	demoIndex int
}

type demo struct {
	label         string
	createFunc    func(rand *rand.Rand) *twodeeparticles.ParticleSystem
	xOriginOffset float64
	yOriginOffset float64
}

type bubbleData struct {
	speed    float64
	alpha    float64
	endScale float64
}

var demos = []demo{
	{"Bubbles", bubbles, 0.5, 0.5},
	{"Fountain", fountain, 0.5, 0.9},
	{"Vortex", vortex, 0.5, 0.5},
}

func main() {
	dot, _, err := ebitenutil.NewImageFromFile("bubble.png")
	if err != nil {
		panic(err)
	}

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))

	g := game{
		dot:       dot,
		rand:      rand,
		particles: demos[0].createFunc(rand),
		drawOpts:  &ebiten.DrawImageOptions{},
	}

	ebiten.SetWindowTitle("twodeeparticles Demo")
	ebiten.SetWindowSize(640, 480)

	_ = ebiten.RunGame(&g)
}

func (g *game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (g *game) Update() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.demoIndex++
		g.demoIndex %= len(demos)

		g.particles = demos[g.demoIndex].createFunc(g.rand)
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		g.particles.Reset()
	}

	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	now := time.Now()
	g.particles.Update(now)

	w, h := screen.Size()
	originX, originY := int(float64(w)*demos[g.demoIndex].xOriginOffset), int(float64(h)*demos[g.demoIndex].yOriginOffset)
	g.particles.ForEachParticle(func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) {
		g.drawParticle(screen, p, t, originX, originY)
	}, now)

	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("Demo: %s (left click for next, right click to reset current)\nParticles: %d\nFPS: %.1f",
			demos[g.demoIndex].label, g.particles.NumParticles(), ebiten.CurrentFPS()),
		10, 10)
	ebitenutil.DebugPrintAt(screen, "github.com/blizzy78/twodeeparticles", 10, h-25)
}

func (g *game) drawParticle(screen *ebiten.Image, p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, originX int, originY int) {
	g.drawOpts.GeoM.Reset()
	g.drawOpts.ColorM.Reset()

	w, h := g.dot.Size()
	g.drawOpts.GeoM.Translate(float64(-w/2), float64(-h/2))

	xScale, yScale := p.Scale()
	g.drawOpts.GeoM.Scale(xScale, yScale)

	g.drawOpts.GeoM.Rotate(p.Angle())

	x, y := p.Position()
	g.drawOpts.GeoM.Translate(x, y)

	g.drawOpts.GeoM.Translate(float64(originX), float64(originY))

	s := t.Duration(p.Lifetime()).Seconds()
	moveTime := p.Lifetime().Seconds() - fadeOutTime

	if bd, ok := p.Data().(*bubbleData); ok {
		if s <= moveTime {
			g.drawOpts.ColorM.Scale(1.0, 1.0, 1.0, bd.alpha*float64(t))
		} else {
			g.drawOpts.ColorM.Scale(1.0, 1.0, 1.0, bd.alpha*(1.0-((s-moveTime)/fadeOutTime)))
		}
	}

	g.drawOpts.Filter = ebiten.FilterLinear

	screen.DrawImage(g.dot, g.drawOpts)
}

func bubbles(rand *rand.Rand) *twodeeparticles.ParticleSystem {
	particleDataPool := &sync.Pool{}
	particleDataPool.New = func() interface{} {
		return &bubbleData{}
	}

	s := twodeeparticles.NewParticleSystem()

	s.MaxParticles = maxParticles

	s.DataOverLifetime = func(old interface{}, t twodeeparticles.NormalizedDuration, delta time.Duration) interface{} {
		if old != nil {
			return old
		}

		data := particleDataPool.Get().(*bubbleData)
		data.speed = startSpeed + (rand.Float64()-0.5)*startSpeedVariance
		data.alpha = minAlpha + rand.Float64()*(1.0-minAlpha)
		data.endScale = endScale + (rand.Float64()-0.5)*endScaleVariance
		return data
	}

	s.DeathFunc = func(p *twodeeparticles.Particle) {
		particleDataPool.Put(p.Data())
	}

	s.EmissionRateOverTime = func(d time.Duration, delta time.Duration) float64 {
		q := float64(int(d.Seconds())%7)/7.0 - 0.5
		v := emissionRateVariance * q
		return emissionRate + v
	}

	s.EmissionPositionOverTime = func(d time.Duration, delta time.Duration) (float64, float64) {
		a := randomValue(0.0, 360.0, rand)
		dx, dy := angleToVector(a)
		return dx * startPositionMaxDistance, dy * startPositionMaxDistance
	}

	s.LifetimeOverTime = func(d time.Duration, delta time.Duration) time.Duration {
		mt := moveTime + (rand.Float64()-0.5)*moveTimeVariance
		return time.Duration((mt+fadeOutTime)*1000.0) * time.Millisecond
	}

	s.VelocityOverLifetime = func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) (float64, float64) {
		data := p.Data().(*bubbleData)

		s := t.Duration(p.Lifetime()).Seconds()
		if s == 0 {
			a := randomValue(0.0, 360.0, rand)
			dx, dy := angleToVector(a)
			return dx * data.speed, dy * data.speed
		}

		moveTime := p.Lifetime().Seconds() - fadeOutTime
		if s > moveTime {
			return 0.0, 0.0
		}

		dx, dy := normalize(p.Velocity())
		m := 1.0 - ease.OutSine(s/moveTime)
		return dx * data.speed * m, dy * data.speed * m
	}

	s.ScaleOverLifetime = func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) (float64, float64) {
		data := p.Data().(*bubbleData)

		s := t.Duration(p.Lifetime()).Seconds()
		if s == 0 {
			return startScale, startScale
		}

		moveTime := p.Lifetime().Seconds() - fadeOutTime
		if s > moveTime {
			m := (1.0-ease.OutSine((s-moveTime)/fadeOutTime))*(data.endScale-startScale) + startScale
			return m, m
		}
		m := ease.OutSine(s/moveTime)*(data.endScale-startScale) + startScale
		return m, m
	}

	return s
}

func fountain(rand *rand.Rand) *twodeeparticles.ParticleSystem {
	s := twodeeparticles.NewParticleSystem()

	s.MaxParticles = 500

	s.EmissionRateOverTime = constant(80.0)
	s.LifetimeOverTime = constantDuration(5 * time.Second)

	s.VelocityOverLifetime = func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) (float64, float64) {
		var vx float64
		var vy float64

		if t == 0 {
			a := 2.0 * math.Pi * randomValue(80.0, 100.0, rand) / 360.0
			s := 450.0 + (rand.Float64()-0.5)*50.0
			dx, dy := angleToVector(a)
			vx, vy = dx*s, dy*s
		} else {
			vx, vy = p.Velocity()
		}

		vy += 30.0 * 9.81 * delta.Seconds()

		return vx, vy
	}

	s.ScaleOverLifetime = particleTwoConstants(0.2, 0.2)

	s.UpdateFunc = func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) {
		if t < 0.1 {
			return
		}

		_, y := p.Position()
		if y < 0 {
			return
		}

		p.Kill()
	}

	return s
}

func vortex(rand *rand.Rand) *twodeeparticles.ParticleSystem {
	s := twodeeparticles.NewParticleSystem()

	s.MaxParticles = 150

	s.EmissionRateOverTime = constant(15.0)
	s.LifetimeOverTime = constantDuration(24 * time.Hour)

	s.EmissionPositionOverTime = func(d time.Duration, delta time.Duration) (float64, float64) {
		a := randomValue(0.0, 360.0, rand)
		dx, dy := angleToVector(a)
		dist := randomValue(140.0, 160.0, rand)
		return dx * dist, dy * dist
	}

	s.VelocityOverLifetime = func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) (float64, float64) {
		if t == 0 {
			dx, dy := normalize(p.Position())
			dx, dy = rotate(dx, dy, 2.0*math.Pi*-90.0/360.0)
			return dx * 200.0, dy * 200.0
		}

		vx, vy := p.Velocity()
		s := magnitude(vx, vy)
		dx, dy := normalize(vx, vy)
		a := randomValue(105.0, 115.0, rand)
		dx, dy = rotate(dx, dy, 2.0*math.Pi*-a/360.0*delta.Seconds())
		return dx * s, dy * s
	}

	s.ScaleOverLifetime = func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) (float64, float64) {
		if t == 0 {
			s := randomValue(0.1, 0.7, rand)
			return s, s
		}

		return p.Scale()
	}

	return s
}

func constant(c float64) twodeeparticles.ValueOverTimeFunc {
	return func(d time.Duration, delta time.Duration) float64 {
		return c
	}
}

func constantDuration(d time.Duration) twodeeparticles.DurationOverTimeFunc {
	return func(dt time.Duration, delta time.Duration) time.Duration {
		return d
	}
}

func particleTwoConstants(c1 float64, c2 float64) twodeeparticles.ParticleTwoValuesOverNormalizedTimeFunc {
	return func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) (float64, float64) {
		return c1, c2
	}
}

func randomValue(min float64, max float64, rand *rand.Rand) float64 {
	return min + rand.Float64()*(max-min)
}

func magnitude(x float64, y float64) float64 {
	return math.Sqrt(x*x + y*y)
}

func normalize(x float64, y float64) (float64, float64) {
	m := magnitude(x, y)
	return x / m, y / m
}

func angleToVector(a float64) (float64, float64) {
	return math.Cos(a), -math.Sin(a)
}

func rotate(x float64, y float64, a float64) (float64, float64) {
	// https://matthew-brett.github.io/teaching/rotation_2d.html
	return x*math.Cos(a) - y*math.Sin(a), x*math.Sin(a) + y*math.Cos(a)
}
