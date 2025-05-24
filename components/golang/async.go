package main

import (
	"runtime"
)

type AsyncResult[T any] interface {
	Ready() bool
	Get() T
	ResourceDrop()
}

func Await[T any](result AsyncResult[T]) T {
	defer result.ResourceDrop()
	for !result.Ready() {
		runtime.Gosched()
	}
	return result.Get()
}
