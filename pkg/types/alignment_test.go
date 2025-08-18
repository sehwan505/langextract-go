package types

import (
	"testing"
)

func TestAlignmentStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   AlignmentStatus
		expected string
	}{
		{"none", AlignmentNone, "none"},
		{"exact", AlignmentExact, "exact"},
		{"fuzzy", AlignmentFuzzy, "fuzzy"},
		{"semantic", AlignmentSemantic, "semantic"},
		{"partial", AlignmentPartial, "partial"},
		{"approximate", AlignmentApproximate, "approximate"},
		{"unknown", AlignmentStatus(999), "unknown(999)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAlignmentStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   AlignmentStatus
		expected bool
	}{
		{"none", AlignmentNone, true},
		{"exact", AlignmentExact, true},
		{"fuzzy", AlignmentFuzzy, true},
		{"semantic", AlignmentSemantic, true},
		{"partial", AlignmentPartial, true},
		{"approximate", AlignmentApproximate, true},
		{"invalid negative", AlignmentStatus(-1), false},
		{"invalid large", AlignmentStatus(999), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.expected {
				t.Errorf("IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAlignmentStatus_Quality(t *testing.T) {
	tests := []struct {
		name     string
		status   AlignmentStatus
		expected int
	}{
		{"none", AlignmentNone, 0},
		{"exact", AlignmentExact, 100},
		{"fuzzy", AlignmentFuzzy, 80},
		{"semantic", AlignmentSemantic, 60},
		{"partial", AlignmentPartial, 40},
		{"approximate", AlignmentApproximate, 20},
		{"unknown", AlignmentStatus(999), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Quality(); got != tt.expected {
				t.Errorf("Quality() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestParseAlignmentStatus(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  AlignmentStatus
		wantError bool
	}{
		{"none", "none", AlignmentNone, false},
		{"exact", "exact", AlignmentExact, false},
		{"fuzzy", "fuzzy", AlignmentFuzzy, false},
		{"semantic", "semantic", AlignmentSemantic, false},
		{"partial", "partial", AlignmentPartial, false},
		{"approximate", "approximate", AlignmentApproximate, false},
		{"uppercase", "EXACT", AlignmentExact, false},
		{"mixed case", "Fuzzy", AlignmentFuzzy, false},
		{"invalid", "invalid", AlignmentNone, true},
		{"empty", "", AlignmentNone, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAlignmentStatus(tt.input)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("ParseAlignmentStatus() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("ParseAlignmentStatus() unexpected error: %v", err)
				return
			}
			
			if got != tt.expected {
				t.Errorf("ParseAlignmentStatus() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewAlignmentResult(t *testing.T) {
	tests := []struct {
		name       string
		status     AlignmentStatus
		confidence float64
		score      float64
		method     string
		wantError  bool
	}{
		{"valid result", AlignmentExact, 0.95, 0.98, "exact_match", false},
		{"invalid status", AlignmentStatus(999), 0.8, 0.7, "test", true},
		{"confidence too low", AlignmentExact, -0.1, 0.8, "test", true},
		{"confidence too high", AlignmentExact, 1.1, 0.8, "test", true},
		{"empty method", AlignmentExact, 0.8, 0.7, "", true},
		{"boundary confidence 0", AlignmentExact, 0.0, 0.5, "test", false},
		{"boundary confidence 1", AlignmentExact, 1.0, 0.5, "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewAlignmentResult(tt.status, tt.confidence, tt.score, tt.method)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("NewAlignmentResult() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewAlignmentResult() unexpected error: %v", err)
				return
			}
			
			if result.Status != tt.status {
				t.Errorf("Status = %v, want %v", result.Status, tt.status)
			}
			if result.Confidence != tt.confidence {
				t.Errorf("Confidence = %f, want %f", result.Confidence, tt.confidence)
			}
			if result.Score != tt.score {
				t.Errorf("Score = %f, want %f", result.Score, tt.score)
			}
			if result.Method != tt.method {
				t.Errorf("Method = %q, want %q", result.Method, tt.method)
			}
		})
	}
}

func TestAlignmentResult_IsGoodAlignment(t *testing.T) {
	tests := []struct {
		name     string
		result   AlignmentResult
		expected bool
	}{
		{"high quality, high confidence", AlignmentResult{AlignmentExact, 0.9, 0.95, "test"}, true},
		{"high quality, low confidence", AlignmentResult{AlignmentExact, 0.5, 0.95, "test"}, false},
		{"low quality, high confidence", AlignmentResult{AlignmentPartial, 0.9, 0.95, "test"}, false},
		{"medium quality, high confidence", AlignmentResult{AlignmentSemantic, 0.8, 0.95, "test"}, true},
		{"boundary case - quality 60", AlignmentResult{AlignmentSemantic, 0.7, 0.95, "test"}, true},
		{"boundary case - confidence 0.7", AlignmentResult{AlignmentExact, 0.7, 0.95, "test"}, true},
		{"below thresholds", AlignmentResult{AlignmentApproximate, 0.6, 0.95, "test"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsGoodAlignment(); got != tt.expected {
				t.Errorf("IsGoodAlignment() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAlignmentResult_String(t *testing.T) {
	result := AlignmentResult{
		Status:     AlignmentFuzzy,
		Confidence: 0.85,
		Score:      0.92,
		Method:     "fuzzy_match",
	}
	
	expected := "AlignmentResult{status=fuzzy, confidence=0.85, score=0.92, method=fuzzy_match}"
	if got := result.String(); got != expected {
		t.Errorf("String() = %q, want %q", got, expected)
	}
}