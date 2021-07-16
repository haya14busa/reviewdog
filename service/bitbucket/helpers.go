package bitbucket

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/reviewdog/reviewdog/proto/rdf"
	"google.golang.org/protobuf/proto"
)

func hash(b []byte) string {
	h := sha256.New()
	_, _ = h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Note that it might be good to create external ReportID from the diagnostic
// content along with *original line* (by using git blame for example) to
// generate unique ReportID, but it hashes the Diagnostic message for simplicity.
func externalIDFromDiagnostic(d *rdf.Diagnostic) string {
	b, err := proto.Marshal(d)
	if err != nil {
		b = []byte(d.OriginalOutput)
	}
	return hash(b)
}

func reportID(ids ...string) string {
	return strings.ReplaceAll(strings.ToLower(strings.Join(ids, "-")), " ", "_")
}

func reportTitle(tool, reporter string) string {
	return fmt.Sprintf("[%s] %s report", tool, reporter)
}

func convertSeverity(severity rdf.Severity) string {
	switch severity {
	case rdf.Severity_INFO:
		return annotationSeverityLow

	case rdf.Severity_WARNING:
		return annotationSeverityMedium

	case rdf.Severity_ERROR:
		return annotationSeverityHigh

	default:
		return ""
	}
}
