package main

import (
	"fmt"
	"image/color"
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

	minAlpha = 0.35
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

var gravity = twodeeparticles.Vector{0.0, 9.81}

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

	s := p.Scale()
	g.drawOpts.GeoM.Scale(s.X, s.Y)

	g.drawOpts.GeoM.Rotate(p.Angle())

	pos := p.Position()
	g.drawOpts.GeoM.Translate(pos.X, pos.Y)

	g.drawOpts.GeoM.Translate(float64(originX), float64(originY))

	_, _, _, a := p.Color().RGBA()
	g.drawOpts.ColorM.Scale(1.0, 1.0, 1.0, float64(a)/65535.0)

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
		data.speed = randomValue(startSpeed-startSpeedVariance/2.0, startSpeed+startSpeedVariance/2.0, rand)
		data.endScale = randomValue(endScale-endScaleVariance/2.0, endScale+endScaleVariance/2.0, rand)
		data.alpha = randomValue(minAlpha, 1.0, rand)
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

	s.EmissionPositionOverTime = func(d time.Duration, delta time.Duration) twodeeparticles.Vector {
		a := randomValue(0.0, 360.0, rand)
		dir := angleToDirection(a)
		return dir.Mul(startPositionMaxDistance)
	}

	s.LifetimeOverTime = func(d time.Duration, delta time.Duration) time.Duration {
		mt := randomValue(moveTime-moveTimeVariance/2.0, moveTime+moveTimeVariance/2.0, rand)
		return time.Duration((mt+fadeOutTime)*1000.0) * time.Millisecond
	}

	s.VelocityOverLifetime = func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) twodeeparticles.Vector {
		data := p.Data().(*bubbleData)

		s := t.Duration(p.Lifetime()).Seconds()
		if s == 0 {
			a := randomValue(0.0, 360.0, rand)
			dir := angleToDirection(a)
			return dir.Mul(data.speed)
		}

		moveTime := p.Lifetime().Seconds() - fadeOutTime
		if s > moveTime {
			return twodeeparticles.ZeroVector
		}

		dir := p.Velocity().Normalize()
		m := 1.0 - ease.OutSine(s/moveTime)
		return dir.Mul(data.speed * m)
	}

	s.ScaleOverLifetime = func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) twodeeparticles.Vector {
		data := p.Data().(*bubbleData)

		s := t.Duration(p.Lifetime()).Seconds()
		if s == 0 {
			return twodeeparticles.Vector{startScale, startScale}
		}

		moveTime := p.Lifetime().Seconds() - fadeOutTime
		if s > moveTime {
			sc := (1.0-ease.OutSine((s-moveTime)/fadeOutTime))*(data.endScale-startScale) + startScale
			return twodeeparticles.Vector{sc, sc}
		}

		sc := ease.OutSine(s/moveTime)*(data.endScale-startScale) + startScale
		return twodeeparticles.Vector{sc, sc}
	}

	s.ColorOverLifetime = func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) color.Color {
		data := p.Data().(*bubbleData)
		s := t.Duration(p.Lifetime()).Seconds()
		moveTime := p.Lifetime().Seconds() - fadeOutTime
		if s <= moveTime {
			return color.RGBA{255, 255, 255, uint8(data.alpha * float64(t) * 255.0)}
		}

		return color.RGBA{255, 255, 255, uint8(data.alpha * (1.0 - ((s - moveTime) / fadeOutTime)) * 255)}
	}

	return s
}

func fountain(rand *rand.Rand) *twodeeparticles.ParticleSystem {
	s := twodeeparticles.NewParticleSystem()

	s.MaxParticles = 500

	s.EmissionRateOverTime = constant(80.0)
	s.LifetimeOverTime = constantDuration(5 * time.Second)

	s.VelocityOverLifetime = func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) twodeeparticles.Vector {
		var v twodeeparticles.Vector

		if t == 0 {
			a := 2.0 * math.Pi * randomValue(80.0, 100.0, rand) / 360.0
			s := randomValue(450.0-25.0, 450.0+25.0, rand)
			dir := angleToDirection(a)
			v = dir.Mul(s)
		} else {
			v = p.Velocity()
		}

		return v.Add(gravity.Mul(30.0 * delta.Seconds()))
	}

	s.ScaleOverLifetime = particleConstantVector(twodeeparticles.Vector{0.2, 0.2})

	s.ColorOverLifetime = func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) color.Color {
		if t == 0 {
			return color.RGBA{255, 255, 255, uint8(randomValue(minAlpha, 1.0, rand) * 255.0)}
		}

		return p.Color()
	}

	s.UpdateFunc = func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) {
		if t < 0.1 || p.Position().Y < 0 {
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

	s.EmissionPositionOverTime = func(d time.Duration, delta time.Duration) twodeeparticles.Vector {
		a := randomValue(0.0, 360.0, rand)
		dir := angleToDirection(a)
		dist := randomValue(140.0, 160.0, rand)
		return dir.Mul(dist)
	}

	s.VelocityOverLifetime = func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) twodeeparticles.Vector {
		if t == 0 {
			dir := p.Position().Normalize()
			dir = rotate(dir, 2.0*math.Pi*-90.0/360.0)
			return dir.Mul(200.0)
		}

		v := p.Velocity()
		s := v.Magnitude()
		dir := v.Normalize()
		a := randomValue(105.0, 115.0, rand)
		dir = rotate(dir, 2.0*math.Pi*-a/360.0*delta.Seconds())
		return dir.Mul(s)
	}

	s.ScaleOverLifetime = func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) twodeeparticles.Vector {
		if t == 0 {
			s := randomValue(0.1, 0.7, rand)
			return twodeeparticles.Vector{s, s}
		}

		return p.Scale()
	}

	s.ColorOverLifetime = func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) color.Color {
		if t == 0 {
			return color.RGBA{255, 255, 255, uint8(randomValue(minAlpha, 1.0, rand) * 255.0)}
		}

		return p.Color()
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

func particleConstantVector(v twodeeparticles.Vector) twodeeparticles.ParticleVectorOverNormalizedTimeFunc {
	return func(p *twodeeparticles.Particle, t twodeeparticles.NormalizedDuration, delta time.Duration) twodeeparticles.Vector {
		return v
	}
}

func randomValue(min float64, max float64, rand *rand.Rand) float64 {
	return min + rand.Float64()*(max-min)
}

func angleToDirection(a float64) twodeeparticles.Vector {
	sin, cos := math.Sincos(a)
	return twodeeparticles.Vector{cos, -sin}
}

func rotate(v twodeeparticles.Vector, a float64) twodeeparticles.Vector {
	// https://matthew-brett.github.io/teaching/rotation_2d.html
	sin, cos := math.Sincos(a)
	return twodeeparticles.Vector{v.X*cos - v.Y*sin, v.X*sin + v.Y*cos}
}
