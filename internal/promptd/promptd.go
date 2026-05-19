package promptd

type Config struct {
	Defaults  Defaults
	Providers map[string]Provider
}

type Defaults struct {
	Chat  string
	FIM   string
	Task  string
	Embed string
}

type Provider struct {
	Scheme   string
	Endpoint string
	Key      string
}
