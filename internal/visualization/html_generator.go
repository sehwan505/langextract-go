package visualization

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/extraction"
)

// HTMLGenerator generates interactive HTML visualizations
type HTMLGenerator struct {
	colorManager     ColorManager
	templateRenderer TemplateRenderer
	options          *VisualizationOptions
}

// NewHTMLGenerator creates a new HTML generator
func NewHTMLGenerator(opts *VisualizationOptions) *HTMLGenerator {
	if opts == nil {
		opts = DefaultVisualizationOptions()
	}
	
	return &HTMLGenerator{
		colorManager:     NewDefaultColorManager(),
		templateRenderer: NewDefaultTemplateRenderer(),
		options:         opts,
	}
}

// NewHTMLGeneratorWithColorManager creates an HTML generator with a custom color manager
func NewHTMLGeneratorWithColorManager(colorManager ColorManager, opts *VisualizationOptions) *HTMLGenerator {
	if opts == nil {
		opts = DefaultVisualizationOptions()
	}
	
	if colorManager == nil {
		colorManager = NewDefaultColorManager()
	}
	
	return &HTMLGenerator{
		colorManager:     colorManager,
		templateRenderer: NewDefaultTemplateRenderer(),
		options:         opts,
	}
}

// Generate creates an HTML visualization from an annotated document
func (g *HTMLGenerator) Generate(ctx context.Context, doc *document.AnnotatedDocument, opts *VisualizationOptions) (string, error) {
	if doc == nil {
		return "", NewValidationError("document cannot be nil", nil)
	}
	
	if doc.Text() == "" {
		return "", NewValidationError("document text cannot be empty", nil)
	}
	
	// Use provided options or fall back to generator options
	if opts == nil {
		opts = g.options
	}
	
	if err := opts.Validate(); err != nil {
		return "", err
	}
	
	// Filter and validate extractions
	validExtractions := g.filterValidExtractions(doc.Extractions(), opts.FilterClasses)
	if len(validExtractions) == 0 {
		return g.generateEmptyVisualization(opts)
	}
	
	// Assign colors to extraction classes
	colorMap := g.colorManager.AssignColors(validExtractions)
	if opts.CustomColors != nil {
		// Override with custom colors
		for class, color := range opts.CustomColors {
			colorMap[class] = color
		}
	}
	
	// Sort extractions by position if requested
	if opts.SortExtractions {
		validExtractions = g.sortExtractionsByPosition(validExtractions)
	}
	
	// Generate highlighted text
	highlightedText, err := g.buildHighlightedText(doc.Text(), validExtractions, colorMap, opts)
	if err != nil {
		return "", NewHTMLGenerationError("failed to build highlighted text", err)
	}
	
	// Prepare extraction data for JavaScript
	extractionData, err := g.prepareExtractionData(doc.Text(), validExtractions, colorMap, opts)
	if err != nil {
		return "", NewHTMLGenerationError("failed to prepare extraction data", err)
	}
	
	// Build legend HTML
	legendHTML := ""
	if opts.ShowLegend {
		legendHTML = g.buildLegendHTML(colorMap)
	}
	
	// Create template variables
	templateVars := &HTMLTemplateVariables{
		CSS:                GetVisualizationCSS(opts),
		Title:              fmt.Sprintf("LangExtract Visualization - %d extractions", len(validExtractions)),
		HighlightedText:    highlightedText,
		LegendHTML:         legendHTML,
		ExtractionData:     extractionData,
		AnimationSpeed:     opts.AnimationSpeed,
		ExtractionCount:    len(validExtractions),
		FirstExtractionPos: g.getFirstExtractionPos(validExtractions),
		GIFOptimized:       opts.GIFOptimized,
		ShowControls:       len(validExtractions) > 1, // Only show controls if multiple extractions
		CustomJavaScript:   g.getCustomJavaScript(opts),
	}
	
	// Render the HTML template
	html, err := g.renderHTMLTemplate(ctx, templateVars, opts)
	if err != nil {
		return "", NewHTMLGenerationError("failed to render HTML template", err)
	}
	
	return html, nil
}

