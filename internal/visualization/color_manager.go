package visualization

import (
	"sort"
	"sync"

	"github.com/sehwan505/langextract-go/pkg/extraction"
)

// DefaultColorPalette is the default color palette based on Google's Material Design colors
// Matches the Python reference implementation palette
var DefaultColorPalette = []string{
	"#D2E3FC", // Light Blue (Primary Container)
	"#C8E6C9", // Light Green (Tertiary Container)
	"#FEF0C3", // Light Yellow (Primary Color)
	"#F9DEDC", // Light Red (Error Container)
	"#FFDDBE", // Light Orange (Tertiary Container)
	"#EADDFF", // Light Purple (Secondary/Tertiary Container)
	"#C4E9E4", // Light Teal (Teal Container)
	"#FCE4EC", // Light Pink (Pink Container)
	"#E8EAED", // Very Light Grey (Neutral Highlight)
	"#DDE8E8", // Pale Cyan (Cyan Container)
}

// DefaultFallbackColor is used when no other color is available
const DefaultFallbackColor = "#FFFF8D"

// DefaultColorManager implements the ColorManager interface
type DefaultColorManager struct {
	// palette is the color palette to use for assignments
	palette []string
	
	// assignments maps extraction classes to assigned colors
	assignments map[string]string
	
	// mutex protects concurrent access to assignments
	mutex sync.RWMutex
	
	// cycleIndex tracks the current position in the palette cycle
	cycleIndex int
}

// NewDefaultColorManager creates a new default color manager
func NewDefaultColorManager() *DefaultColorManager {
	return &DefaultColorManager{
		palette:     make([]string, len(DefaultColorPalette)),
		assignments: make(map[string]string),
		cycleIndex:  0,
	}
}

// NewColorManagerWithPalette creates a color manager with a custom palette
func NewColorManagerWithPalette(palette []string) *DefaultColorManager {
	if len(palette) == 0 {
		palette = DefaultColorPalette
	}
	
	cm := &DefaultColorManager{
		palette:     make([]string, len(palette)),
		assignments: make(map[string]string),
		cycleIndex:  0,
	}
	
	// Copy palette to avoid external modifications
	copy(cm.palette, palette)
	
	return cm
}

// AssignColors assigns colors to extraction classes based on the extractions provided
func (cm *DefaultColorManager) AssignColors(extractions []*extraction.Extraction) map[string]string {
	if len(extractions) == 0 {
		return make(map[string]string)
	}
	
	// Extract unique classes from valid extractions
	classSet := make(map[string]bool)
	for _, ext := range extractions {
		if ext != nil && ext.Class() != "" && ext.Interval() != nil {
			classSet[ext.Class()] = true
		}
	}
	
	// Convert to sorted slice for consistent color assignment
	var classes []string
	for class := range classSet {
		classes = append(classes, class)
	}
	sort.Strings(classes)
	
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	// Assign colors to new classes
	for _, class := range classes {
		if _, exists := cm.assignments[class]; !exists {
			color := cm.getNextColor()
			cm.assignments[class] = color
		}
	}
	
	// Return a copy of current assignments
	result := make(map[string]string)
	for class, color := range cm.assignments {
		result[class] = color
	}
	
	return result
}

// GetColor returns the color for a specific class
func (cm *DefaultColorManager) GetColor(class string) string {
	if class == "" {
		return DefaultFallbackColor
	}
	
	cm.mutex.RLock()
	color, exists := cm.assignments[class]
	cm.mutex.RUnlock()
	
	if exists {
		return color
	}
	
	// Assign a new color if not found
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	// Double-check after acquiring write lock
	if color, exists := cm.assignments[class]; exists {
		return color
	}
	
	color = cm.getNextColor()
	cm.assignments[class] = color
	
	return color
}

// GetPalette returns a copy of the current color palette
func (cm *DefaultColorManager) GetPalette() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	result := make([]string, len(cm.palette))
	copy(result, cm.palette)
	return result
}

