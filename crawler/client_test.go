package crawler

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockContent struct {
	*bytes.Buffer
	io.Closer
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
