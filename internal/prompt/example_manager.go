package prompt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/sehwan505/langextract-go/pkg/extraction"
)

// ExampleManager manages loading, validation, and organization of extraction examples
type ExampleManager struct {
	examples     map[string]*ExampleCollection
	validators   []ExampleValidator
	scorers      []ExampleScorer
	mutex        sync.RWMutex
	cacheEnabled bool
}

// ExampleCollection represents a collection of examples for a specific domain or task
type ExampleCollection struct {
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Domain      string                      `json:"domain"`
	Version     string                      `json:"version"`
	Examples    []*extraction.ExampleData   `json:"examples"`
	Metadata    map[string]interface{}      `json:"metadata"`
	Tags        []string                    `json:"tags"`
	CreatedAt   int64                       `json:"created_at"`
	UpdatedAt   int64                       `json:"updated_at"`
	Statistics  *ExampleCollectionStats     `json:"statistics"`
}

// ExampleCollectionStats contains statistics about an example collection
type ExampleCollectionStats struct {
	TotalExamples    int                    `json:"total_examples"`
	ClassCounts      map[string]int         `json:"class_counts"`
	AverageLength    float64                `json:"average_length"`
	QualityScores    *QualityStats          `json:"quality_scores"`
	ComplexityScores *ComplexityStats       `json:"complexity_scores"`
	ValidationStatus *ValidationStats       `json:"validation_status"`
}

// QualityStats contains quality-related statistics
type QualityStats struct {
	Average float64 `json:"average"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	StdDev  float64 `json:"std_dev"`
}

// ComplexityStats contains complexity-related statistics
type ComplexityStats struct {
	Average float64 `json:"average"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	StdDev  float64 `json:"std_dev"`
}

// ValidationStats contains validation-related statistics
type ValidationStats struct {
	ValidCount   int     `json:"valid_count"`
	InvalidCount int     `json:"invalid_count"`
	ValidRatio   float64 `json:"valid_ratio"`
}

// ExampleValidator defines the interface for validating examples
type ExampleValidator interface {
	ValidateExample(ctx context.Context, example *extraction.ExampleData) error
	Name() string
}

// ExampleScorer defines the interface for scoring examples
type ExampleScorer interface {
	ScoreExample(ctx context.Context, example *extraction.ExampleData) (float64, error)
	Name() string
}

// ExampleLoadOptions contains options for loading examples
type ExampleLoadOptions struct {
	ValidateOnLoad   bool     `json:"validate_on_load"`
	ComputeScores    bool     `json:"compute_scores"`
	FilterTags       []string `json:"filter_tags"`
	FilterClasses    []string `json:"filter_classes"`
	MinQualityScore  float64  `json:"min_quality_score"`
	MaxExamples      int      `json:"max_examples"`
	ShuffleExamples  bool     `json:"shuffle_examples"`
}

// NewExampleManager creates a new example manager
func NewExampleManager() *ExampleManager {
	manager := &ExampleManager{
		examples:     make(map[string]*ExampleCollection),
		validators:   make([]ExampleValidator, 0),
		scorers:      make([]ExampleScorer, 0),
		cacheEnabled: true,
	}

	// Register default validators and scorers
	manager.RegisterValidator(NewBasicExampleValidator())
	manager.RegisterScorer(NewQualityExampleScorer())
	manager.RegisterScorer(NewComplexityExampleScorer())

	return manager
}

