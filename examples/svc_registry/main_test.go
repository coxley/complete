package main

import (
	"maps"
	"slices"
	"testing"

	"github.com/coxley/complete/cmpcobra"
	"github.com/coxley/complete/cmptest"
)

func TestRegistry(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
		want   []string
	}{
		{
			name:   "services",
			prompt: "svc_registry <TAB>",
			want:   slices.Collect(maps.Keys(services)),
		},
		{
			name:   "server1",
			prompt: "svc_registry server1 -f <TAB>",
			want:   []string{"grpc_addr"},
		},
		{
			name:   "server2",
			prompt: "svc_registry server2 -f <TAB>",
			want:   []string{"grpc_addr"},
		},
		{
			name:   "consumer1",
			prompt: "svc_registry consumer1 -f <TAB>",
			want:   []string{"pubsub_topic", "pubsub_subscription"},
		},
		{
			name:   "consumer2",
			prompt: "svc_registry consumer2 -f <TAB>",
			want:   []string{"pubsub_topic", "pubsub_subscription"},
		},
		{
			name:   "nothing",
			prompt: "svc_registry -f <TAB>",
			want:   []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completer := cmpcobra.New(Command())
			cmptest.Assert(t, completer, tt.prompt, tt.want)
		})
	}
}
