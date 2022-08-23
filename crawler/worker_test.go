package crawler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func mockWorkFn(execTime time.Duration) WorkerFunc {
	return func(ctx context.Context, url URL) Result {
		select {
		case <-time.After(execTime):
			return ResultOK
		case <-ctx.Done():
			return ResultCANCEL
		}
	}
}

func Test_PoolShutdown(t *testing.T) {
	actual := make([]Result, 0)
	pool := NewWorker(mockWorkFn(100*time.Millisecond), 0, 0, 500*time.Millisecond)
	out := pool.SubmitTasks([]URL{"https://example.com", "https://google.com"})
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
		execTime time.Duration
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
				execTime: 10 * time.Millisecond,
				urls:     []URL{"https://yandex.ru", "https://google.com", "https://example.com"},
				poolSize: 1,
			},
			want: []Result{ResultOK, ResultOK, ResultOK},
		},
		{
			name: "успешное выполнение с параллельном режиме",
			args: args{
				execTime: 10 * time.Millisecond,
				urls:     []URL{"https://yandex.ru", "https://google.com", "https://example.com"},
				poolSize: 2,
			},
			want: []Result{ResultOK, ResultOK, ResultOK},
		},
		{
			name: "сработал таймаут ожидания с очередью на выполнение",
			args: args{
				execTime: 10 * time.Millisecond,
				timeout:  15 * time.Millisecond,
				urls:     []URL{"https://yandex.ru", "https://google.com", "https://example.com"},
				poolSize: 1,
			},
			want: []Result{ResultOK},
		},
		{
			name: "сработал таймаут ожидания",
			args: args{
				execTime: 10 * time.Millisecond,
				timeout:  5 * time.Millisecond,
				urls:     []URL{"https://yandex.ru", "https://google.com"},
				poolSize: 4,
			},
			want: []Result{},
		},
	}

	//t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual := make([]Result, 0)
			pool := NewWorker(mockWorkFn(tt.args.execTime), tt.args.poolSize, tt.args.timeout, 100*time.Millisecond)
			out := pool.SubmitTasks(tt.args.urls)
			assert.Eventuallyf(t, func() bool {
				if r, ok := <-out; ok {
					actual = append(actual, r)
					return !ok
				}
				return true
			}, 50*time.Second, 100*time.Microsecond, "tasks not completed during 5 secord")
			assert.Equal(t, tt.want, actual, "task results")

		})
	}

}
