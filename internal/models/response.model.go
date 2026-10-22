package models

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Results interface{} `json:"results"`
}

func SuccessResponse(message string, results interface{}) Response {
	return Response{
		Success: true,
		Message: message,
		Results: results,
	}
}

func ErrorResponse(message string) Response {
	return Response{
		Success: false,
		Message: message,
		Results: nil,
	}
}
