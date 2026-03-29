package tools

import (
	"encoding/json"
	"testing"
)

func TestFormatJSON(t *testing.T) {
	data := json.RawMessage(`{"a":1,"b":2}`)
	got := formatJSON(data)
	want := "{\n  \"a\": 1,\n  \"b\": 2\n}"
	if got != want {
		t.Errorf("formatJSON = %q, want %q", got, want)
	}
}

func TestFormatRate(t *testing.T) {
	tests := []struct {
		value  float64
		metric string
		want   string
	}{
		{1.5e9, "bytes", "1.50 Gbps"},
		{2.5e6, "bytes", "2.50 Mbps"},
		{3.5e3, "bytes", "3.50 Kbps"},
		{500, "bytes", "500.00 bps"},
		{1.5e6, "packets", "1.50M"},
		{2.5e3, "packets", "2.50K"},
		{42, "packets", "42.00"},
	}
	for _, tt := range tests {
		got := formatRate(tt.value, tt.metric)
		if got != tt.want {
			t.Errorf("formatRate(%v, %q) = %q, want %q", tt.value, tt.metric, got, tt.want)
		}
	}
}
