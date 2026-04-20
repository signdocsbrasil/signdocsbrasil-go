package signdocsbrasil

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestWebhookEventType_LockstepWithOpenAPI verifies the Go constants
// match the OpenAPI spec one-for-one. The spec lives two directories up
// at openapi/openapi.yaml in the external-api repo layout. When the
// SDK is published standalone (module/vendored) and the spec isn't
// adjacent, the test skips rather than fails — we do not want to break
// downstream builds.
func TestWebhookEventType_LockstepWithOpenAPI(t *testing.T) {
	specPath := findOpenAPISpec(t)
	if specPath == "" {
		t.Skip("openapi.yaml not found next to the SDK; skipping lockstep check")
	}

	specEvents, err := parseWebhookEventEnum(specPath)
	if err != nil {
		t.Fatalf("failed to parse OpenAPI spec: %v", err)
	}

	sdkEvents := allWebhookEventTypes()

	if len(specEvents) != 17 {
		t.Errorf("expected 17 events in spec, got %d: %v", len(specEvents), specEvents)
	}
	if len(sdkEvents) != 17 {
		t.Errorf("expected 17 events in SDK, got %d: %v", len(sdkEvents), sdkEvents)
	}

	specSet := map[string]bool{}
	for _, e := range specEvents {
		specSet[e] = true
	}
	sdkSet := map[string]bool{}
	for _, e := range sdkEvents {
		sdkSet[string(e)] = true
	}

	for e := range specSet {
		if !sdkSet[e] {
			t.Errorf("spec event %q missing from SDK", e)
		}
	}
	for e := range sdkSet {
		if !specSet[e] {
			t.Errorf("SDK event %q missing from spec", e)
		}
	}
}

func TestIsNT65Event(t *testing.T) {
	nt65 := []WebhookEventType{
		WebhookEventTransactionDeadlineApproaching,
		WebhookEventStepPurposeDisclosureSent,
	}
	for _, e := range nt65 {
		if !IsNT65Event(e) {
			t.Errorf("expected %q to be an NT65 event", e)
		}
	}

	nonNT65 := []WebhookEventType{
		WebhookEventTransactionCreated,
		WebhookEventTransactionCompleted,
		WebhookEventStepStarted,
		WebhookEventQuotaWarning,
		WebhookEventAPIDeprecation,
		WebhookEventSigningSessionCreated,
		WebhookEventSigningSessionExpired,
	}
	for _, e := range nonNT65 {
		if IsNT65Event(e) {
			t.Errorf("expected %q NOT to be an NT65 event", e)
		}
	}
}

func TestWebhookEventType_DeprecatedAliases(t *testing.T) {
	// The 1.2.x truncated names must keep pointing at the full strings.
	if WebhookEventStepPurposeDisclosure != WebhookEventStepPurposeDisclosureSent {
		t.Error("WebhookEventStepPurposeDisclosure drifted from canonical value")
	}
	if WebhookEventTransactionDeadline != WebhookEventTransactionDeadlineApproaching {
		t.Error("WebhookEventTransactionDeadline drifted from canonical value")
	}
	if string(WebhookEventStepPurposeDisclosure) != "STEP.PURPOSE_DISCLOSURE_SENT" {
		t.Errorf("unexpected value: %q", WebhookEventStepPurposeDisclosure)
	}
	if string(WebhookEventTransactionDeadline) != "TRANSACTION.DEADLINE_APPROACHING" {
		t.Errorf("unexpected value: %q", WebhookEventTransactionDeadline)
	}
}

// allWebhookEventTypes returns every WebhookEventType constant exposed
// by the SDK. Must be kept in sync with models.go — any new event
// should appear here too.
func allWebhookEventTypes() []WebhookEventType {
	return []WebhookEventType{
		WebhookEventTransactionCreated,
		WebhookEventTransactionCompleted,
		WebhookEventTransactionCancelled,
		WebhookEventTransactionFailed,
		WebhookEventTransactionExpired,
		WebhookEventTransactionFallback,
		WebhookEventTransactionDeadlineApproaching,
		WebhookEventStepStarted,
		WebhookEventStepCompleted,
		WebhookEventStepFailed,
		WebhookEventStepPurposeDisclosureSent,
		WebhookEventQuotaWarning,
		WebhookEventAPIDeprecation,
		WebhookEventSigningSessionCreated,
		WebhookEventSigningSessionCompleted,
		WebhookEventSigningSessionCancelled,
		WebhookEventSigningSessionExpired,
	}
}

// findOpenAPISpec walks a few known relative paths to locate the spec.
// Returns "" if nothing plausible is found.
func findOpenAPISpec(t *testing.T) string {
	t.Helper()

	candidates := []string{
		"../../openapi/openapi.yaml",
		"../../../openapi/openapi.yaml",
		"../../../external-api/openapi/openapi.yaml",
	}
	for _, rel := range candidates {
		abs, err := filepath.Abs(rel)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}
	return ""
}

// parseWebhookEventEnum extracts the `WebhookEventType.enum` values
// from the OpenAPI spec without pulling in a YAML dependency. We look
// for the `WebhookEventType:` section and collect subsequent `- FOO`
// lines until we hit a non-enum line.
func parseWebhookEventEnum(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var events []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 4*1024*1024)

	state := 0 // 0=searching, 1=inside WebhookEventType block, 2=inside enum list
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		switch state {
		case 0:
			if strings.HasPrefix(trimmed, "WebhookEventType:") {
				state = 1
			}
		case 1:
			if strings.HasPrefix(trimmed, "enum:") {
				state = 2
			} else if trimmed != "" && !strings.HasPrefix(trimmed, "type:") && !strings.HasPrefix(trimmed, "description:") && !strings.HasPrefix(trimmed, "|") && !strings.HasPrefix(line, " ") {
				// Out of the WebhookEventType block without finding enum.
				return nil, nil
			}
		case 2:
			if strings.HasPrefix(trimmed, "- ") {
				val := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
				// Strip quotes if present.
				val = strings.Trim(val, `"'`)
				events = append(events, val)
			} else {
				// End of enum list.
				sort.Strings(events)
				return events, nil
			}
		}
	}

	sort.Strings(events)
	return events, scanner.Err()
}
