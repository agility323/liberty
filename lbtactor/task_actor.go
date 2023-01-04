package lbtactor

import (
	"context"
)

type taskWithReturn func() struct{}

func RunTaskActor(ctx context.Context, task taskWithReturn) {
	go func() {
		ch := make(chan<- struct{}, 1)
		select {
		case <-ctx.Done():
		case ch <-task():
		}
	} ()
}
