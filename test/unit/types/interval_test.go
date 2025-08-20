package types_test

import (
	"testing"

	"github.com/sehwan505/langextract-go/pkg/types"
)

func TestCharInterval_NewCharInterval(t *testing.T) {
	tests := []struct {
		name      string
		start     int
		end       int
		wantError bool
	}{
		{"valid interval", 0, 10, false},
		{"empty interval", 5, 5, false},
		{"negative start", -1, 10, true},
		{"end before start", 10, 5, true},
		{"large interval", 0, 1000000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interval, err := types.NewCharInterval(tt.start, tt.end)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("NewCharInterval() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewCharInterval() unexpected error: %v", err)
				return
			}
			
			if interval.StartPos != tt.start {
				t.Errorf("StartPos = %d, want %d", interval.StartPos, tt.start)
			}
			if interval.EndPos != tt.end {
				t.Errorf("EndPos = %d, want %d", interval.EndPos, tt.end)
			}
		})
	}
}

func TestCharInterval_Length(t *testing.T) {
	tests := []struct {
		name     string
		interval types.CharInterval
		expected int
	}{
		{"empty interval", types.CharInterval{StartPos: 0, EndPos: 0}, 0},
		{"single character", types.CharInterval{StartPos: 0, EndPos: 1}, 1},
		{"multiple characters", types.CharInterval{StartPos: 5, EndPos: 10}, 5},
		{"large interval", types.CharInterval{StartPos: 100, EndPos: 1000}, 900},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.Length(); got != tt.expected {
				t.Errorf("Length() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestCharInterval_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		interval types.CharInterval
		expected bool
	}{
		{"empty interval", types.CharInterval{StartPos: 0, EndPos: 0}, true},
		{"empty interval at position", types.CharInterval{StartPos: 5, EndPos: 5}, true},
		{"non-empty interval", types.CharInterval{StartPos: 0, EndPos: 1}, false},
		{"large interval", types.CharInterval{StartPos: 10, EndPos: 100}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.IsEmpty(); got != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCharInterval_Contains(t *testing.T) {
	interval := types.CharInterval{StartPos: 5, EndPos: 10}
	
	tests := []struct {
		name     string
		pos      int
		expected bool
	}{
		{"before interval", 3, false},
		{"at start", 5, true},
		{"inside interval", 7, true},
		{"at end-1", 9, true},
		{"at end", 10, false},
		{"after interval", 15, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := interval.Contains(tt.pos); got != tt.expected {
				t.Errorf("Contains(%d) = %v, want %v", tt.pos, got, tt.expected)
			}
		})
	}
}

func TestCharInterval_Overlaps(t *testing.T) {
	interval1 := types.CharInterval{StartPos: 5, EndPos: 10}
	
	tests := []struct {
		name      string
		interval2 types.CharInterval
		expected  bool
	}{
		{"no overlap before", types.CharInterval{StartPos: 0, EndPos: 3}, false},
		{"no overlap after", types.CharInterval{StartPos: 12, EndPos: 15}, false},
		{"touching before", types.CharInterval{StartPos: 0, EndPos: 5}, false},
		{"touching after", types.CharInterval{StartPos: 10, EndPos: 15}, false},
		{"partial overlap before", types.CharInterval{StartPos: 3, EndPos: 7}, true},
		{"partial overlap after", types.CharInterval{StartPos: 8, EndPos: 12}, true},
		{"complete overlap", types.CharInterval{StartPos: 6, EndPos: 9}, true},
		{"identical", types.CharInterval{StartPos: 5, EndPos: 10}, true},
		{"containing", types.CharInterval{StartPos: 0, EndPos: 15}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := interval1.Overlaps(tt.interval2); got != tt.expected {
				t.Errorf("Overlaps(%v) = %v, want %v", tt.interval2, got, tt.expected)
			}
		})
	}
}

func TestCharInterval_Union(t *testing.T) {
	tests := []struct {
		name      string
		interval1 types.CharInterval
		interval2 types.CharInterval
		expected  types.CharInterval
	}{
		{"adjacent intervals", types.CharInterval{StartPos: 0, EndPos: 5}, types.CharInterval{StartPos: 5, EndPos: 10}, types.CharInterval{StartPos: 0, EndPos: 10}},
		{"overlapping intervals", types.CharInterval{StartPos: 3, EndPos: 8}, types.CharInterval{StartPos: 6, EndPos: 12}, types.CharInterval{StartPos: 3, EndPos: 12}},
		{"identical intervals", types.CharInterval{StartPos: 5, EndPos: 10}, types.CharInterval{StartPos: 5, EndPos: 10}, types.CharInterval{StartPos: 5, EndPos: 10}},
		{"contained interval", types.CharInterval{StartPos: 2, EndPos: 15}, types.CharInterval{StartPos: 5, EndPos: 10}, types.CharInterval{StartPos: 2, EndPos: 15}},
		{"separate intervals", types.CharInterval{StartPos: 0, EndPos: 3}, types.CharInterval{StartPos: 7, EndPos: 10}, types.CharInterval{StartPos: 0, EndPos: 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.interval1.Union(tt.interval2)
			if got != tt.expected {
				t.Errorf("Union() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCharInterval_Intersection(t *testing.T) {
	tests := []struct {
		name      string
		interval1 types.CharInterval
		interval2 types.CharInterval
		expected  *types.CharInterval
	}{
		{"no overlap", types.CharInterval{StartPos: 0, EndPos: 5}, types.CharInterval{StartPos: 7, EndPos: 10}, nil},
		{"touching", types.CharInterval{StartPos: 0, EndPos: 5}, types.CharInterval{StartPos: 5, EndPos: 10}, nil},
		{"partial overlap", types.CharInterval{StartPos: 3, EndPos: 8}, types.CharInterval{StartPos: 6, EndPos: 12}, &types.CharInterval{StartPos: 6, EndPos: 8}},
		{"identical", types.CharInterval{StartPos: 5, EndPos: 10}, types.CharInterval{StartPos: 5, EndPos: 10}, &types.CharInterval{StartPos: 5, EndPos: 10}},
		{"contained", types.CharInterval{StartPos: 2, EndPos: 15}, types.CharInterval{StartPos: 5, EndPos: 10}, &types.CharInterval{StartPos: 5, EndPos: 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.interval1.Intersection(tt.interval2)
			if tt.expected == nil {
				if got != nil {
					t.Errorf("Intersection() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("Intersection() = nil, want %v", tt.expected)
				} else if *got != *tt.expected {
					t.Errorf("Intersection() = %v, want %v", *got, *tt.expected)
				}
			}
		})
	}
}

func TestCharInterval_String(t *testing.T) {
	tests := []struct {
		name     string
		interval types.CharInterval
		expected string
	}{
		{"empty interval", types.CharInterval{StartPos: 0, EndPos: 0}, "[0:0)"},
		{"single character", types.CharInterval{StartPos: 5, EndPos: 6}, "[5:6)"},
		{"multiple characters", types.CharInterval{StartPos: 10, EndPos: 20}, "[10:20)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTokenInterval_NewTokenInterval(t *testing.T) {
	tests := []struct {
		name      string
		start     int
		end       int
		wantError bool
	}{
		{"valid interval", 0, 10, false},
		{"empty interval", 5, 5, false},
		{"negative start", -1, 10, true},
		{"end before start", 10, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interval, err := types.NewTokenInterval(tt.start, tt.end)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("NewTokenInterval() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewTokenInterval() unexpected error: %v", err)
				return
			}
			
			if interval.StartToken != tt.start {
				t.Errorf("StartToken = %d, want %d", interval.StartToken, tt.start)
			}
			if interval.EndToken != tt.end {
				t.Errorf("EndToken = %d, want %d", interval.EndToken, tt.end)
			}
		})
	}
}

func TestTokenInterval_Length(t *testing.T) {
	tests := []struct {
		name     string
		interval types.TokenInterval
		expected int
	}{
		{"empty interval", types.TokenInterval{StartToken: 0, EndToken: 0}, 0},
		{"single token", types.TokenInterval{StartToken: 0, EndToken: 1}, 1},
		{"multiple tokens", types.TokenInterval{StartToken: 5, EndToken: 10}, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.Length(); got != tt.expected {
				t.Errorf("Length() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestTokenInterval_Contains(t *testing.T) {
	interval := types.TokenInterval{StartToken: 5, EndToken: 10}
	
	tests := []struct {
		name     string
		token    int
		expected bool
	}{
		{"before interval", 3, false},
		{"at start", 5, true},
		{"inside interval", 7, true},
		{"at end-1", 9, true},
		{"at end", 10, false},
		{"after interval", 15, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := interval.Contains(tt.token); got != tt.expected {
				t.Errorf("Contains(%d) = %v, want %v", tt.token, got, tt.expected)
			}
		})
	}
}

func TestTokenInterval_String(t *testing.T) {
	tests := []struct {
		name     string
		interval types.TokenInterval
		expected string
	}{
		{"empty interval", types.TokenInterval{StartToken: 0, EndToken: 0}, "tokens[0:0)"},
		{"single token", types.TokenInterval{StartToken: 5, EndToken: 6}, "tokens[5:6)"},
		{"multiple tokens", types.TokenInterval{StartToken: 10, EndToken: 20}, "tokens[10:20)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}