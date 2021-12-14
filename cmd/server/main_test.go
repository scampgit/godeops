package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounterHandlerFunc(t *testing.T) {
	tests := []struct {
		name       string
		requestURL string
		wantCode   int
	}{
		{
			name:       "test one",
			requestURL: "/update/counter/TestMetric/28",
			wantCode:   200,
		},
		{
			name:       "test two",
			requestURL: "/update/counter/TestMetric/aaa",
			wantCode:   400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.requestURL, nil)
			w := httptest.NewRecorder()

			h := http.HandlerFunc(counterHndlr)

			h.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, res.StatusCode, tt.wantCode)
			defer res.Body.Close()
		})
	}
}

func TestGaugeHandler(t *testing.T) {
	tests := []struct {
		name       string
		requestURL string
		wantCode   int
	}{
		{
			name:       "test one",
			requestURL: "/update/gauge/TestMetric/6464.5",
			wantCode:   200,
		},
		{
			name:       "test two",
			requestURL: "/update/gauge/TestMetric/aaa",
			wantCode:   400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.requestURL, nil)
			w := httptest.NewRecorder()

			h := http.HandlerFunc(gaugeHndlr)

			h.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, res.StatusCode, tt.wantCode)
			defer res.Body.Close()
		})
	}
}

func TestGetBody(t *testing.T) {
	type args struct {
		requestURL string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want2   string
		wantErr bool
	}{
		{
			name: "One",
			args: args{
				requestURL: "/update/gauge/Alloc/10",
			},
			want:    "Alloc",
			want2:   "10",
			wantErr: false,
		},
		{
			name: "Two",
			args: args{
				requestURL: "/update/counter/PollCount/12",
			},
			want:    "PollCount",
			want2:   "12",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.args.requestURL, nil)

			got, got2 := getterReqBody(request)
			assert.Equal(t, got, tt.want)
			assert.Equal(t, got2, tt.want2)
		})
	}
}
