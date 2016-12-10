package main

import (
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"

	"github.com/loov/zombieroom/g"
)

type Player struct {
	Controller Controller
	Updater    ControllerUpdater

	Survivor Entity
	Hammer   Hammer
}

func (player *Player) Entities() []*Entity {
	xs := []*Entity{}
	xs = append(xs, player.Survivor.Entities()...)
	xs = append(xs, player.Hammer.Entities()...)

	return xs
}

type Hammer struct {
	Entity

	Tension float32
	Length  float32

	NormalLength      float32
	MaxLength         float32
	TensionMultiplier float32
	VelocityDampening float32
}

func NewPlayer(updater ControllerUpdater) *Player {
	player := &Player{}
	player.Updater = updater

	player.Survivor.Radius = 0.5
	player.Survivor.Mass = 1.0
	player.Survivor.Elasticity = 0.2
	player.Survivor.Dampening = 0.999

	player.Survivor.CollisionLayer = PlayerLayer
	player.Survivor.CollisionMask = HammerLayer

	player.Hammer.Mass = 0.05
	player.Hammer.Elasticity = 0.4
	player.Hammer.Radius = 0.4

	player.Hammer.NormalLength = 2
	player.Hammer.MaxLength = 3
	player.Hammer.TensionMultiplier = 20
	player.Hammer.Dampening = 0.999

	player.Hammer.CollisionLayer = HammerLayer
	player.Hammer.CollisionMask = 0

	return player
}

func (player *Player) UpdateInput(window *glfw.Window) {
	// update input
	player.Updater.Update(&player.Controller, window)
}

func (player *Player) Update(dt float32) {
	const MovementForce = 30

	survivor, hammer := &player.Survivor, &player.Hammer

	{ // add survivor movement forces
		in := player.Controller.Inputs[0]

		if in.Direction.Length() > 0.001 {
			survivor.Force = in.Direction.Scale(MovementForce)
			// todo scale lateral movement

			lateral := in.Direction.Rotate90c().Normalize()
			scale := lateral.Dot(survivor.Velocity)

			survivor.AddForce(lateral.Scale(-scale * 4.0))

			hammer.Dampening = 0.999
		} else {
			survivor.AddForce(survivor.Velocity.Normalize().Negate().Scale(MovementForce))

			hammer.Dampening = 0.98
		}
	}

	{ // hammer forces
		offset := hammer.Position.Sub(survivor.Position)
		length := offset.Length()

		delta := g.Max(length-hammer.NormalLength, 0)
		//delta := length - hammer.NormalLength
		tension := delta * hammer.TensionMultiplier

		pull := g.V2{
			offset.X * tension / (length + 1),
			offset.Y * tension / (length + 1),
		}

		if delta > hammer.MaxLength {
			survivor.AddForce(pull)
			survivor.AddForce(pull)
		} else {
			hammer.AddForce(pull.Negate())
			survivor.AddForce(pull)
		}
	}
}

func (player *Player) ApplyConstraints(bounds g.Rect) {
	survivor, hammer := &player.Survivor, &player.Hammer

	g.EnforceInside(&survivor.Position, &survivor.Velocity, bounds, survivor.Elasticity)
	g.EnforceInside(&hammer.Position, &hammer.Velocity, bounds, hammer.Elasticity)

	dist := hammer.Position.Sub(survivor.Position)
	if n := dist.Length(); n > hammer.MaxLength {
		hammer.Position = survivor.Position.AddScale(dist, hammer.MaxLength/n)
	}
}

func (player *Player) Render(game *Game) {
	survivor, hammer := &player.Survivor, &player.Hammer

	{
		rope := game.Assets.TextureRepeat("assets/rope.png")
		rope.Line(
			hammer.Position,
			survivor.Position,
			hammer.Radius/2)
	}

	rotation := -(survivor.Position.Sub(hammer.Position).Angle() - g.Tau/4)

	gl.PushMatrix()
	{
		gl.Translatef(survivor.Position.X, survivor.Position.Y, 0)

		gl.Rotatef(g.RadToDeg(rotation), 0, 0, -1)

		tex := game.Assets.TextureRepeat("assets/player.png")
		tex.Draw(g.NewCircleRect(survivor.Radius))
	}
	gl.PopMatrix()

	gl.PushMatrix()
	{
		gl.Translatef(hammer.Position.X, hammer.Position.Y, 0)

		gl.Rotatef(g.RadToDeg(rotation), 0, 0, -1)

		tex := game.Assets.TextureRepeat("assets/hammer.png")
		tex.Draw(g.NewCircleRect(hammer.Radius))
	}
	gl.PopMatrix()
}
