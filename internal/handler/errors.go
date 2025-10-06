package handler

const (
	ErrCodeUnauthorized         = "UNAUTHORIZED"
	ErrCodeInvalidCredentials   = "INVALID_CREDENTIALS"
	ErrCodeUserNotFound         = "USER_NOT_FOUND"
	ErrCodeEmailTaken           = "EMAIL_TAKEN"
	ErrCodeUsernameTaken        = "USERNAME_TAKEN"
	ErrCodePostNotFound         = "POST_NOT_FOUND"
	ErrCodeSlugTaken            = "SLUG_TAKEN"
	ErrCodePostAlreadyPublished = "POST_ALREADY_PUBLISHED"
	ErrCodeInvalidStatusChange  = "INVALID_STATUS_CHANGE"
	ErrCodeForbidden            = "FORBIDDEN"
	ErrCodeValidationFailed     = "VALIDATION_FAILED"
	ErrCodeInternalServer       = "INTERNAL_SERVER_ERROR"
	ErrCodeConflict             = "CONFLICT"
)