// GetSupportedFormats returns the formats this visualizer supports
func (g *HTMLGenerator) GetSupportedFormats() []OutputFormat {
	return []OutputFormat{OutputFormatHTML}
}

// Name returns the name of the visualizer
func (g *HTMLGenerator) Name() string {
	return "HTMLGenerator"
}

// Validate validates the HTML generator configuration
func (g *HTMLGenerator) Validate() error {
	if g.colorManager == nil {
		return NewValidationError("color manager cannot be nil", nil)
	}
	
	if g.templateRenderer == nil {
		return NewValidationError("template renderer cannot be nil", nil)
	}
	
	if g.options != nil {
		return g.options.Validate()
	}
	
	return nil
}

// filterValidExtractions filters extractions to only include those with valid intervals
func (g *HTMLGenerator) filterValidExtractions(extractions []*extraction.Extraction, filterClasses []string) []*extraction.Extraction {
	if len(extractions) == 0 {
		return nil
	}
	
	// Create filter set if classes are specified
	var filterSet map[string]bool
	if len(filterClasses) > 0 {
		filterSet = make(map[string]bool)
		for _, class := range filterClasses {
			filterSet[class] = true
		}
	}
	
	var validExtractions []*extraction.Extraction
	for _, ext := range extractions {
		if ext == nil || ext.Interval() == nil {
			continue
		}
		
		interval := ext.Interval()
		if interval.Start() < 0 || interval.End() <= interval.Start() {
			continue
		}
		
		// Apply class filter if specified
		if filterSet != nil && !filterSet[ext.Class()] {
			continue
		}
		
		validExtractions = append(validExtractions, ext)
	}
	
	return validExtractions
}

// sortExtractionsByPosition sorts extractions by their position in the text
func (g *HTMLGenerator) sortExtractionsByPosition(extractions []*extraction.Extraction) []*extraction.Extraction {
	sorted := make([]*extraction.Extraction, len(extractions))
	copy(sorted, extractions)
	
	sort.Slice(sorted, func(i, j int) bool {
		intervalI := sorted[i].Interval()
		intervalJ := sorted[j].Interval()
		
		// Sort by start position first
		if intervalI.Start() != intervalJ.Start() {
			return intervalI.Start() < intervalJ.Start()
		}
		
		// For same start position, longer spans come first (for proper HTML nesting)
		lengthI := intervalI.End() - intervalI.Start()
		lengthJ := intervalJ.End() - intervalJ.Start()
		return lengthI > lengthJ
	})
	
	return sorted
}

