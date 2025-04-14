package main

import (
	"math/rand/v2"
	"physics-game/physics"
	m "physics-game/physics/math"

	raylib "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	raylib.InitWindow(800, 450, "raylib [core] example - basic window")
	raylib.SetTargetFPS(60)

	camera := raylib.NewCamera3D(
		raylib.Vector3{X: 5, Y: 5, Z: 5},
		raylib.Vector3{X: 1, Y: 1, Z: 1},
		raylib.Vector3{Y: 1},
		90,
		raylib.CameraPerspective,
	)

	plane := physics.NewCollisionPlane(m.Vector3{0.0, 1.0, 0.0}, 0.0)

	entities := make([]Entity, 0)
	for range 5 {
		entity := NewEntity()
		entities = append(entities, entity)
	}

	for !raylib.WindowShouldClose() {
		if raylib.IsKeyDown(raylib.KeyW) {
			entities[0].Collider.GetBody().Velocity = m.Vector3{-10, 0, 0}
		} else if raylib.IsKeyDown(raylib.KeyS) {
			entities[0].Collider.GetBody().Velocity = m.Vector3{10, 0, 0}
		} else if raylib.IsKeyDown(raylib.KeyA) {
			entities[0].Collider.GetBody().Velocity = m.Vector3{0, 0, 10}
		} else if raylib.IsKeyDown(raylib.KeyD) {
			entities[0].Collider.GetBody().Velocity = m.Vector3{0, 0, -10}
		}

		if raylib.IsKeyDown(raylib.KeyE) {
			entity := NewEntity()
			entities = append(entities, entity)
		}

		for _, entity := range entities {
			entity.Update()
		}
		foundContacts, contacts := generateContacts(plane, entities, float64(raylib.GetFrameTime()))
		if foundContacts {
			physics.ResolveContacts(len(contacts)*8, contacts, m.Real(raylib.GetFrameTime()))
		}

		camera.Target = raylib.Vector3Transform(raylib.Vector3{}, toRaylibMatrix(entities[0].Collider.GetBody().GetTransform()))

		raylib.BeginDrawing()
		raylib.ClearBackground(raylib.RayWhite)

		raylib.BeginMode3D(camera)
		raylib.DrawGrid(10, 10)

		for _, entity := range entities {
			entity.Render()
		}

		raylib.EndMode3D()

		raylib.DrawFPS(5, 5)

		raylib.EndDrawing()
	}

	raylib.CloseWindow()
}

type Entity struct {
	Model    raylib.Model
	Color    raylib.Color
	Collider physics.Collider
}

func NewEntity() Entity {
	model := raylib.LoadModelFromMesh(raylib.GenMeshSphere(0.5, 10, 10))

	color := raylib.NewColor(
		uint8(rand.UintN(255)),
		uint8(rand.UintN(255)),
		uint8(rand.UintN(255)),
		255,
	)

	collider := physics.NewCollisionSphere(nil, 0.5)
	collider.Body.SetMass(10)
	collider.Body.Velocity = m.Vector3{0, 10, 0}
	collider.Body.Acceleration = m.Vector3{0, -8, 0}

	var inertia m.Matrix3
	inertia.SetBlockInertiaTensor(&m.Vector3{0.5, 0.5, 0.5}, 8.0)
	collider.Body.SetInertiaTensor(&inertia)

	collider.Body.CalculateDerivedData()
	collider.CalculateDerivedData()

	return Entity{
		Model:    model,
		Color:    color,
		Collider: collider,
	}
}

func (e *Entity) Update() {
	body := e.Collider.GetBody()
	body.Integrate(m.Real(raylib.GetFrameTime()))
	e.Collider.CalculateDerivedData()
}

func (e *Entity) Render() {
	e.Model.Transform = toRaylibMatrix(e.Collider.GetBody().GetTransform())
	raylib.DrawModel(e.Model, raylib.Vector3{}, 1, e.Color)
	raylib.DrawModelWires(e.Model, raylib.Vector3{}, 1, raylib.Red)
}

func toRaylibVector(vec m.Vector3) raylib.Vector3 {
	return raylib.Vector3{
		X: float32(vec[0]),
		Y: float32(vec[1]),
		Z: float32(vec[2]),
	}
}

func toRaylibMatrix(m m.Matrix3x4) raylib.Matrix {
	return raylib.Matrix{
		// First column
		M0: float32(m[0]),
		M1: float32(m[1]),
		M2: float32(m[2]),
		M3: 0.0, // Implicit 0

		// Second column
		M4: float32(m[3]),
		M5: float32(m[4]),
		M6: float32(m[5]),
		M7: 0.0, // Implicit 0

		// Third column
		M8:  float32(m[6]),
		M9:  float32(m[7]),
		M10: float32(m[8]),
		M11: 0.0, // Implicit 0

		// Fourth column
		M12: float32(m[9]),
		M13: float32(m[10]),
		M14: float32(m[11]),
		M15: 1.0, // Implicit 1
	}
}

func generateContacts(plane *physics.CollisionPlane, cubes []Entity, delta float64) (bool, []*physics.Contact) {
	var returnFound bool
	var found bool
	var contacts []*physics.Contact

	for _, cube := range cubes {
		// see if we have a collision with the ground
		found, contacts = cube.Collider.CheckAgainstHalfSpace(plane, contacts)
		if found == true {
			returnFound = true
		}

		// check it against the other cubes. yes this is O(n^2) and not good practice
		for _, otherCube := range cubes {
			if cube == otherCube {
				continue
			}
			found, contacts = physics.CheckForCollisions(cube.Collider, otherCube.Collider, contacts)
			if found == true {
				returnFound = true
			}
		}
	}

	return returnFound, contacts
}
