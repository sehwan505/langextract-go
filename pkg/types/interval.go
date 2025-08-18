package types

import (
	"errors"
	"fmt"
)

// CharInterval represents a character position range in text.
// StartPos is inclusive, EndPos is exclusive, following Go's slice conventions.
type CharInterval struct {
	StartPos int // Inclusive start position
	EndPos   int // Exclusive end position
}

// NewCharInterval creates a new CharInterval with validation.
func NewCharInterval(start, end int) (*CharInterval, error) {
	if start < 0 {
		return nil, errors.New("start position cannot be negative")
	}
	if end < start {
		return nil, errors.New("end position cannot be less than start position")
	}
	return &CharInterval{StartPos: start, EndPos: end}, nil
}

// Length returns the number of characters in the interval.
func (ci CharInterval) Length() int {
	return ci.EndPos - ci.StartPos
}

// IsEmpty returns true if the interval contains no characters.
func (ci CharInterval) IsEmpty() bool {
	return ci.StartPos == ci.EndPos
}

// Contains checks if the given position is within this interval.
func (ci CharInterval) Contains(pos int) bool {
	return pos >= ci.StartPos && pos < ci.EndPos
}

// Overlaps checks if this interval overlaps with another interval.
func (ci CharInterval) Overlaps(other CharInterval) bool {
	return ci.StartPos < other.EndPos && other.StartPos < ci.EndPos
}

// Union returns the minimal interval that contains both intervals.
func (ci CharInterval) Union(other CharInterval) CharInterval {
	start := ci.StartPos
	if other.StartPos < start {
		start = other.StartPos
	}
	end := ci.EndPos
	if other.EndPos > end {
		end = other.EndPos
	}
	return CharInterval{StartPos: start, EndPos: end}
}

// Intersection returns the overlapping portion of two intervals.
// Returns nil if there is no overlap.
func (ci CharInterval) Intersection(other CharInterval) *CharInterval {
	if !ci.Overlaps(other) {
		return nil
	}
	start := ci.StartPos
	if other.StartPos > start {
		start = other.StartPos
	}
	end := ci.EndPos
	if other.EndPos < end {
		end = other.EndPos
	}
	return &CharInterval{StartPos: start, EndPos: end}
}

// String returns a string representation of the interval.
func (ci CharInterval) String() string {
	return fmt.Sprintf("[%d:%d)", ci.StartPos, ci.EndPos)
}

// TokenInterval represents a token position range in tokenized text.
// Similar to CharInterval but operates on token indices rather than characters.
type TokenInterval struct {
	StartToken int // Inclusive start token index
	EndToken   int // Exclusive end token index
}

// NewTokenInterval creates a new TokenInterval with validation.
func NewTokenInterval(start, end int) (*TokenInterval, error) {
	if start < 0 {
		return nil, errors.New("start token index cannot be negative")
	}
	if end < start {
		return nil, errors.New("end token index cannot be less than start token index")
	}
	return &TokenInterval{StartToken: start, EndToken: end}, nil
}

// Length returns the number of tokens in the interval.
func (ti TokenInterval) Length() int {
	return ti.EndToken - ti.StartToken
}

// IsEmpty returns true if the interval contains no tokens.
func (ti TokenInterval) IsEmpty() bool {
	return ti.StartToken == ti.EndToken
}

// Contains checks if the given token index is within this interval.
func (ti TokenInterval) Contains(tokenIndex int) bool {
	return tokenIndex >= ti.StartToken && tokenIndex < ti.EndToken
}

// Overlaps checks if this interval overlaps with another token interval.
func (ti TokenInterval) Overlaps(other TokenInterval) bool {
	return ti.StartToken < other.EndToken && other.StartToken < ti.EndToken
}

// Union returns the minimal interval that contains both token intervals.
func (ti TokenInterval) Union(other TokenInterval) TokenInterval {
	start := ti.StartToken
	if other.StartToken < start {
		start = other.StartToken
	}
	end := ti.EndToken
	if other.EndToken > end {
		end = other.EndToken
	}
	return TokenInterval{StartToken: start, EndToken: end}
}

// Intersection returns the overlapping portion of two token intervals.
// Returns nil if there is no overlap.
func (ti TokenInterval) Intersection(other TokenInterval) *TokenInterval {
	if !ti.Overlaps(other) {
		return nil
	}
	start := ti.StartToken
	if other.StartToken > start {
		start = other.StartToken
	}
	end := ti.EndToken
	if other.EndToken < end {
		end = other.EndToken
	}
	return &TokenInterval{StartToken: start, EndToken: end}
}

// String returns a string representation of the token interval.
func (ti TokenInterval) String() string {
	return fmt.Sprintf("tokens[%d:%d)", ti.StartToken, ti.EndToken)
}