// buildHighlightedText creates HTML with span highlights for extractions
func (g *HTMLGenerator) buildHighlightedText(text string, extractions []*extraction.Extraction, colorMap map[string]string, opts *VisualizationOptions) (string, error) {
	if len(extractions) == 0 {
		return html.EscapeString(text), nil
	}
	
	// Create span points for proper HTML nesting
	var points []SpanPoint
	spanLengths := make(map[int]int)
	
	for index, ext := range extractions {
		interval := ext.Interval()
		startPos := interval.Start()
		endPos := interval.End()
		
		if startPos < 0 || endPos > len(text) || startPos >= endPos {
			continue // Skip invalid intervals
		}
		
		points = append(points, SpanPoint{
			Position:   startPos,
			TagType:    TagTypeStart,
			SpanIndex:  index,
			Extraction: ext,
		})
		
		points = append(points, SpanPoint{
			Position:   endPos,
			TagType:    TagTypeEnd,
			SpanIndex:  index,
			Extraction: ext,
		})
		
		spanLengths[index] = endPos - startPos
	}
	
	// Sort points for proper HTML nesting
	g.sortSpanPoints(points, spanLengths)
	
	// Build HTML with highlights
	var htmlParts []string
	cursor := 0
	
	for _, point := range points {
		// Add text before the current point
		if point.Position > cursor {
			htmlParts = append(htmlParts, html.EscapeString(text[cursor:point.Position]))
		}
		
		if point.TagType == TagTypeStart {
			// Create opening span tag
			color := colorMap[point.Extraction.Class()]
			if color == "" {
				color = DefaultFallbackColor
			}
			
			highlightClass := ""
			if point.SpanIndex == 0 && opts != nil && len(extractions) > 1 {
				highlightClass = " lx-current-highlight"
			}
			
			debugAttr := ""
			if opts != nil && opts.Debug {
				debugAttr = fmt.Sprintf(` title="Index: %d, Class: %s, Pos: %d-%d"`, 
					point.SpanIndex, point.Extraction.Class(), 
					point.Extraction.Interval().Start(), point.Extraction.Interval().End())
			}
			
			spanHTML := fmt.Sprintf(`<span class="lx-highlight%s" data-idx="%d" style="background-color:%s;"%s>`,
				highlightClass, point.SpanIndex, color, debugAttr)
			htmlParts = append(htmlParts, spanHTML)
			
		} else { // TagTypeEnd
			htmlParts = append(htmlParts, "</span>")
		}
		
		cursor = point.Position
	}
	
	// Add remaining text
	if cursor < len(text) {
		htmlParts = append(htmlParts, html.EscapeString(text[cursor:]))
	}
	
	return strings.Join(htmlParts, ""), nil
}

// sortSpanPoints sorts span points for proper HTML nesting
func (g *HTMLGenerator) sortSpanPoints(points []SpanPoint, spanLengths map[int]int) {
	sort.Slice(points, func(i, j int) bool {
		pointI := points[i]
		pointJ := points[j]
		
		// Sort by position first
		if pointI.Position != pointJ.Position {
			return pointI.Position < pointJ.Position
		}
		
		// For same position, handle nesting
		lengthI := spanLengths[pointI.SpanIndex]
		lengthJ := spanLengths[pointJ.SpanIndex]
		
		// End tags come before start tags
		if pointI.TagType != pointJ.TagType {
			return pointI.TagType == TagTypeEnd
		}
		
		// Among end tags: shorter spans close first
		if pointI.TagType == TagTypeEnd {
			return lengthI < lengthJ
		}
		
		// Among start tags: longer spans open first
		return lengthI > lengthJ
	})
}

// prepareExtractionData prepares JavaScript data for extractions
func (g *HTMLGenerator) prepareExtractionData(text string, extractions []*extraction.Extraction, colorMap map[string]string, opts *VisualizationOptions) (string, error) {
	var extractionData []ExtractionData
	contextChars := opts.ContextChars
	if contextChars <= 0 {
		contextChars = 150
	}
	
	for i, ext := range extractions {
		interval := ext.Interval()
		startPos := interval.Start()
		endPos := interval.End()
		
		// Calculate context window
		contextStart := startPos - contextChars
		if contextStart < 0 {
			contextStart = 0
		}
		
		contextEnd := endPos + contextChars
		if contextEnd > len(text) {
			contextEnd = len(text)
		}
		
		// Extract text segments
		beforeText := text[contextStart:startPos]
		extractionText := text[startPos:endPos]
		afterText := text[endPos:contextEnd]
		
		// Get color
		color := colorMap[ext.Class()]
		if color == "" {
			color = DefaultFallbackColor
		}
		
		// Build attributes HTML
		attributesHTML := g.buildAttributesHTML(ext, opts.IncludeMetadata)
		
		// Prepare metadata if requested
		var metadata map[string]interface{}
		if opts.IncludeMetadata {
			metadata = g.extractMetadata(ext)
		}
		
		extractionData = append(extractionData, ExtractionData{
			Index:          i,
			Class:          ext.Class(),
			Text:           ext.Text(),
			Color:          color,
			StartPos:       startPos,
			EndPos:         endPos,
			BeforeText:     html.EscapeString(beforeText),
			ExtractionText: html.EscapeString(extractionText),
			AfterText:      html.EscapeString(afterText),
			AttributesHTML: attributesHTML,
			Confidence:     ext.Confidence(),
			Metadata:       metadata,
		})
	}
	
	// Marshal to JSON
	jsonData, err := json.Marshal(extractionData)
	if err != nil {
		return "", NewDataProcessingError("failed to marshal extraction data to JSON", err)
	}
	
	return string(jsonData), nil
}

