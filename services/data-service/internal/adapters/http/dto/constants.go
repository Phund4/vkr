package dto

import "time"

const (
	StatusOK    = true
	StatusError = "error"

	MessageOK = "ok"

	ErrProductIDQueryField         = "query param product_id is required"
	ErrProductIDQueryFieldType     = "query param product_id must be int64"
	ErrProductCreationNotFound     = "product creation not found"
	ErrInvalidRateLimitIdentifier  = "invalid rate limit identifier"
	ErrApiRateLimitExceeded        = "api rate limit exceeded"
	ErrGetProductRateLimitExceeded = "get_product rate limit exceeded"

	ApiV1RateLimitRPS         = 40
	ApiV1RateLimitBurst       = 80
	GetProductRateLimitRPS    = 10
	GetProductRateLimitBurst  = 20
	RateLimitExpiresInMinutes = 3  * time.Minute
)
