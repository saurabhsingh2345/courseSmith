package main

import "github.com/enfec/coursesmith/internal/studio"

func main() {
	if err := studio.WriteOpenAPISpec("studio/openapi.json"); err != nil {
		panic(err)
	}
}
