package common

import (
	"fmt"
	"strconv"
	"testing"
)

func TestCounter_String(t *testing.T) {
	tests := []struct {
		name string
		c    int64
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
			if got := strconv.FormatInt(tt.c, 10); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGauge_String(t *testing.T) {
	tests := []struct {
		name string
		g    float64
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
			if got := fmt.Sprintf("%v", tt.g); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
