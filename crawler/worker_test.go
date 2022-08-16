package crawler

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var mockWorkFn = func(ctx context.Context, url URL) Result {
	select {
	case <-time.After(10 * time.Millisecond):
		return ResultOK
	case <-ctx.Done():
		return ResultFAIL
	}
}

func Test_PoolShutdown(t *testing.T) {
	actual := make([]Result, 0)
	pool := NewWorker(mockWorkFn, 10, 0)
	out := pool.SubmitTasks([]URL{"https://example.com"})
	pool.Shutdown()
	assert.Eventuallyf(t, func() bool {
		if r, ok := <-out; ok {
			actual = append(actual, r)
			return !ok
		}
		return true
	}, 1*time.Second, 100*time.Microsecond, "tasks not completed during 1 secord")
	assert.Equal(t, []Result{}, actual, "task results")
}

func Test_PoolSubmitTasks(t *testing.T) {

	type args struct {
		timeout  time.Duration
		urls     []URL
		poolSize int
	}

	tests := []struct {
		name string
		args args
		want []Result
	}{
		{
			name: "nil tasks",
			args: args{
				urls:     nil,
				poolSize: 1,
			},
			want: []Result{},
		},
		{
			name: "empty tasks",
			args: args{
				urls:     []URL{},
				poolSize: 1,
			},
			want: []Result{},
		},
		{
			name: "успешное выполнение с однопотоковом режиме",
			args: args{
				urls:     []URL{"https://yandex.ru", "https://google.com", "https://example.com"},
				poolSize: 1,
			},
			want: []Result{ResultOK, ResultOK, ResultOK},
		},
		{
			name: "успешное выполнение с параллельном режиме",
			args: args{
				urls:     []URL{"https://yandex.ru", "https://google.com", "https://example.com"},
				poolSize: 2,
			},
			want: []Result{ResultOK, ResultOK, ResultOK},
		},
		{
			name: "сработал таймаут ожидания",
			args: args{
				timeout:  15 * time.Millisecond,
				urls:     []URL{"https://yandex.ru", "https://google.com", "https://example.com"},
				poolSize: 1,
			},
			want: []Result{ResultOK, ResultFAIL, ResultFAIL},
		},
	}

	t.Parallel()
	for _, tt := range tests {
		tt := tt
		fmt.Println(tt.name)
		t.Run(tt.name, func(t *testing.T) {
			actual := make([]Result, 0)
			pool := NewWorker(mockWorkFn, tt.args.poolSize, tt.args.timeout)
			out := pool.SubmitTasks(tt.args.urls)
			assert.Eventuallyf(t, func() bool {
				if r, ok := <-out; ok {
					actual = append(actual, r)
					return !ok
				}
				return true
			}, 5*time.Second, 100*time.Microsecond, "tasks not completed during 5 secord")
			assert.Equal(t, tt.want, actual, "task results")

		})
	}

}
