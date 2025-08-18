package types

import (
	"testing"
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
			interval, err := NewCharInterval(tt.start, tt.end)
			
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
		interval CharInterval
		expected int
	}{
		{"empty interval", CharInterval{0, 0}, 0},
		{"single character", CharInterval{0, 1}, 1},
		{"multiple characters", CharInterval{5, 10}, 5},
		{"large interval", CharInterval{100, 1000}, 900},
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
		interval CharInterval
		expected bool
	}{
		{"empty interval", CharInterval{0, 0}, true},
		{"empty interval at position", CharInterval{5, 5}, true},
		{"non-empty interval", CharInterval{0, 1}, false},
		{"large interval", CharInterval{10, 100}, false},
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
	interval := CharInterval{5, 10}
	
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
	interval1 := CharInterval{5, 10}
	
	tests := []struct {
		name      string
		interval2 CharInterval
		expected  bool
	}{
		{"no overlap before", CharInterval{0, 3}, false},
		{"no overlap after", CharInterval{12, 15}, false},
		{"touching before", CharInterval{0, 5}, false},
		{"touching after", CharInterval{10, 15}, false},
		{"partial overlap before", CharInterval{3, 7}, true},
		{"partial overlap after", CharInterval{8, 12}, true},
		{"complete overlap", CharInterval{6, 9}, true},
		{"identical", CharInterval{5, 10}, true},
		{"containing", CharInterval{0, 15}, true},
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
		interval1 CharInterval
		interval2 CharInterval
		expected  CharInterval
	}{
		{"adjacent intervals", CharInterval{0, 5}, CharInterval{5, 10}, CharInterval{0, 10}},
		{"overlapping intervals", CharInterval{3, 8}, CharInterval{6, 12}, CharInterval{3, 12}},
		{"identical intervals", CharInterval{5, 10}, CharInterval{5, 10}, CharInterval{5, 10}},
		{"contained interval", CharInterval{2, 15}, CharInterval{5, 10}, CharInterval{2, 15}},
		{"separate intervals", CharInterval{0, 3}, CharInterval{7, 10}, CharInterval{0, 10}},
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
		interval1 CharInterval
		interval2 CharInterval
		expected  *CharInterval
	}{
		{"no overlap", CharInterval{0, 5}, CharInterval{7, 10}, nil},
		{"touching", CharInterval{0, 5}, CharInterval{5, 10}, nil},
		{"partial overlap", CharInterval{3, 8}, CharInterval{6, 12}, &CharInterval{6, 8}},
		{"identical", CharInterval{5, 10}, CharInterval{5, 10}, &CharInterval{5, 10}},
		{"contained", CharInterval{2, 15}, CharInterval{5, 10}, &CharInterval{5, 10}},
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
		interval CharInterval
		expected string
	}{
		{"empty interval", CharInterval{0, 0}, "[0:0)"},
		{"single character", CharInterval{5, 6}, "[5:6)"},
		{"multiple characters", CharInterval{10, 20}, "[10:20)"},
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
			interval, err := NewTokenInterval(tt.start, tt.end)
			
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
		interval TokenInterval
		expected int
	}{
		{"empty interval", TokenInterval{0, 0}, 0},
		{"single token", TokenInterval{0, 1}, 1},
		{"multiple tokens", TokenInterval{5, 10}, 5},
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
	interval := TokenInterval{5, 10}
	
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
		interval TokenInterval
		expected string
	}{
		{"empty interval", TokenInterval{0, 0}, "tokens[0:0)"},
		{"single token", TokenInterval{5, 6}, "tokens[5:6)"},
		{"multiple tokens", TokenInterval{10, 20}, "tokens[10:20)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}