package wopo_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DmitySH/wopo"
	"github.com/stretchr/testify/require"
)

func ExamplePool() {
	h := func(ctx context.Context, x int) (int, error) {
		return x * x, nil
	}
	p := wopo.NewPool(
		h,
		wopo.WithWorkerCount[int, int](1),
		wopo.WithResultBufferSize[int, int](3),
		wopo.WithResultBufferSize[int, int](3),
	)

	p.Start()

	ctx := context.Background()
	go func() {
		for i := 0; i < 3; i++ {
			p.PushTask(ctx, i)
		}
		p.Stop()
	}()

	for res := range p.Results() {
		fmt.Println(res.Data, res.Err)
	}
	// Output:
	// 0 <nil>
	// 1 <nil>
	// 4 <nil>
}

type behavior string

const (
	success = behavior("success")
	err     = behavior("error")
	panica  = behavior("panic")
)

func handler(_ context.Context, beh behavior) (int, error) {
	switch beh {
	case success:
		return 5, nil
	case err:
		return 0, errors.New("some error")
	case panica:
		panic("some panic")
	default:
		panic("unknown behavior")
	}
}

func TestDefaultPool(t *testing.T) {
	t.Parallel()

	p := wopo.NewPool(handler)
	ctx := context.Background()

	p.Start()
	resCh := p.Results()

	t.Run("success", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			p.PushTask(ctx, success)
			res := <-resCh

			require.NoError(t, res.Err)
			require.Equal(t, 5, res.Data)
		}
	})

	t.Run("with_error", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			p.PushTask(ctx, err)
			res := <-resCh

			require.ErrorContains(t, res.Err, "some error")
		}
	})

	t.Run("with_panic", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			p.PushTask(ctx, panica)
			res := <-resCh

			require.ErrorContains(t, res.Err, "recovered from panic")
		}
	})

	p.Stop()
	<-p.Results()
}

func TestBufferedPool(t *testing.T) {
	t.Parallel()

	p := wopo.NewPool(handler,
		wopo.WithTaskBufferSize[behavior, int](1),
		wopo.WithResultBufferSize[behavior, int](1),
	)
	ctx := context.Background()

	resCh := p.Results()

	p.PushTask(ctx, success)
	p.Start()
	p.PushTask(ctx, err)

	res := <-resCh

	require.NoError(t, res.Err)
	require.Equal(t, 5, res.Data)

	res = <-resCh
	require.ErrorContains(t, res.Err, "some error")

	p.Stop()
	<-p.Results()
}
