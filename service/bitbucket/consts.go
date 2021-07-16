package bitbucket

import "time"

const (
	httpTimeout = time.Second * 10

	reportTypeBug = "BUG"
	// reportTypeSecurity = "SECURITY"
	// reportTypeCoverage = "COVERAGE"
	// reportTypeTest     = "TEST"

	reportResultPassed  = "PASSED"
	reportResultFailed  = "FAILED"
	reportResultPending = "PENDING"

	annotationTypeCodeSmell = "CODE_SMELL"
	// annotationTypeVulnerability = "VULNERABILITY"
	// annotationTypeBug           = "BUG"

	annotationSeverityHigh   = "HIGH"
	annotationSeverityMedium = "MEDIUM"
	annotationSeverityLow    = "LOW"
	// annotationSeverityCritical = "CRITICAL"

	// list possible, but not used for now annotation results
	// annotationResultPassed  = "PASSED"
	// annotationResultFailed  = "FAILED"
	// annotationResultSkipped = "SKIPPED"
	// annotationResultIgnored = "IGNORED"
	// annotationResultPending = "PENDING"

	// list of possible, but not used for now
	// report data types
	// reportDataTypeBool       = "BOOLEAN"
	// reportDataTypeDate       = "DATE"
	// reportDataTypeDuration   = "DURATION"
	// reportDataTypeLink       = "LINK"
	// reportDataTypeNumber     = "NUMBER"
	// reportDataTypePercentage = "PERCENTAGE"
	// reportDataTypeText       = "TEXT"
)
