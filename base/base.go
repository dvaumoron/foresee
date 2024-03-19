package base

type Addable interface {
	Numeric | ~string
}

type Boolean interface {
	~bool
}

type Channel[T any] interface {
	~chan T
}

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Numeric interface {
	Integer | ~float32 | ~float64 | ~complex64 | ~complex128
}

type Receiver[T any] interface {
	~<-chan T
}

type Sender[T any] interface {
	~chan<- T
}
