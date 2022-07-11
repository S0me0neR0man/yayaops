package common

import (
	"testing"
)

func TestCounter_String(t *testing.T) {
	tests := []struct {
		name string
		c    Counter
		want string
	}{
		{
			name: "test Counter",
			c:    10,
			want: "10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGauge_String(t *testing.T) {
	tests := []struct {
		name string
		g    Gauge
		want string
	}{
		{
			name: "test Gauge",
			g:    10.01,
			want: "10.01",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
