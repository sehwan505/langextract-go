package visualization

// DefaultVisualizationCSS contains the CSS styles for HTML visualizations
// Based on the Python langextract reference implementation
const DefaultVisualizationCSS = `<style>
.lx-highlight { 
    position: relative; 
    border-radius: 3px; 
    padding: 1px 2px;
}

.lx-highlight .lx-tooltip {
    visibility: hidden;
    opacity: 0;
    transition: opacity 0.2s ease-in-out;
    background: #333;
    color: #fff;
    text-align: left;
    border-radius: 4px;
    padding: 6px 8px;
    position: absolute;
    z-index: 1000;
    bottom: 125%;
    left: 50%;
    transform: translateX(-50%);
    font-size: 12px;
    max-width: 240px;
    white-space: normal;
    box-shadow: 0 2px 6px rgba(0,0,0,0.3);
}

.lx-highlight:hover .lx-tooltip { 
    visibility: visible; 
    opacity: 1; 
}

.lx-animated-wrapper { 
    max-width: 100%; 
    font-family: Arial, sans-serif; 
}

.lx-controls {
    background: #fafafa; 
    border: 1px solid #90caf9; 
    border-radius: 8px;
    padding: 12px; 
    margin-bottom: 16px;
}

.lx-button-row {
    display: flex; 
    justify-content: center; 
    gap: 8px; 
    margin-bottom: 12px;
}

.lx-control-btn {
    background: #4285f4; 
    color: white; 
    border: none; 
    border-radius: 4px;
    padding: 8px 16px; 
    cursor: pointer; 
    font-size: 13px; 
    font-weight: 500;
    transition: background-color 0.2s;
}

.lx-control-btn:hover { 
    background: #3367d6; 
}

.lx-control-btn:disabled {
    background: #cccccc;
    cursor: not-allowed;
}

.lx-progress-container {
    margin-bottom: 8px;
}

.lx-progress-slider {
    width: 100%; 
    margin: 0; 
    appearance: none; 
    height: 6px;
    background: #ddd; 
    border-radius: 3px; 
    outline: none;
}

.lx-progress-slider::-webkit-slider-thumb {
    appearance: none; 
    width: 18px; 
    height: 18px; 
    background: #4285f4;
    border-radius: 50%; 
    cursor: pointer;
}

.lx-progress-slider::-moz-range-thumb {
    width: 18px; 
    height: 18px; 
    background: #4285f4; 
    border-radius: 50%;
    cursor: pointer; 
    border: none;
}

.lx-status-text {
    text-align: center; 
    font-size: 12px; 
    color: #666; 
    margin-top: 4px;
}

.lx-text-window {
    font-family: monospace; 
    white-space: pre-wrap; 
    border: 1px solid #90caf9;
    padding: 12px; 
    max-height: 260px; 
    overflow-y: auto; 
    margin-bottom: 12px;
    line-height: 1.6;
}

.lx-attributes-panel {
    background: #fafafa; 
    border: 1px solid #90caf9; 
    border-radius: 6px;
    padding: 8px 10px; 
    margin-top: 8px; 
    font-size: 13px;
}

.lx-current-highlight {
    border-bottom: 4px solid #ff4444;
    font-weight: bold;
    animation: lx-pulse 1s ease-in-out;
}

@keyframes lx-pulse {
    0% { text-decoration-color: #ff4444; }
    50% { text-decoration-color: #ff0000; }
    100% { text-decoration-color: #ff4444; }
}

.lx-legend {
    font-size: 12px; 
    margin-bottom: 8px;
    padding-bottom: 8px; 
    border-bottom: 1px solid #e0e0e0;
}

.lx-label {
    display: inline-block;
    padding: 2px 4px;
    border-radius: 3px;
    margin-right: 4px;
    color: #000;
}

.lx-attr-key {
    font-weight: 600;
    color: #1565c0;
    letter-spacing: 0.3px;
}

.lx-attr-value {
    font-weight: 400;
    opacity: 0.85;
    letter-spacing: 0.2px;
}

/* GIF optimizations with larger fonts and better readability */
.lx-gif-optimized .lx-text-window { 
    font-size: 16px; 
    line-height: 1.8; 
}

.lx-gif-optimized .lx-attributes-panel { 
    font-size: 15px; 
}

.lx-gif-optimized .lx-current-highlight { 
    text-decoration-thickness: 4px; 
}

/* Responsive design adjustments */
@media (max-width: 768px) {
    .lx-animated-wrapper {
        font-size: 14px;
    }
    
    .lx-button-row {
        flex-direction: column;
        gap: 4px;
    }
    
    .lx-control-btn {
        font-size: 12px;
        padding: 6px 12px;
    }
    
    .lx-text-window {
        max-height: 200px;
        font-size: 12px;
    }
}

/* Print styles */
@media print {
    .lx-controls {
        display: none;
    }
    
    .lx-text-window {
        max-height: none;
        border: none;
        overflow: visible;
    }
    
    .lx-highlight {
        border: 1px solid #333;
    }
}

/* Dark mode support */
@media (prefers-color-scheme: dark) {
    .lx-animated-wrapper {
        color: #ffffff;
    }
    
    .lx-controls, .lx-attributes-panel {
        background: #2d2d2d;
        border-color: #555555;
        color: #ffffff;
    }
    
    .lx-text-window {
        background: #1e1e1e;
        border-color: #555555;
        color: #ffffff;
    }
    
    .lx-status-text {
        color: #cccccc;
    }
    
    .lx-legend {
        border-bottom-color: #555555;
    }
    
    .lx-progress-slider {
        background: #555555;
    }
}

/* Accessibility improvements */
.lx-highlight:focus {
    outline: 2px solid #4285f4;
    outline-offset: 2px;
}

.lx-control-btn:focus {
    outline: 2px solid #ffffff;
    outline-offset: 2px;
}

/* Animation controls for reduced motion preference */
@media (prefers-reduced-motion: reduce) {
    .lx-current-highlight {
        animation: none;
    }
    
    .lx-highlight .lx-tooltip {
        transition: none;
    }
}

/* High contrast mode support */
@media (prefers-contrast: high) {
    .lx-highlight {
        border: 2px solid #000000;
    }
    
    .lx-control-btn {
        border: 2px solid #000000;
    }
    
    .lx-text-window, .lx-controls, .lx-attributes-panel {
        border: 2px solid #000000;
    }
}
</style>`