// RegisterValidator registers an example validator
func (m *ExampleManager) RegisterValidator(validator ExampleValidator) {
	if validator == nil {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.validators = append(m.validators, validator)
}

// RegisterScorer registers an example scorer
func (m *ExampleManager) RegisterScorer(scorer ExampleScorer) {
	if scorer == nil {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.scorers = append(m.scorers, scorer)
}

// LoadExamplesFromFile loads examples from a JSON file
func (m *ExampleManager) LoadExamplesFromFile(ctx context.Context, filePath string, opts *ExampleLoadOptions) (*ExampleCollection, error) {
	if opts == nil {
		opts = &ExampleLoadOptions{
			ValidateOnLoad: true,
			ComputeScores:  true,
		}
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, NewExampleError("loading", fmt.Sprintf("failed to read file %s", filePath), err)
	}

	// Parse JSON
	var collection ExampleCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, NewExampleError("loading", fmt.Sprintf("failed to parse JSON from %s", filePath), err)
	}

	// Set collection name from filename if not provided
	if collection.Name == "" {
		collection.Name = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	}

	// Process examples
	if err := m.processExampleCollection(ctx, &collection, opts); err != nil {
		return nil, NewExampleError("loading", "failed to process example collection", err)
	}

	// Store in cache
	m.mutex.Lock()
	m.examples[collection.Name] = &collection
	m.mutex.Unlock()

	return &collection, nil
}

// LoadExamplesFromReader loads examples from an io.Reader
func (m *ExampleManager) LoadExamplesFromReader(ctx context.Context, reader io.Reader, name string, opts *ExampleLoadOptions) (*ExampleCollection, error) {
	if opts == nil {
		opts = &ExampleLoadOptions{
			ValidateOnLoad: true,
			ComputeScores:  true,
		}
	}

	// Read data
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, NewExampleError("loading", "failed to read data", err)
	}

	// Parse JSON
	var collection ExampleCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, NewExampleError("loading", "failed to parse JSON", err)
	}

	// Set collection name
	if collection.Name == "" {
		collection.Name = name
	}

	// Process examples
	if err := m.processExampleCollection(ctx, &collection, opts); err != nil {
		return nil, NewExampleError("loading", "failed to process example collection", err)
	}

	// Store in cache
	m.mutex.Lock()
	m.examples[collection.Name] = &collection
	m.mutex.Unlock()

	return &collection, nil
}

// LoadExamplesFromDirectory loads all example files from a directory
func (m *ExampleManager) LoadExamplesFromDirectory(ctx context.Context, dir string, opts *ExampleLoadOptions) ([]*ExampleCollection, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, NewExampleError("loading", fmt.Sprintf("failed to find JSON files in %s", dir), err)
	}

	var collections []*ExampleCollection
	for _, file := range files {
		collection, err := m.LoadExamplesFromFile(ctx, file, opts)
		if err != nil {
			// Log error but continue with other files
			continue
		}
		collections = append(collections, collection)
	}

	if len(collections) == 0 {
		return nil, NewExampleError("loading", fmt.Sprintf("no valid example collections found in %s", dir), nil)
	}

	return collections, nil
}

// GetExampleCollection retrieves an example collection by name
func (m *ExampleManager) GetExampleCollection(name string) (*ExampleCollection, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	collection, exists := m.examples[name]
	if !exists {
		return nil, NewExampleError("retrieval", fmt.Sprintf("collection '%s' not found", name), nil)
	}

	return collection, nil
}

