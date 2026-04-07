package signdocsbrasil

// CreateEnvelopeRequest represents a request to create a new envelope.
type CreateEnvelopeRequest struct {
	SigningMode      string            `json:"signingMode"`
	TotalSigners     int               `json:"totalSigners"`
	Document         EnvelopeDocument  `json:"document"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	Locale           string            `json:"locale,omitempty"`
	ReturnURL        string            `json:"returnUrl,omitempty"`
	CancelURL        string            `json:"cancelUrl,omitempty"`
	ExpiresInMinutes *int              `json:"expiresInMinutes,omitempty"`
}

// EnvelopeDocument represents an inline document for an envelope.
type EnvelopeDocument struct {
	Content  string `json:"content"`
	Filename string `json:"filename,omitempty"`
}

// Envelope is returned when an envelope is created.
type Envelope struct {
	EnvelopeID   string `json:"envelopeId"`
	Status       string `json:"status"`
	SigningMode  string `json:"signingMode"`
	TotalSigners int    `json:"totalSigners"`
	DocumentHash string `json:"documentHash"`
	CreatedAt    string `json:"createdAt"`
	ExpiresAt    string `json:"expiresAt"`
}

// AddEnvelopeSessionRequest represents a request to add a signing session to an envelope.
type AddEnvelopeSessionRequest struct {
	Signer      EnvelopeSessionSigner `json:"signer"`
	Policy      EnvelopeSessionPolicy `json:"policy"`
	Purpose     string                `json:"purpose,omitempty"`
	SignerIndex int                   `json:"signerIndex"`
	ReturnURL   string                `json:"returnUrl,omitempty"`
	CancelURL   string                `json:"cancelUrl,omitempty"`
	Metadata    map[string]string     `json:"metadata,omitempty"`
}

// EnvelopeSessionSigner represents the signer details for an envelope session.
type EnvelopeSessionSigner struct {
	Name           string `json:"name"`
	UserExternalID string `json:"userExternalId"`
	CPF            string `json:"cpf,omitempty"`
	CNPJ           string `json:"cnpj,omitempty"`
	Email          string `json:"email,omitempty"`
	Phone          string `json:"phone,omitempty"`
	BirthDate      string `json:"birthDate,omitempty"`
	OTPChannel     string `json:"otpChannel,omitempty"`
}

// EnvelopeSessionPolicy represents the verification policy for an envelope session.
type EnvelopeSessionPolicy struct {
	Profile string `json:"profile"`
}

// EnvelopeSession is returned when a session is added to an envelope.
type EnvelopeSession struct {
	SessionID     string `json:"sessionId"`
	TransactionID string `json:"transactionId"`
	SignerIndex   int    `json:"signerIndex"`
	Status        string `json:"status"`
	URL           string `json:"url"`
	ClientSecret  string `json:"clientSecret"`
	ExpiresAt     string `json:"expiresAt"`
}

// EnvelopeSessionSummary represents a session summary within an envelope detail.
type EnvelopeSessionSummary struct {
	SessionID     string `json:"sessionId"`
	TransactionID string `json:"transactionId"`
	SignerIndex   int    `json:"signerIndex"`
	SignerName    string `json:"signerName"`
	Status        string `json:"status"`
	CompletedAt   string `json:"completedAt,omitempty"`
	EvidenceID    string `json:"evidenceId,omitempty"`
}

// EnvelopeDetail is returned when retrieving an envelope by ID.
type EnvelopeDetail struct {
	EnvelopeID           string                   `json:"envelopeId"`
	Status               string                   `json:"status"`
	SigningMode          string                   `json:"signingMode"`
	TotalSigners         int                      `json:"totalSigners"`
	AddedSessions        int                      `json:"addedSessions"`
	CompletedSessions    int                      `json:"completedSessions"`
	DocumentHash         string                   `json:"documentHash"`
	Sessions             []EnvelopeSessionSummary `json:"sessions"`
	CreatedAt            string                   `json:"createdAt"`
	UpdatedAt            string                   `json:"updatedAt"`
	ExpiresAt            string                   `json:"expiresAt"`
	CombinedSignedPdfURL string                   `json:"combinedSignedPdfUrl,omitempty"`
}

// EnvelopeCombinedStampResponse is returned when requesting the combined stamp for an envelope.
type EnvelopeCombinedStampResponse struct {
	EnvelopeID  string `json:"envelopeId"`
	DownloadURL string `json:"downloadUrl"`
	ExpiresIn   int    `json:"expiresIn"`
	SignerCount int    `json:"signerCount"`
}
