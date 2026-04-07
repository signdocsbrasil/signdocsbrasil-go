package signdocsbrasil

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// TransactionsService provides access to transaction CRUD operations.
type TransactionsService struct {
	http *httpClient
}

func newTransactionsService(h *httpClient) *TransactionsService {
	return &TransactionsService{http: h}
}

// CreateOption is a functional option for transaction creation.
type CreateOption func(*createOptions)

type createOptions struct {
	idempotencyKey string
}

// WithIdempotencyKey sets a specific idempotency key for the request.
// If not provided, a random UUID is generated automatically.
func WithIdempotencyKey(key string) CreateOption {
	return func(o *createOptions) {
		o.idempotencyKey = key
	}
}

// Create creates a new transaction. An X-Idempotency-Key header is automatically
// included. Use WithIdempotencyKey to provide a specific key.
func (s *TransactionsService) Create(ctx context.Context, req *CreateTransactionRequest, opts ...CreateOption) (*Transaction, error) {
	o := &createOptions{}
	for _, opt := range opts {
		opt(o)
	}

	var result Transaction
	err := s.http.requestWithIdempotency(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   "/v1/transactions",
		Body:   req,
	}, &result, o.idempotencyKey)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// List returns a paginated list of transactions matching the optional filter parameters.
func (s *TransactionsService) List(ctx context.Context, params *TransactionListParams) (*TransactionListResponse, error) {
	query := make(map[string]string)
	if params != nil {
		if params.Status != "" {
			query["status"] = string(params.Status)
		}
		if params.UserExternalID != "" {
			query["userExternalId"] = params.UserExternalID
		}
		if params.DocumentGroupID != "" {
			query["documentGroupId"] = params.DocumentGroupID
		}
		if params.StartDate != "" {
			query["startDate"] = params.StartDate
		}
		if params.EndDate != "" {
			query["endDate"] = params.EndDate
		}
		if params.Limit > 0 {
			query["limit"] = strconv.Itoa(params.Limit)
		}
		if params.NextToken != "" {
			query["nextToken"] = params.NextToken
		}
	}

	var result TransactionListResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   "/v1/transactions",
		Query:  query,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves a single transaction by ID.
func (s *TransactionsService) Get(ctx context.Context, transactionID string) (*Transaction, error) {
	var result Transaction
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/v1/transactions/%s", transactionID),
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Cancel cancels a transaction. Note: this returns 200 with a JSON body, not 204.
func (s *TransactionsService) Cancel(ctx context.Context, transactionID string) (*CancelTransactionResponse, error) {
	var result CancelTransactionResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodDelete,
		Path:   fmt.Sprintf("/v1/transactions/%s", transactionID),
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Finalize completes a transaction after all steps are done.
func (s *TransactionsService) Finalize(ctx context.Context, transactionID string) (*FinalizeResponse, error) {
	var result FinalizeResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1/transactions/%s/finalize", transactionID),
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListAutoPaginate returns a PageIterator that automatically fetches all pages
// of transactions matching the given parameters. The context for each page
// fetch is provided via the PageIterator.Next(ctx) call.
func (s *TransactionsService) ListAutoPaginate(params *TransactionListParams) *PageIterator[Transaction] {
	return newPageIterator(func(fetchCtx context.Context, nextToken string) ([]Transaction, string, error) {
		p := &TransactionListParams{}
		if params != nil {
			*p = *params
		}
		p.NextToken = nextToken

		resp, err := s.List(fetchCtx, p)
		if err != nil {
			return nil, "", err
		}

		return resp.Transactions, resp.NextToken, nil
	})
}