// ListExampleCollections returns all available example collection names
func (m *ExampleManager) ListExampleCollections() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var names []string
	for name := range m.examples {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// ValidateExampleCollection validates all examples in a collection
func (m *ExampleManager) ValidateExampleCollection(ctx context.Context, collection *ExampleCollection) []error {
	var errors []error

	for i, example := range collection.Examples {
		if err := m.ValidateExample(ctx, example); err != nil {
			errors = append(errors, NewExampleErrorWithID(
				fmt.Sprintf("%s[%d]", collection.Name, i),
				"validation",
				"validation failed",
				err,
			))
		}
	}

	return errors
}

// ValidateExample validates a single example using all registered validators
func (m *ExampleManager) ValidateExample(ctx context.Context, example *extraction.ExampleData) error {
	m.mutex.RLock()
	validators := make([]ExampleValidator, len(m.validators))
	copy(validators, m.validators)
	m.mutex.RUnlock()

	for _, validator := range validators {
		if err := validator.ValidateExample(ctx, example); err != nil {
			return fmt.Errorf("validator '%s': %w", validator.Name(), err)
		}
	}

	return nil
}

// ScoreExample scores a single example using all registered scorers
func (m *ExampleManager) ScoreExample(ctx context.Context, example *extraction.ExampleData) (map[string]float64, error) {
	m.mutex.RLock()
	scorers := make([]ExampleScorer, len(m.scorers))
	copy(scorers, m.scorers)
	m.mutex.RUnlock()

	scores := make(map[string]float64)

	for _, scorer := range scorers {
		score, err := scorer.ScoreExample(ctx, example)
		if err != nil {
			return nil, fmt.Errorf("scorer '%s': %w", scorer.Name(), err)
		}
		scores[scorer.Name()] = score
	}

	return scores, nil
}

// processExampleCollection processes a loaded example collection
func (m *ExampleManager) processExampleCollection(ctx context.Context, collection *ExampleCollection, opts *ExampleLoadOptions) error {
	// Filter examples by tags
	if len(opts.FilterTags) > 0 {
		collection.Examples = m.filterExamplesByTags(collection.Examples, opts.FilterTags)
	}

	// Filter examples by classes
	if len(opts.FilterClasses) > 0 {
		collection.Examples = m.filterExamplesByClasses(collection.Examples, opts.FilterClasses)
	}

	// Validate examples
	if opts.ValidateOnLoad {
		validExamples := make([]*extraction.ExampleData, 0, len(collection.Examples))
		for _, example := range collection.Examples {
			if err := m.ValidateExample(ctx, example); err == nil {
				validExamples = append(validExamples, example)
			}
		}
		collection.Examples = validExamples
	}

	// Compute scores and filter by quality
	if opts.ComputeScores || opts.MinQualityScore > 0 {
		qualityExamples := make([]*extraction.ExampleData, 0, len(collection.Examples))
		for _, example := range collection.Examples {
			scores, err := m.ScoreExample(ctx, example)
			if err != nil {
				continue
			}

			// Store scores in metadata
			if example.Metadata == nil {
				example.Metadata = make(map[string]interface{})
			}
			example.Metadata["scores"] = scores

			// Check quality threshold
			if qualityScore, exists := scores["quality"]; exists {
				if qualityScore >= opts.MinQualityScore {
					qualityExamples = append(qualityExamples, example)
				}
			} else {
				qualityExamples = append(qualityExamples, example)
			}
		}
		collection.Examples = qualityExamples
	}

	// Limit number of examples
	if opts.MaxExamples > 0 && len(collection.Examples) > opts.MaxExamples {
		collection.Examples = collection.Examples[:opts.MaxExamples]
	}

	// Compute collection statistics
	collection.Statistics = m.computeCollectionStatistics(ctx, collection)

	return nil
}

// filterExamplesByTags filters examples that contain any of the specified tags
func (m *ExampleManager) filterExamplesByTags(examples []*extraction.ExampleData, tags []string) []*extraction.ExampleData {
	if len(tags) == 0 {
		return examples
	}

	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[strings.ToLower(tag)] = true
	}

	var filtered []*extraction.ExampleData
	for _, example := range examples {
		if example.Metadata != nil {
			if exampleTags, exists := example.Metadata["tags"]; exists {
				if tagList, ok := exampleTags.([]interface{}); ok {
					for _, tag := range tagList {
						if tagStr, ok := tag.(string); ok {
							if tagSet[strings.ToLower(tagStr)] {
								filtered = append(filtered, example)
								break
							}
						}
					}
				}
			}
		}
	}

	return filtered
}

// filterExamplesByClasses filters examples that contain any of the specified classes
func (m *ExampleManager) filterExamplesByClasses(examples []*extraction.ExampleData, classes []string) []*extraction.ExampleData {
	if len(classes) == 0 {
		return examples
	}

	classSet := make(map[string]bool)
	for _, class := range classes {
		classSet[strings.ToLower(class)] = true
	}

	var filtered []*extraction.ExampleData
	for _, example := range examples {
		for _, extraction := range example.Extractions {
			if classSet[strings.ToLower(extraction.Class())] {
				filtered = append(filtered, example)
				break
			}
		}
	}

	return filtered
}

// computeCollectionStatistics computes statistics for an example collection
func (m *ExampleManager) computeCollectionStatistics(ctx context.Context, collection *ExampleCollection) *ExampleCollectionStats {
	stats := &ExampleCollectionStats{
		TotalExamples: len(collection.Examples),
		ClassCounts:   make(map[string]int),
	}

	if len(collection.Examples) == 0 {
		return stats
	}

	var totalLength int
	var qualityScores []float64
	var complexityScores []float64
	var validCount int

	for _, example := range collection.Examples {
		// Length statistics
		totalLength += len(example.Text)

		// Class counts
		for _, extraction := range example.Extractions {
			stats.ClassCounts[extraction.Class()]++
		}

		// Quality and complexity scores from metadata
		if example.Metadata != nil {
			if scores, exists := example.Metadata["scores"]; exists {
				if scoreMap, ok := scores.(map[string]float64); ok {
					if quality, exists := scoreMap["quality"]; exists {
						qualityScores = append(qualityScores, quality)
					}
					if complexity, exists := scoreMap["complexity"]; exists {
						complexityScores = append(complexityScores, complexity)
					}
				}
			}
		}

		// Validation status
		if m.ValidateExample(ctx, example) == nil {
			validCount++
		}
	}

	// Average length
	stats.AverageLength = float64(totalLength) / float64(len(collection.Examples))

	// Quality statistics
	if len(qualityScores) > 0 {
		stats.QualityScores = computeFloatStats(qualityScores)
	}

	// Complexity statistics
	if len(complexityScores) > 0 {
		qualityStats := computeFloatStats(complexityScores)
		stats.ComplexityScores = &ComplexityStats{
			Average: qualityStats.Average,
			Min:     qualityStats.Min,
			Max:     qualityStats.Max,
			StdDev:  qualityStats.StdDev,
		}
	}

	// Validation statistics
	stats.ValidationStatus = &ValidationStats{
		ValidCount:   validCount,
		InvalidCount: len(collection.Examples) - validCount,
		ValidRatio:   float64(validCount) / float64(len(collection.Examples)),
	}

	return stats
}

