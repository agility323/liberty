package lbtactor

import (
	"context"

	"github.com/agility323/liberty/lbtutil"
)

type taskWithReturn func() struct{}

func RunTaskActor(ctx context.Context, name string, task taskWithReturn) {
	go func() {
		defer lbtutil.Recover("RunTaskActor." + name, nil)

		ch := make(chan<- struct{}, 1)
		select {
		case <-ctx.Done():
		case ch <-task():
		}
	}()
}
