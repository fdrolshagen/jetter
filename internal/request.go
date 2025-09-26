package internal

type Request struct {
	Name    string
	Method  string
	Url     string
	Headers map[string]string
	Body    string
}
