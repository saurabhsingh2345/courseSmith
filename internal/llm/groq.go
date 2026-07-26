package llm

import (
	"fmt"
	"os"
)

// GroqBaseURL is Groq's OpenAI-compatible API root.
const GroqBaseURL = "https://api.groq.com/openai/v1"

// EnvGroqKey is the environment variable holding the Groq API key.
const EnvGroqKey = "GROQ_API_KEY"

// NewGroq returns a Groq provider using the given API key.
func NewGroq(apiKey string, opts ...Option) Provider {
	return newHTTPProvider("groq", GroqBaseURL, apiKey, opts...)
}

// GroqFromEnv returns a Groq provider configured from GROQ_API_KEY.
func GroqFromEnv(opts ...Option) (Provider, error) {
	key := os.Getenv(EnvGroqKey)
	if key == "" {
		return nil, fmt.Errorf(
			"%s is not set — create a free API key at https://console.groq.com/keys and run: export %s=<your-key>",
			EnvGroqKey, EnvGroqKey,
		)
	}
	return NewGroq(key, opts...), nil
}
