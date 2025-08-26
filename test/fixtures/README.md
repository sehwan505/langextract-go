# Test Fixtures

This directory contains comprehensive test data used across the langextract-go test suite, following patterns from the Python langextract reference implementation.

## Structure

```
fixtures/
├── documents/     # Sample documents for testing extraction
├── examples/      # Few-shot learning examples with expected extractions
├── schemas/       # Schema definitions for structured extraction
├── golden/        # Golden standard outputs for regression testing
└── prompts/       # Test prompts and templates
```

## Documents

### medical_reports.txt
Comprehensive medical report containing various entity types:
- **Person entities**: Patients, doctors, medical staff
- **Medical conditions**: Diagnoses, symptoms, severity levels  
- **Medications**: Drugs, dosages, administration routes
- **Organizations**: Hospitals, clinics, departments
- **Locations**: Cities, states, facility locations
- **Procedures**: Medical interventions, tests, surgeries
- **Vital signs**: BP, HR, temperature, oxygen saturation
- **Dates**: Admission, diagnosis, procedure dates

### business_news.txt
Business news article with financial and corporate entities:
- **Companies**: Public companies with ticker symbols, sectors
- **People**: Executives, analysts, key business figures
- **Financial metrics**: Revenue, growth rates, valuations
- **Locations**: Corporate headquarters, business centers
- **Products**: Software, services, platforms
- **Events**: Earnings announcements, mergers, acquisitions
- **Market indices**: Stock market benchmarks
- **Dates**: Quarterly periods, announcement dates

## Schemas

### medical_schema.json
Comprehensive schema for medical entity extraction:
- **8 entity classes**: person, condition, medication, organization, location, procedure, vital_sign, date
- **Detailed field definitions**: Role, severity, dosage, type constraints
- **Validation rules**: Enums for controlled vocabularies, numeric ranges
- **Global attributes**: Confidence scores, context, source sections

### business_schema.json  
Business entity extraction schema:
- **8 entity classes**: company, person, financial_metric, location, product, event, market_index, date
- **Business-specific constraints**: Ticker symbols, market cap categories
- **Financial validations**: Currency units, metric types, time periods
- **Sentiment analysis**: Positive/negative/neutral sentiment tracking

## Examples

### medical_examples.json
5 curated medical examples demonstrating:
- **Character-level alignment**: Precise start/end positions
- **Rich attributes**: Role, severity, confidence, source sections
- **Complex scenarios**: Multiple extractions per text, overlapping entities
- **Validation patterns**: Schema-compliant attribute structures

### business_examples.json
5 business examples showcasing:
- **Financial entity extraction**: Companies, metrics, market data
- **Relationship modeling**: Person-company associations
- **Sentiment integration**: Market impact assessment
- **Multi-class extraction**: Companies, people, events, metrics in single text

## Golden Files

### medical_golden_output.json
Reference output for medical document processing:
- **Complete extraction metadata**: Model info, processing time, statistics
- **Validation results**: Schema compliance, error/warning reporting
- **Quality metrics**: Confidence distribution, alignment statistics
- **Entity statistics**: Count by class, quality distribution

### business_golden_output.json
Reference output for business news processing:
- **Financial-focused metadata**: Sentiment analysis, relevance scoring
- **Market-specific statistics**: Entity distribution, confidence metrics
- **Comprehensive validation**: Schema compliance verification

## Usage in Tests

### Unit Tests
```go
// Load schema for validation testing
schemaData, _ := os.ReadFile("../fixtures/schemas/medical_schema.json")
schema, _ := extraction.SchemaFromJSON(schemaData)

// Use examples for parameterized tests
exampleData, _ := os.ReadFile("../fixtures/examples/medical_examples.json")
var examples []*extraction.ExampleData
json.Unmarshal(exampleData, &examples)
```

### Integration Tests
```go
// Test with realistic documents
docPath := "../fixtures/documents/medical_reports.txt"
result, err := langextract.Extract(ctx, docPath, options)

// Compare against golden files
goldenData, _ := os.ReadFile("../fixtures/golden/medical_golden_output.json")
// Assert result matches expected output
```

### E2E Tests
```bash
# CLI testing with fixture files
./langextract extract --input fixtures/documents/medical_reports.txt \
                     --schema fixtures/schemas/medical_schema.json \
                     --examples fixtures/examples/medical_examples.json
```

## Validation

### Automated Validation
```bash
# From test directory
make validate-fixtures
```

### Manual Validation
```bash
# Validate JSON structure
python3 -m json.tool fixtures/schemas/medical_schema.json

# Validate schema compliance  
go run ../cmd/langextract validate --schema fixtures/schemas/medical_schema.json \
                                  --examples fixtures/examples/medical_examples.json
```

## Adding New Fixtures

When adding new test fixtures:

1. **Follow naming conventions**: Use descriptive, domain-specific names
2. **Include comprehensive metadata**: Document entity types, use cases, complexity
3. **Validate structure**: Ensure JSON validity and schema compliance
4. **Add documentation**: Update this README with new fixture descriptions
5. **Create golden files**: Generate expected outputs for regression testing
6. **Test integration**: Verify fixtures work across unit, integration, and E2E tests

## Maintenance

- **Regular updates**: Keep fixtures current with evolving schema definitions
- **Performance testing**: Use fixtures for benchmarking and optimization
- **Quality assurance**: Maintain high-quality, realistic test data
- **Version control**: Track changes to ensure test reproducibility