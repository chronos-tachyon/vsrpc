package vsrpc

import (
	"context"
)

func Watch(ctx context.Context, callback func(), body func() error) error {
	doneCh := ctx.Done()
	taskCh := make(chan void, 0)
	exitCh := make(chan void, 0)

	go func() {
		defer close(exitCh)
		select {
		case <-taskCh:
			return

		case <-doneCh:
			callback()
			return
		}
	}()

	defer func() {
		select {
		case <-taskCh:
			// pass
		default:
			close(taskCh)
		}
		<-exitCh
	}()

	err := body()
	close(taskCh)
	return err
}
