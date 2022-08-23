package crawler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockContent struct {
	*bytes.Buffer
}

func (m mockContent) Close() error {
	m.Buffer.Reset()
	return nil
}

func NewContent(s string) *mockContent {
	return &mockContent{Buffer: bytes.NewBufferString(s)}
}

func TestResult_content(t *testing.T) {

	tests := []struct {
		name    string
		r       Result
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "empty content OK status code",
			r: Result{
				Body:       nil,
				StatusCode: 200,
			},
			want:    "",
			wantErr: assert.NoError,
		},
		{
			name: "empty content failed status code",
			r: Result{
				Body:       nil,
				StatusCode: 500,
			},
			want:    "",
			wantErr: assert.NoError,
		},
		{
			name: "not empty content",
			r: Result{
				Body:       NewContent("content"),
				StatusCode: 200,
			},
			want:    "content",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.content()
			if !tt.wantErr(t, err, fmt.Sprintf("content()")) {
				return
			}
			assert.Equalf(t, tt.want, got, "content()")
		})
	}
}

func Test_httpClient_Request(t *testing.T) {
	testDummy := func(w http.ResponseWriter, r *http.Request) {
		param := r.FormValue("p")
		switch param {
		case "timeout":
			time.Sleep(1 * time.Second)
			w.WriteHeader(http.StatusGatewayTimeout)
		case "5xx":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			//w.Write(body)
			w.WriteHeader(http.StatusOK)
		}
		return
	}
	ts := httptest.NewServer(http.HandlerFunc(testDummy))

	canceledCtx, _ := context.WithDeadline(context.Background(), time.Now())
	type fields struct {
		client http.Client
	}
	type args struct {
		ctx context.Context
		url URL
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Result
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "timeout request",
			fields: fields{
				http.Client{Timeout: 100 * time.Millisecond},
			},
			args: args{
				ctx: context.Background(),
				url: URL(fmt.Sprintf("%s?p=%s", ts.URL, "timeout")),
			},
			want:    ResultFAIL,
			wantErr: assert.Error,
		},
		{
			name: "success request",
			fields: fields{
				http.Client{},
			},
			args: args{
				ctx: context.Background(),
				url: URL(ts.URL),
			},
			want:    ResultOK,
			wantErr: assert.NoError,
		},
		{
			name: "5xx error request",
			fields: fields{
				http.Client{},
			},
			args: args{
				ctx: context.Background(),
				url: URL(fmt.Sprintf("%s?p=%s", ts.URL, "5xx")),
			},
			want:    ResultFAIL,
			wantErr: assert.NoError,
		},
		{
			name: "cancel request",
			fields: fields{
				http.Client{},
			},
			args: args{
				ctx: canceledCtx,
				url: URL(fmt.Sprintf("%s?p=%s", ts.URL, "5xx")),
			},
			want:    ResultCANCEL,
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := WorkerHandler(tt.fields.client, MetricMock{})
			got := handler(tt.args.ctx, tt.args.url)
			assert.Equalf(t, tt.want, got, "handler(%v, %v)", tt.args.ctx, tt.args.url)
		})
	}
}
