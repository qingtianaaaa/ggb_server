package aiModule

type ChatCompletionInterface interface {
	ChatCompletion() (Content, error)
	ChatCompletionStream() (Content, error)
}

type ProcessTemplate interface {
	Classify(string) (map[string]string, error)
	ExtractElementsStream(map[string]string, string) (string, error)
	GenerateGGB(string) (string, error)
	GenerateHTML(string) (string, error)
}