// MinimalVisualizationCSS provides a minimal CSS for lightweight visualizations
const MinimalVisualizationCSS = `<style>
.lx-highlight { 
    position: relative; 
    border-radius: 2px; 
    padding: 1px;
}

.lx-animated-wrapper { 
    font-family: Arial, sans-serif; 
}

.lx-text-window {
    font-family: monospace; 
    white-space: pre-wrap; 
    border: 1px solid #ccc;
    padding: 8px; 
    margin-bottom: 8px;
}

.lx-legend {
    font-size: 12px; 
    margin-bottom: 4px;
    padding-bottom: 4px; 
    border-bottom: 1px solid #e0e0e0;
}

.lx-label {
    display: inline-block;
    padding: 1px 3px;
    border-radius: 2px;
    margin-right: 3px;
    color: #000;
}
</style>`

// PrintVisualizationCSS provides CSS optimized for printing
const PrintVisualizationCSS = `<style>
@media print {
    .lx-animated-wrapper {
        font-family: serif;
    }
    
    .lx-highlight {
        border: 1px solid #000;
        padding: 0;
    }
    
    .lx-text-window {
        border: 1px solid #000;
        max-height: none;
        overflow: visible;
    }
    
    .lx-legend {
        page-break-inside: avoid;
    }
    
    .lx-label {
        border: 1px solid #000;
        background: none !important;
    }
}
</style>`

// GetVisualizationCSS returns the appropriate CSS based on options
func GetVisualizationCSS(opts *VisualizationOptions) string {
	if opts == nil {
		return DefaultVisualizationCSS
	}
	
	// Check for template overrides
	if cssOverride, exists := opts.TemplateOverrides["css"]; exists {
		return "<style>" + cssOverride + "</style>"
	}
	
	// Return appropriate CSS based on options
	if opts.Debug {
		return DefaultVisualizationCSS + getDebugCSS()
	}
	
	return DefaultVisualizationCSS
}

// getDebugCSS returns additional CSS for debug mode
func getDebugCSS() string {
	return `
<style>
/* Debug mode styles */
.lx-debug {
    border: 1px dashed #ff0000;
    background: rgba(255, 0, 0, 0.1);
}

.lx-debug-info {
    position: absolute;
    top: -20px;
    left: 0;
    background: #ff0000;
    color: #ffffff;
    font-size: 10px;
    padding: 2px 4px;
    border-radius: 2px;
    z-index: 1001;
}

.lx-span-boundaries {
    position: relative;
}

.lx-span-boundaries::before {
    content: '[';
    color: #ff0000;
    font-weight: bold;
}

.lx-span-boundaries::after {
    content: ']';
    color: #ff0000;
    font-weight: bold;
}
</style>`
}

// HTMLTemplateVariables contains placeholders used in HTML templates
type HTMLTemplateVariables struct {
	// CSS contains the CSS styles
	CSS string
	
	// Title is the document title
	Title string
	
	// HighlightedText is the text with HTML highlighting
	HighlightedText string
	
	// LegendHTML is the legend showing class colors
	LegendHTML string
	
	// ExtractionData is the JavaScript data for extractions
	ExtractionData string
	
	// AnimationSpeed is the animation speed in seconds
	AnimationSpeed float64
	
	// ExtractionCount is the total number of extractions
	ExtractionCount int
	
	// FirstExtractionPos is the position info for the first extraction
	FirstExtractionPos string
	
	// GIFOptimized indicates if GIF optimizations should be applied
	GIFOptimized bool
	
	// ShowControls indicates if animation controls should be shown
	ShowControls bool
	
	// CustomJavaScript contains any custom JavaScript code
	CustomJavaScript string
}