// buildLegendHTML builds the legend HTML showing extraction classes and colors
func (g *HTMLGenerator) buildLegendHTML(colorMap map[string]string) string {
	if len(colorMap) == 0 {
		return ""
	}
	
	// Sort classes for consistent display
	var classes []string
	for class := range colorMap {
		classes = append(classes, class)
	}
	sort.Strings(classes)
	
	var legendItems []string
	for _, class := range classes {
		color := colorMap[class]
		legendItem := fmt.Sprintf(`<span class="lx-label" style="background-color:%s;">%s</span>`,
			color, html.EscapeString(class))
		legendItems = append(legendItems, legendItem)
	}
	
	return fmt.Sprintf(`<div class="lx-legend">Highlights Legend: %s</div>`,
		strings.Join(legendItems, " "))
}

// buildAttributesHTML builds the HTML representation of extraction attributes
func (g *HTMLGenerator) buildAttributesHTML(ext *extraction.Extraction, includeMetadata bool) string {
	var parts []string
	
	// Always include class
	parts = append(parts, fmt.Sprintf(`<div><strong>class:</strong> %s</div>`,
		html.EscapeString(ext.Class())))
	
	// Add confidence if available
	if ext.Confidence() > 0 {
		parts = append(parts, fmt.Sprintf(`<div><strong>confidence:</strong> %.3f</div>`,
			ext.Confidence()))
	}
	
	// Add position information
	interval := ext.Interval()
	parts = append(parts, fmt.Sprintf(`<div><strong>position:</strong> [%d-%d]</div>`,
		interval.Start(), interval.End()))
	
	// Add attributes if available and requested
	if includeMetadata {
		attributes := g.extractAttributes(ext)
		if len(attributes) > 0 {
			attributesStr := g.formatAttributes(attributes)
			parts = append(parts, fmt.Sprintf(`<div><strong>attributes:</strong> %s</div>`,
				attributesStr))
		}
	}
	
	return strings.Join(parts, "")
}

// extractAttributes extracts attributes from an extraction
func (g *HTMLGenerator) extractAttributes(ext *extraction.Extraction) map[string]interface{} {
	// This would be implemented based on the actual extraction interface
	// For now, return empty map
	return make(map[string]interface{})
}

// extractMetadata extracts metadata from an extraction
func (g *HTMLGenerator) extractMetadata(ext *extraction.Extraction) map[string]interface{} {
	metadata := make(map[string]interface{})
	
	// Add basic metadata
	metadata["class"] = ext.Class()
	metadata["text"] = ext.Text()
	metadata["confidence"] = ext.Confidence()
	
	interval := ext.Interval()
	metadata["start_pos"] = interval.Start()
	metadata["end_pos"] = interval.End()
	metadata["length"] = interval.End() - interval.Start()
	
	return metadata
}

// formatAttributes formats attributes as a readable string
func (g *HTMLGenerator) formatAttributes(attributes map[string]interface{}) string {
	if len(attributes) == 0 {
		return "{}"
	}
	
	var parts []string
	for key, value := range attributes {
		if value == nil || value == "" || value == "null" {
			continue
		}
		
		var valueStr string
		switch v := value.(type) {
		case []interface{}:
			var items []string
			for _, item := range v {
				items = append(items, fmt.Sprintf("%v", item))
			}
			valueStr = strings.Join(items, ", ")
		default:
			valueStr = fmt.Sprintf("%v", v)
		}
		
		parts = append(parts, fmt.Sprintf(`<span class="lx-attr-key">%s</span>: <span class="lx-attr-value">%s</span>`,
			html.EscapeString(key), html.EscapeString(valueStr)))
	}
	
	if len(parts) == 0 {
		return "{}"
	}
	
	return "{" + strings.Join(parts, ", ") + "}"
}

