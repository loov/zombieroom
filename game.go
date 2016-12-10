package main

import (
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"

	"github.com/loov/zombieroom/g"
)

const (
	ZeroLayer = CollisionLayer(1 << iota)
	PlayerLayer
	HammerLayer
	ZombieLayer
)

type Game struct {
	Assets *Assets

	Room    *Room
	Players []*Player
	Zombies []*Zombie

	Clock float64
}

func NewGame() *Game {
	game := &Game{}

	game.Assets = NewAssets()

	game.Room = NewRoom()

	game.Players = append(game.Players, NewPlayer(&Keyboard_1))
	game.Players = append(game.Players, NewPlayer(&Keyboard_0))

	for i := 0; i < 10; i++ {
		game.Zombies = append(game.Zombies, NewZombie(game.Room.Bounds))
	}

	return game
}

func (game *Game) Update(window *glfw.Window, now float64) {
	dt := float32(now - game.Clock)
	game.Clock = now

	game.Assets.Reload()

	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	width, height := window.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))

	screenRatio := float32(height) / float32(width)
	roomSize := game.Room.Bounds.Size()

	var screenSize g.V2
	roomRatio := roomSize.Y / roomSize.X

	if screenRatio < roomRatio {
		screenSize.Y = roomSize.Y + 2
		screenSize.X = screenSize.Y / screenRatio
	} else {
		screenSize.X = roomSize.X + 2
		screenSize.Y = screenSize.X * screenRatio
	}

	gl.Ortho(
		float64(-screenSize.X/2),
		float64(screenSize.X/2),
		float64(-screenSize.Y/2),
		float64(screenSize.Y/2),
		10, -10)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.MULTISAMPLE)
	gl.Enable(gl.ALPHA_TEST)

	{
		// list all entities
		entities := []*Entity{}
		for _, player := range game.Players {
			entities = append(entities, player.Entities()...)
		}

		for _, zombie := range game.Zombies {
			entities = append(entities, zombie.Entities()...)
		}

		// reset entities
		for _, entity := range entities {
			entity.ResetForces()
		}

		// update survivors and hammers
		for _, player := range game.Players {
			player.UpdateInput(window)
			player.Update(dt)
		}

		// update zombies
		for _, zombie := range game.Zombies {
			zombie.Update(dt)
		}

		// update collision info
		HandleCollisions(entities)

		// integrate forces
		for _, entity := range entities {
			entity.IntegrateForces(dt)
		}

		// apply constraints
		for _, player := range game.Players {
			player.ApplyConstraints(game.Room.Bounds)
		}

		// respawn dead zombies
		for _, zombie := range game.Zombies {
			zombie.Respawn(game.Room.Bounds)
		}

		game.Room.Render(game)

		for _, zombie := range game.Zombies {
			zombie.Render(game)
		}

		for _, player := range game.Players {
			player.Render(game)
		}
	}
}

func (game *Game) Unload() {
	game.Assets.Unload()
}
