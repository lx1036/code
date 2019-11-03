package internal

type spinlock struct{ lock uintptr }

type noteQueue struct {
	mu    spinlock
	notes []interface{}
}