// getFirstExtractionPos returns position info for the first extraction
func (g *HTMLGenerator) getFirstExtractionPos(extractions []*extraction.Extraction) string {
	if len(extractions) == 0 {
		return "[0-0]"
	}
	
	interval := extractions[0].Interval()
	return fmt.Sprintf("[%d-%d]", interval.Start(), interval.End())
}

// getCustomJavaScript returns custom JavaScript based on options
func (g *HTMLGenerator) getCustomJavaScript(opts *VisualizationOptions) string {
	if opts == nil {
		return ""
	}
	
	if customJS, exists := opts.TemplateOverrides["javascript"]; exists {
		return customJS
	}
	
	return ""
}

// generateEmptyVisualization generates a visualization for documents with no extractions
func (g *HTMLGenerator) generateEmptyVisualization(opts *VisualizationOptions) (string, error) {
	emptyHTML := `<div class="lx-animated-wrapper"><p>No valid extractions to visualize.</p></div>`
	fullHTML := GetVisualizationCSS(opts) + emptyHTML
	
	if opts.GIFOptimized {
		fullHTML = strings.Replace(fullHTML,
			`class="lx-animated-wrapper"`,
			`class="lx-animated-wrapper lx-gif-optimized"`,
			1)
	}
	
	return fullHTML, nil
}

// renderHTMLTemplate renders the complete HTML template
func (g *HTMLGenerator) renderHTMLTemplate(ctx context.Context, vars *HTMLTemplateVariables, opts *VisualizationOptions) (string, error) {
	// Check for custom template override
	if customTemplate, exists := opts.TemplateOverrides["html"]; exists {
		return g.templateRenderer.Render(ctx, "custom", map[string]interface{}{
			"vars":     vars,
			"template": customTemplate,
		})
	}
	
	// Use default template
	return g.buildDefaultHTMLTemplate(vars), nil
}

// buildDefaultHTMLTemplate builds the default HTML template
func (g *HTMLGenerator) buildDefaultHTMLTemplate(vars *HTMLTemplateVariables) string {
	if len(vars.ExtractionData) == 0 || vars.ExtractionData == "[]" {
		return g.buildEmptyHTMLTemplate(vars)
	}
	
	wrapperClass := "lx-animated-wrapper"
	if vars.GIFOptimized {
		wrapperClass += " lx-gif-optimized"
	}
	
	controlsSection := ""
	if vars.ShowControls {
		controlsSection = g.buildControlsSection(vars)
	}
	
	attributesSection := `<div class="lx-attributes-panel">
		` + vars.LegendHTML + `
		<div id="attributesContainer"></div>
	</div>`
	
	textSection := fmt.Sprintf(`<div class="lx-text-window" id="textWindow">
		%s
	</div>`, vars.HighlightedText)
	
	javascriptSection := ""
	if vars.ShowControls {
		javascriptSection = g.buildJavaScriptSection(vars)
	}
	
	customJS := ""
	if vars.CustomJavaScript != "" {
		customJS = fmt.Sprintf(`<script>
		%s
		</script>`, vars.CustomJavaScript)
	}
	
	return fmt.Sprintf(`%s
<div class="%s">
	%s
	%s
	%s
</div>
%s
%s`, 
		vars.CSS,
		wrapperClass,
		attributesSection,
		textSection,
		controlsSection,
		javascriptSection,
		customJS)
}

