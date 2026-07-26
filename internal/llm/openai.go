package llm

import (
	"fmt"
	"os"
)

// OpenAIBaseURL is OpenAI's API root.
const OpenAIBaseURL = "https://api.openai.com/v1"

// EnvOpenAIKey is the environment variable holding the OpenAI API key.
const EnvOpenAIKey = "OPENAI_API_KEY"

// NewOpenAI returns an OpenAI provider using the given API key.
func NewOpenAI(apiKey string, opts ...Option) Provider {
	return newHTTPProvider("openai", OpenAIBaseURL, apiKey, opts...)
}

// OpenAIFromEnv returns an OpenAI provider configured from OPENAI_API_KEY.
func OpenAIFromEnv(opts ...Option) (Provider, error) {
	key := os.Getenv(EnvOpenAIKey)
	if key == "" {
		return nil, fmt.Errorf(
			"%s is not set — create an API key at https://platform.openai.com/api-keys and run: export %s=<your-key>",
			EnvOpenAIKey, EnvOpenAIKey,
		)
	}
	return NewOpenAI(key, opts...), nil
}
