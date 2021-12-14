package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func handlers() http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return r
}

func Test_SendData(t *testing.T) {	
	type fields struct {
		name     string
		typename string
		value    float64
	}
	tests := []struct {
		name    string
		fields  fields
		client  *http.Client
		want    bool
		wantErr bool
	}{
		{
			name:    "test one",
			fields:  fields{name: "Alloc", typename: "gauge", value: 1.5},
			client:  &http.Client{Timeout: 2 * time.Second},
			want:    true,
			wantErr: false,
		},
		{
			name:    "test two",
			fields:  fields{name: "PollCounter", typename: "counter", value: 1},
			client:  &http.Client{Timeout: 2 * t	ime.Second},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := metric{
				name:     tt.fields.name,
				typename: tt.fields.typename,
				value:    tt.fields.value,
			}

			l, err := net.Listen("tcp", "localhost:8080")
			if err != nil {
				log.Fatal(err)
			}
			srv := httptest.NewUnstartedServer(handlers())
			srv.Listener.Close()
			srv.Listener = l
			srv.Start()

			defer srv.Close()

			got, err := m.senderDatas(tt.client)
			if !tt.wantErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			assert.Equal(t, got, tt.want)
		})
	}
}
