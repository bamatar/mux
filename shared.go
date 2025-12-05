package mux

import (
	"net/http"
	"sync"
)

type responder struct {
	status int
	http.ResponseWriter
}

func (w *responder) WriteHeader(status int) {
	w.status = status
}

func (w *responder) Write([]byte) (int, error) {
	return 0, nil // discard body
}

type pool[T any] struct {
	sp sync.Pool
}

func (p *pool[T]) get() *T {
	if v := p.sp.Get(); v != nil {
		return v.(*T)
	}
	return new(T)
}

func (p *pool[T]) put(x *T) {
	p.sp.Put(x)
}
