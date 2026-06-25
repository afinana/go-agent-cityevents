package main

import "testing"

func TestShouldRunIngestion(t *testing.T) {
	tests := []struct {
		name        string
		skipIngest  string
		existingCnt int
		want        bool
	}{
		{name: "ingest when explicitly enabled", skipIngest: "false", existingCnt: 10, want: true},
		{name: "skip when data already exists", skipIngest: "true", existingCnt: 10, want: false},
		{name: "ingest when collection is empty", skipIngest: "true", existingCnt: 0, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRunIngestion(tt.skipIngest, tt.existingCnt); got != tt.want {
				t.Fatalf("shouldRunIngestion(%q, %d) = %v, want %v", tt.skipIngest, tt.existingCnt, got, tt.want)
			}
		})
	}
}