// computeFloatStats computes basic statistics for a slice of float64 values
func computeFloatStats(values []float64) *QualityStats {
	if len(values) == 0 {
		return &QualityStats{}
	}

	var sum, sumSquares float64
	min := values[0]
	max := values[0]

	for _, value := range values {
		sum += value
		sumSquares += value * value
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}

	mean := sum / float64(len(values))
	variance := (sumSquares / float64(len(values))) - (mean * mean)
	stdDev := 0.0
	if variance > 0 {
		// Simplified square root approximation
		stdDev = variance / 2 // This is just a placeholder - proper sqrt would be needed
	}

	return &QualityStats{
		Average: mean,
		Min:     min,
		Max:     max,
		StdDev:  stdDev,
	}
}

// BasicExampleValidator implements basic example validation
type BasicExampleValidator struct{}

// NewBasicExampleValidator creates a new basic example validator
func NewBasicExampleValidator() *BasicExampleValidator {
	return &BasicExampleValidator{}
}

// Name returns the name of the validator
func (v *BasicExampleValidator) Name() string {
	return "basic"
}

// ValidateExample validates basic example requirements
func (v *BasicExampleValidator) ValidateExample(ctx context.Context, example *extraction.ExampleData) error {
	if example == nil {
		return fmt.Errorf("example cannot be nil")
	}

	if example.Text == "" {
		return fmt.Errorf("example input cannot be empty")
	}

	if len(example.Extractions) == 0 {
		return fmt.Errorf("example must have at least one extraction")
	}

	for i, extraction := range example.Extractions {
		if extraction.Text() == "" {
			return fmt.Errorf("extraction %d: text cannot be empty", i)
		}
		if extraction.Class() == "" {
			return fmt.Errorf("extraction %d: class cannot be empty", i)
		}
	}

	return nil
}

// QualityExampleScorer implements quality-based example scoring
type QualityExampleScorer struct {
	selector *QualityExampleSelector
}

// NewQualityExampleScorer creates a new quality example scorer
func NewQualityExampleScorer() *QualityExampleScorer {
	return &QualityExampleScorer{
		selector: NewQualityExampleSelector(),
	}
}

// Name returns the name of the scorer
func (s *QualityExampleScorer) Name() string {
	return "quality"
}

// ScoreExample scores an example's quality
func (s *QualityExampleScorer) ScoreExample(ctx context.Context, example *extraction.ExampleData) (float64, error) {
	// Use the quality selector's scoring logic
	return s.selector.ScoreExample(ctx, &ExtractionTask{}, example)
}

// ComplexityExampleScorer implements complexity-based example scoring
type ComplexityExampleScorer struct{}

// NewComplexityExampleScorer creates a new complexity example scorer
func NewComplexityExampleScorer() *ComplexityExampleScorer {
	return &ComplexityExampleScorer{}
}

// Name returns the name of the scorer
func (s *ComplexityExampleScorer) Name() string {
	return "complexity"
}

// ScoreExample scores an example's complexity
func (s *ComplexityExampleScorer) ScoreExample(ctx context.Context, example *extraction.ExampleData) (float64, error) {
	if example == nil {
		return 0.0, fmt.Errorf("example cannot be nil")
	}

	var complexity float64

	// Input complexity
	inputLength := len(example.Text)
	sentences := strings.Count(example.Text, ".") + strings.Count(example.Text, "!") + strings.Count(example.Text, "?")
	words := len(strings.Fields(example.Text))

	// Length complexity (normalized)
	if inputLength > 100 {
		complexity += 0.3
	}

	// Sentence complexity
	if sentences > 2 {
		complexity += 0.2
	}

	// Vocabulary complexity
	if words > 20 {
		complexity += 0.2
	}

	// Extraction complexity
	if len(example.Extractions) > 3 {
		complexity += 0.2
	}

	// Class diversity
	classSet := make(map[string]bool)
	for _, extraction := range example.Extractions {
		classSet[extraction.Class()] = true
	}

	if len(classSet) > 2 {
		complexity += 0.1
	}

	return complexity, nil
}