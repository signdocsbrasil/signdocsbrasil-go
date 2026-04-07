package signdocsbrasil

import (
	"context"
	"fmt"
	"net/http"
)

// DocumentsService provides access to document upload, presign, confirm, and download.
type DocumentsService struct {
	http *httpClient
}

func newDocumentsService(h *httpClient) *DocumentsService {
	return &DocumentsService{http: h}
}

// Upload uploads a document to a transaction using inline base64 content.
func (s *DocumentsService) Upload(ctx context.Context, transactionID string, req *UploadDocumentRequest) (*DocumentUploadResponse, error) {
	var result DocumentUploadResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1/transactions/%s/document", transactionID),
		Body:   req,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Presign generates a presigned URL for uploading a document directly to S3.
func (s *DocumentsService) Presign(ctx context.Context, transactionID string, req *PresignRequest) (*PresignResponse, error) {
	var result PresignResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1/transactions/%s/document/presign", transactionID),
		Body:   req,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Confirm confirms a presigned document upload after the file has been uploaded to S3.
func (s *DocumentsService) Confirm(ctx context.Context, transactionID string, req *ConfirmDocumentRequest) (*ConfirmDocumentResponse, error) {
	var result ConfirmDocumentResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1/transactions/%s/document/confirm", transactionID),
		Body:   req,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Download generates a presigned download URL for the transaction's document.
func (s *DocumentsService) Download(ctx context.Context, transactionID string) (*DownloadResponse, error) {
	var result DownloadResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/v1/transactions/%s/download", transactionID),
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
