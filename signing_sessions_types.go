package signdocsbrasil

// CreateSigningSessionRequest represents a request to create a signing session.
type CreateSigningSessionRequest struct {
	Purpose          string                    `json:"purpose"`
	Policy           Policy                    `json:"policy"`
	Signer           Signer                    `json:"signer"`
	Document         *DocumentRequest          `json:"document,omitempty"`
	Action           *ActionMetadata           `json:"action,omitempty"`
	ReturnURL        string                    `json:"returnUrl,omitempty"`
	CancelURL        string                    `json:"cancelUrl,omitempty"`
	Metadata         map[string]string         `json:"metadata,omitempty"`
	Locale           string                    `json:"locale,omitempty"`
	ExpiresInMinutes int                       `json:"expiresInMinutes,omitempty"`
	Appearance       *SigningSessionAppearance `json:"appearance,omitempty"`
}

// DocumentRequest represents an inline document.
type DocumentRequest struct {
	Content  string `json:"content"`
	Filename string `json:"filename,omitempty"`
}

// SigningSessionAppearance controls the look and feel of the signing UI.
type SigningSessionAppearance struct {
	BrandColor      string `json:"brandColor,omitempty"`
	LogoURL         string `json:"logoUrl,omitempty"`
	CompanyName     string `json:"companyName,omitempty"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	TextColor       string `json:"textColor,omitempty"`
	ButtonTextColor string `json:"buttonTextColor,omitempty"`
	BorderRadius    string `json:"borderRadius,omitempty"`
	HeaderStyle     string `json:"headerStyle,omitempty"`
	FontFamily      string `json:"fontFamily,omitempty"`
}

// SigningSession is returned when a signing session is created.
type SigningSession struct {
	SessionID     string `json:"sessionId"`
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"`
	URL           string `json:"url"`
	ClientSecret  string `json:"clientSecret"`
	ExpiresAt     string `json:"expiresAt"`
	CreatedAt     string `json:"createdAt"`
}

// SigningSessionStatus is the lightweight status used for polling.
type SigningSessionStatus struct {
	SessionID     string `json:"sessionId"`
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"`
	CompletedAt   string `json:"completedAt,omitempty"`
	EvidenceID    string `json:"evidenceId,omitempty"`
}

// CancelSigningSessionResponse is returned after cancelling a session.
type CancelSigningSessionResponse struct {
	SessionID     string `json:"sessionId"`
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"`
	CancelledAt   string `json:"cancelledAt"`
}

// SigningSessionListParams contains filter and pagination parameters.
type SigningSessionListParams struct {
	Status string
	Limit  int
	Cursor string
}

// SigningSessionListResponse contains the paginated list of sessions.
type SigningSessionListResponse struct {
	Sessions   []SigningSessionListItem `json:"sessions"`
	NextCursor string                   `json:"nextCursor,omitempty"`
}

// SigningSessionListItem is a session in a list response.
type SigningSessionListItem struct {
	SessionID     string `json:"sessionId"`
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"`
	CreatedAt     string `json:"createdAt"`
	ExpiresAt     string `json:"expiresAt"`
	Locale        string `json:"locale"`
}

// AdvanceSessionAction represents the action to perform when advancing a session.
type AdvanceSessionAction string

const (
	AdvanceActionAccept           AdvanceSessionAction = "accept"
	AdvanceActionVerifyOTP        AdvanceSessionAction = "verify_otp"
	AdvanceActionResendOTP        AdvanceSessionAction = "resend_otp"
	AdvanceActionStartLiveness    AdvanceSessionAction = "start_liveness"
	AdvanceActionCompleteLiveness AdvanceSessionAction = "complete_liveness"
	AdvanceActionPrepareSigning   AdvanceSessionAction = "prepare_signing"
	AdvanceActionCompleteSigning  AdvanceSessionAction = "complete_signing"
)

// AdvanceSessionRequest is the request body for advancing a signing session step.
type AdvanceSessionRequest struct {
	Action              AdvanceSessionAction `json:"action"`
	OtpCode             string               `json:"otpCode,omitempty"`
	LivenessSessionID   string               `json:"livenessSessionId,omitempty"`
	CertificateChainPems []string             `json:"certificateChainPems,omitempty"`
	SignatureRequestID  string               `json:"signatureRequestId,omitempty"`
	RawSignatureBase64  string               `json:"rawSignatureBase64,omitempty"`
	Geolocation         *Geolocation         `json:"geolocation,omitempty"`
}

// AdvanceSessionStep represents a step in an advance session response.
type AdvanceSessionStep struct {
	StepID string `json:"stepId"`
	Type   string `json:"type"`
	Status string `json:"status,omitempty"`
}

// SandboxData contains sandbox-only data (HML environment).
type SandboxData struct {
	OtpCode string `json:"otpCode,omitempty"`
}

// AdvanceSessionResponse is returned when a signing session step is advanced.
type AdvanceSessionResponse struct {
	SessionID          string              `json:"sessionId"`
	Status             string              `json:"status"`
	CurrentStep        *AdvanceSessionStep `json:"currentStep,omitempty"`
	NextStep           *AdvanceSessionStep `json:"nextStep,omitempty"`
	EvidenceID         string              `json:"evidenceId,omitempty"`
	RedirectURL        string              `json:"redirectUrl,omitempty"`
	CompletedAt        string              `json:"completedAt,omitempty"`
	HostedURL          string              `json:"hostedUrl,omitempty"`
	LivenessSessionID  string              `json:"livenessSessionId,omitempty"`
	SignatureRequestID string              `json:"signatureRequestId,omitempty"`
	HashToSign         string              `json:"hashToSign,omitempty"`
	HashAlgorithm      string              `json:"hashAlgorithm,omitempty"`
	SignatureAlgorithm string              `json:"signatureAlgorithm,omitempty"`
	Sandbox            *SandboxData        `json:"sandbox,omitempty"`
}

// BootstrapSigner contains masked signer information in a bootstrap response.
type BootstrapSigner struct {
	Name        string `json:"name"`
	MaskedEmail string `json:"maskedEmail,omitempty"`
	MaskedCPF   string `json:"maskedCpf,omitempty"`
}

// BootstrapStep represents a step in a bootstrap response.
type BootstrapStep struct {
	StepID string `json:"stepId"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Order  int    `json:"order"`
}

// BootstrapDocument contains document info in a bootstrap response.
type BootstrapDocument struct {
	PresignedURL string `json:"presignedUrl,omitempty"`
	Filename     string `json:"filename,omitempty"`
	Hash         string `json:"hash,omitempty"`
}

// SigningSessionBootstrap contains full bootstrap data for a signing session.
type SigningSessionBootstrap struct {
	SessionID     string                    `json:"sessionId"`
	TransactionID string                    `json:"transactionId"`
	Status        string                    `json:"status"`
	Purpose       string                    `json:"purpose"`
	Signer        BootstrapSigner           `json:"signer"`
	Steps         []BootstrapStep           `json:"steps"`
	Locale        string                    `json:"locale"`
	ExpiresAt     string                    `json:"expiresAt"`
	Document      *BootstrapDocument        `json:"document,omitempty"`
	Action        *ActionMetadata           `json:"action,omitempty"`
	Appearance    *SigningSessionAppearance  `json:"appearance,omitempty"`
	ReturnURL     string                    `json:"returnUrl,omitempty"`
	CancelURL     string                    `json:"cancelUrl,omitempty"`
}