// buildEmptyHTMLTemplate builds a template for empty visualizations
func (g *HTMLGenerator) buildEmptyHTMLTemplate(vars *HTMLTemplateVariables) string {
	wrapperClass := "lx-animated-wrapper"
	if vars.GIFOptimized {
		wrapperClass += " lx-gif-optimized"
	}
	
	return fmt.Sprintf(`%s
<div class="%s">
	<p>No valid extractions to visualize.</p>
</div>`, 
		vars.CSS,
		wrapperClass)
}

// buildControlsSection builds the interactive controls section
func (g *HTMLGenerator) buildControlsSection(vars *HTMLTemplateVariables) string {
	return fmt.Sprintf(`<div class="lx-controls">
		<div class="lx-button-row">
			<button class="lx-control-btn" onclick="playPause()">▶️ Play</button>
			<button class="lx-control-btn" onclick="prevExtraction()">⏮ Previous</button>
			<button class="lx-control-btn" onclick="nextExtraction()">⏭ Next</button>
		</div>
		<div class="lx-progress-container">
			<input type="range" id="progressSlider" class="lx-progress-slider"
				   min="0" max="%d" value="0"
				   onchange="jumpToExtraction(this.value)">
		</div>
		<div class="lx-status-text">
			Entity <span id="entityInfo">1/%d</span> |
			Pos <span id="posInfo">%s</span>
		</div>
	</div>`,
		vars.ExtractionCount-1,
		vars.ExtractionCount,
		vars.FirstExtractionPos)
}

// buildJavaScriptSection builds the JavaScript section for interactivity
func (g *HTMLGenerator) buildJavaScriptSection(vars *HTMLTemplateVariables) string {
	return fmt.Sprintf(`<script>
(function() {
	const extractions = %s;
	let currentIndex = 0;
	let isPlaying = false;
	let animationInterval = null;
	let animationSpeed = %f;

	function updateDisplay() {
		const extraction = extractions[currentIndex];
		if (!extraction) return;

		document.getElementById('attributesContainer').innerHTML = extraction.attributesHtml;
		document.getElementById('entityInfo').textContent = (currentIndex + 1) + '/' + extractions.length;
		document.getElementById('posInfo').textContent = '[' + extraction.startPos + '-' + extraction.endPos + ']';
		
		const progressSlider = document.getElementById('progressSlider');
		if (progressSlider) progressSlider.value = currentIndex;

		const playBtn = document.querySelector('.lx-control-btn');
		if (playBtn) playBtn.textContent = isPlaying ? '⏸ Pause' : '▶️ Play';

		// Update highlighting
		const prevHighlight = document.querySelector('.lx-text-window .lx-current-highlight');
		if (prevHighlight) prevHighlight.classList.remove('lx-current-highlight');
		
		const currentSpan = document.querySelector('.lx-text-window span[data-idx="' + currentIndex + '"]');
		if (currentSpan) {
			currentSpan.classList.add('lx-current-highlight');
			currentSpan.scrollIntoView({block: 'center', behavior: 'smooth'});
		}
	}

	function nextExtraction() {
		currentIndex = (currentIndex + 1) %% extractions.length;
		updateDisplay();
	}

	function prevExtraction() {
		currentIndex = (currentIndex - 1 + extractions.length) %% extractions.length;
		updateDisplay();
	}

	function jumpToExtraction(index) {
		currentIndex = parseInt(index);
		updateDisplay();
	}

	function playPause() {
		if (isPlaying) {
			clearInterval(animationInterval);
			isPlaying = false;
		} else {
			animationInterval = setInterval(nextExtraction, animationSpeed * 1000);
			isPlaying = true;
		}
		updateDisplay();
	}

	// Export functions to global scope
	window.playPause = playPause;
	window.nextExtraction = nextExtraction;
	window.prevExtraction = prevExtraction;
	window.jumpToExtraction = jumpToExtraction;

	// Initialize display
	updateDisplay();
})();
</script>`,
		vars.ExtractionData,
		vars.AnimationSpeed)
}