// SetPalette sets a new color palette and resets assignments
func (cm *DefaultColorManager) SetPalette(palette []string) {
	if len(palette) == 0 {
		return // Don't set empty palette
	}
	
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	cm.palette = make([]string, len(palette))
	copy(cm.palette, palette)
	cm.cycleIndex = 0
	
	// Clear existing assignments as they may reference old palette
	cm.assignments = make(map[string]string)
}

// GetAssignments returns a copy of current class-to-color assignments
func (cm *DefaultColorManager) GetAssignments() map[string]string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	result := make(map[string]string)
	for class, color := range cm.assignments {
		result[class] = color
	}
	return result
}

// ClearAssignments clears all color assignments
func (cm *DefaultColorManager) ClearAssignments() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	cm.assignments = make(map[string]string)
	cm.cycleIndex = 0
}

// SetColorForClass manually sets a color for a specific class
func (cm *DefaultColorManager) SetColorForClass(class, color string) error {
	if class == "" {
		return NewColorError("class name cannot be empty", nil)
	}
	
	if color == "" {
		return NewColorError("color cannot be empty", map[string]interface{}{
			"class": class,
		})
	}
	
	// Basic color validation (hex color format)
	if !isValidHexColor(color) {
		return NewColorError("invalid hex color format", map[string]interface{}{
			"class": class,
			"color": color,
		})
	}
	
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	cm.assignments[class] = color
	return nil
}

// getNextColor returns the next color in the palette (internal method, assumes lock held)
func (cm *DefaultColorManager) getNextColor() string {
	if len(cm.palette) == 0 {
		return DefaultFallbackColor
	}
	
	color := cm.palette[cm.cycleIndex%len(cm.palette)]
	cm.cycleIndex++
	return color
}

// isValidHexColor performs basic validation for hex color format
func isValidHexColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}
	
	for i := 1; i < len(color); i++ {
		c := color[i]
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	
	return true
}

// StaticColorManager is a simple implementation that uses a fixed color mapping
type StaticColorManager struct {
	assignments map[string]string
	fallback    string
}

// NewStaticColorManager creates a static color manager with predefined assignments
func NewStaticColorManager(assignments map[string]string) *StaticColorManager {
	if assignments == nil {
		assignments = make(map[string]string)
	}
	
	return &StaticColorManager{
		assignments: assignments,
		fallback:    DefaultFallbackColor,
	}
}

// AssignColors returns the predefined color assignments
func (cm *StaticColorManager) AssignColors(extractions []*extraction.Extraction) map[string]string {
	// Extract unique classes and return assignments for them
	classSet := make(map[string]bool)
	for _, ext := range extractions {
		if ext != nil && ext.Class() != "" && ext.Interval() != nil {
			classSet[ext.Class()] = true
		}
	}
	
	result := make(map[string]string)
	for class := range classSet {
		if color, exists := cm.assignments[class]; exists {
			result[class] = color
		} else {
			result[class] = cm.fallback
		}
	}
	
	return result
}

// GetColor returns the color for a specific class
func (cm *StaticColorManager) GetColor(class string) string {
	if color, exists := cm.assignments[class]; exists {
		return color
	}
	return cm.fallback
}

// GetPalette returns the colors from the static assignments
func (cm *StaticColorManager) GetPalette() []string {
	var colors []string
	seen := make(map[string]bool)
	
	for _, color := range cm.assignments {
		if !seen[color] {
			colors = append(colors, color)
			seen[color] = true
		}
	}
	
	sort.Strings(colors)
	return colors
}

// SetPalette is not supported by StaticColorManager
func (cm *StaticColorManager) SetPalette(palette []string) {
	// No-op for static color manager
}

// SetFallbackColor sets the fallback color for unknown classes
func (cm *StaticColorManager) SetFallbackColor(color string) {
	if isValidHexColor(color) {
		cm.fallback = color
	}
}

// GetFallbackColor returns the current fallback color
func (cm *StaticColorManager) GetFallbackColor() string {
	return cm.fallback
}