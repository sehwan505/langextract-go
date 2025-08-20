package types_test

import (
	"testing"

	"github.com/sehwan505/langextract-go/pkg/types"
)

func TestAlignmentStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   types.AlignmentStatus
		expected string
	}{
		{"none", types.AlignmentNone, "none"},
		{"exact", types.AlignmentExact, "exact"},
		{"fuzzy", types.AlignmentFuzzy, "fuzzy"},
		{"semantic", types.AlignmentSemantic, "semantic"},
		{"partial", types.AlignmentPartial, "partial"},
		{"approximate", types.AlignmentApproximate, "approximate"},
		{"unknown", types.AlignmentStatus(999), "unknown(999)"},
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
		status   types.AlignmentStatus
		expected bool
	}{
		{"none", types.AlignmentNone, true},
		{"exact", types.AlignmentExact, true},
		{"fuzzy", types.AlignmentFuzzy, true},
		{"semantic", types.AlignmentSemantic, true},
		{"partial", types.AlignmentPartial, true},
		{"approximate", types.AlignmentApproximate, true},
		{"invalid negative", types.AlignmentStatus(-1), false},
		{"invalid large", types.AlignmentStatus(999), false},
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
		status   types.AlignmentStatus
		expected int
	}{
		{"none", types.AlignmentNone, 0},
		{"exact", types.AlignmentExact, 100},
		{"fuzzy", types.AlignmentFuzzy, 80},
		{"semantic", types.AlignmentSemantic, 60},
		{"partial", types.AlignmentPartial, 40},
		{"approximate", types.AlignmentApproximate, 20},
		{"unknown", types.AlignmentStatus(999), 0},
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
		expected  types.AlignmentStatus
		wantError bool
	}{
		{"none", "none", types.AlignmentNone, false},
		{"exact", "exact", types.AlignmentExact, false},
		{"fuzzy", "fuzzy", types.AlignmentFuzzy, false},
		{"semantic", "semantic", types.AlignmentSemantic, false},
		{"partial", "partial", types.AlignmentPartial, false},
		{"approximate", "approximate", types.AlignmentApproximate, false},
		{"uppercase", "EXACT", types.AlignmentExact, false},
		{"mixed case", "Fuzzy", types.AlignmentFuzzy, false},
		{"invalid", "invalid", types.AlignmentNone, true},
		{"empty", "", types.AlignmentNone, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.ParseAlignmentStatus(tt.input)
			
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
		status     types.AlignmentStatus
		confidence float64
		score      float64
		method     string
		wantError  bool
	}{
		{"valid result", types.AlignmentExact, 0.95, 0.98, "exact_match", false},
		{"invalid status", types.AlignmentStatus(999), 0.8, 0.7, "test", true},
		{"confidence too low", types.AlignmentExact, -0.1, 0.8, "test", true},
		{"confidence too high", types.AlignmentExact, 1.1, 0.8, "test", true},
		{"empty method", types.AlignmentExact, 0.8, 0.7, "", true},
		{"boundary confidence 0", types.AlignmentExact, 0.0, 0.5, "test", false},
		{"boundary confidence 1", types.AlignmentExact, 1.0, 0.5, "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := types.NewAlignmentResult(tt.status, tt.confidence, tt.score, tt.method)
			
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
		result   types.AlignmentResult
		expected bool
	}{
		{"high quality, high confidence", types.AlignmentResult{Status: types.AlignmentExact, Confidence: 0.9, Score: 0.95, Method: "test"}, true},
		{"high quality, low confidence", types.AlignmentResult{Status: types.AlignmentExact, Confidence: 0.5, Score: 0.95, Method: "test"}, false},
		{"low quality, high confidence", types.AlignmentResult{Status: types.AlignmentPartial, Confidence: 0.9, Score: 0.95, Method: "test"}, false},
		{"medium quality, high confidence", types.AlignmentResult{Status: types.AlignmentSemantic, Confidence: 0.8, Score: 0.95, Method: "test"}, true},
		{"boundary case - quality 60", types.AlignmentResult{Status: types.AlignmentSemantic, Confidence: 0.7, Score: 0.95, Method: "test"}, true},
		{"boundary case - confidence 0.7", types.AlignmentResult{Status: types.AlignmentExact, Confidence: 0.7, Score: 0.95, Method: "test"}, true},
		{"below thresholds", types.AlignmentResult{Status: types.AlignmentApproximate, Confidence: 0.6, Score: 0.95, Method: "test"}, false},
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
	result := types.AlignmentResult{
		Status:     types.AlignmentFuzzy,
		Confidence: 0.85,
		Score:      0.92,
		Method:     "fuzzy_match",
	}
	
	expected := "AlignmentResult{status=fuzzy, confidence=0.85, score=0.92, method=fuzzy_match}"
	if got := result.String(); got != expected {
		t.Errorf("String() = %q, want %q", got, expected)
	}
}