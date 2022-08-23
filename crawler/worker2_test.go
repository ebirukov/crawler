package crawler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_PoolV2Shutdown(t *testing.T) {
	actual := make([]Result, 0)
	pool := NewWorkerV2(mockWorkFn(100*time.Millisecond), 0, 0, 500*time.Millisecond, MetricMock{})
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

func Test_PoolV2SubmitTasks(t *testing.T) {

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
			want: []Result{ResultOK, ResultCANCEL},
		},
		{
			name: "сработал таймаут ожидания",
			args: args{
				execTime: 10 * time.Millisecond,
				timeout:  5 * time.Millisecond,
				urls:     []URL{"https://yandex.ru", "https://google.com"},
				poolSize: 4,
			},
			want: []Result{ResultCANCEL, ResultCANCEL},
		},
	}

	//t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual := make([]Result, 0)
			pool := NewWorkerV2(mockWorkFn(tt.args.execTime), tt.args.poolSize, tt.args.timeout, 100*time.Millisecond, MetricMock{})
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
