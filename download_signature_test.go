package signdocsbrasil

import (
	"encoding/json"
	"testing"
)

// Non-PDF transactions come back as documentFormat "generic" with a detached
// CAdES signature instead of an embedded signedUrl.
func TestDownloadResponseUnmarshalsDetachedSignature(t *testing.T) {
	body := `{"transactionId":"tx_2","expiresIn":900,"documentFormat":"generic",` +
		`"originalUrl":"https://s3.example.com/document.docx",` +
		`"signatureUrl":"https://s3.example.com/signature.p7s"}`

	var resp DownloadResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.DocumentFormat != "generic" {
		t.Errorf("DocumentFormat = %q, want %q", resp.DocumentFormat, "generic")
	}
	if resp.SignatureURL != "https://s3.example.com/signature.p7s" {
		t.Errorf("SignatureURL = %q", resp.SignatureURL)
	}
	if resp.SignedURL != "" {
		t.Errorf("SignedURL = %q, want empty", resp.SignedURL)
	}
}

// The new fields must stay out of the wire format when unset.
func TestDownloadResponseOmitsEmptyDetachedFields(t *testing.T) {
	out, err := json.Marshal(DownloadResponse{TransactionID: "tx_1", ExpiresIn: 60})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(out)
	if want := `{"transactionId":"tx_1","expiresIn":60}`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
