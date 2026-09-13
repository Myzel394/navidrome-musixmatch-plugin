package utils

type HTTPResponse struct {
	Body       []byte
	StatusCode int
	Headers    map[string]string
}
