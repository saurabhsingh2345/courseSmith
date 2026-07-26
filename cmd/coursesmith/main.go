// Command coursesmith turns markdown lesson outlines into complete video
// course lessons: scripts, TTS audio, captions, diagrams, quizzes, and MP4s.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
