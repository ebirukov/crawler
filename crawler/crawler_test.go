package crawler

import (
	"reflect"
	"testing"
)

type workerMock struct{}

func (p *workerMock) Shutdown() {}

func (p *workerMock) SubmitTasks(urls []URL) <-chan Result {
	out := make(chan Result)
	go func() {
		for i := 0; i < len(urls); i++ {
			out <- ResultOK
		}
		close(out)
	}()
	return out
}

func _Test_processor_walk(t *testing.T) {
	type args struct {
		urls []URL
	}
	tests := []struct {
		name    string
		args    args
		want    []Result
		wantErr bool
	}{
		{
			name: "nil urls",
			args: args{
				urls: nil,
			},
			want:    []Result{},
			wantErr: false,
		},
		{
			name:    "empty urls",
			args:    args{urls: []URL{}},
			want:    []Result{},
			wantErr: false,
		},
		{
			name:    "one urls",
			args:    args{urls: []URL{"https://google.com"}},
			want:    []Result{ResultOK},
			wantErr: false,
		},
		{
			name:    "two urls",
			args:    args{urls: []URL{"https://google.com", "https://yandex.ru"}},
			want:    []Result{ResultOK, ResultOK},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(&workerMock{}, &MetricMock{})
			got, err := p.Walk(tt.args.urls)
			if (err != nil) != tt.wantErr {
				t.Errorf("walk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("walk() got = %v, want %v", got, tt.want)
			}
		})
	}
}
