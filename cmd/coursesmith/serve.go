package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/studio"
)

// studioUIDirName is where the built frontend lands (studio/dist).
const studioUIDirName = "studio/dist"

// newServeCmd runs the Studio: JSON API + SSE + artifact serving, plus the
// built React frontend when studio/dist exists.
func newServeCmd() *cobra.Command {
	var addr string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the Studio: JSON API + SSE + the browser UI (if built)",
		Long: "Serves the coursesmith Studio on localhost: project state as JSON,\n" +
			"pipeline runs with live SSE progress, artifact files, the token/cost\n" +
			"ledger, and the React UI from studio/dist when it has been built\n" +
			"(cd studio && npm install && npm run build).",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			env := newEnv(cmd)
			uiDir := ""
			if _, err := os.Stat(filepath.Join(studioUIDirName, "index.html")); err == nil {
				uiDir = studioUIDirName
			}
			server := studio.New(env, coursesDirName, llm.DefaultStateDir, uiDir)

			srv := &http.Server{
				Addr:              addr,
				Handler:           server.Handler(),
				ReadHeaderTimeout: 5 * time.Second,
			}
			go func() {
				<-ctx.Done()
				shutdownCtx, done := context.WithTimeout(context.Background(), 3*time.Second)
				defer done()
				_ = srv.Shutdown(shutdownCtx)
			}()

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "coursesmith studio on http://%s\n", addr)
			if uiDir == "" {
				fmt.Fprintf(out, "  UI not built yet — cd studio && npm install && npm run build\n")
			}
			fmt.Fprintf(out, "  API: /api/courses /api/lessons/{course}/{id} /api/run /api/events /api/ledger /api/openapi.json\n")
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:8787", "listen address (localhost-only by default)")
	return cmd
}
