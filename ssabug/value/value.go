package value

type ValueItf[T any] interface {
	Get(func())
}

type Mapper[T any] struct {
	V ValueItf[T]
}

func (m *Mapper[T]) Get(fn func()) {
	m.V.Get(func() {
	})
}

func Map[T any]() ValueItf[T] {
	return &Mapper[T]{}
}
