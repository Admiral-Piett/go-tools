package models

// QUESTION - Could I also do these with gin.H?
func init() {
	ErrorResponses = &errorResponses{
		GeneralError: ErrorResponse{
			Code:    "GENERAL_ERROR",
			Message: "Server error, please try again",
		},
		BadRequest: ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Bad Request",
		},
		ValidationError: ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Validation Error",
		},
		UnauthorizedError: ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Unauthorized",
		},
	}
}

var ErrorResponses *errorResponses

type errorResponses struct {
	GeneralError      ErrorResponse
	BadRequest        ErrorResponse
	ValidationError   ErrorResponse
	UnauthorizedError ErrorResponse
}
