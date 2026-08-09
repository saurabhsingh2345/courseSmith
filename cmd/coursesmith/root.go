package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/enfec/coursesmith/internal/project"
)

// coursesDirName is where courses live relative to the project root.
const coursesDirName = "courses"

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "coursesmith",
		Short: "AI course production engine",
		Long: "coursesmith turns markdown lesson outlines into complete video course\n" +
			"lessons: voiceover scripts, TTS audio, captions, SVG diagrams, quizzes,\n" +
			"and assembled MP4s.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newInitCmd())
	root.AddCommand(newAuthorCmd())
	root.AddCommand(newSnippetCmd())
	root.AddCommand(newComboCmd())
	root.AddCommand(newNoCodeCmd())
	root.AddCommand(newRunCmd())
	root.AddCommand(newStatusCmd())
	root.AddCommand(newDoctorCmd())
	root.AddCommand(newFootageCmd())
	root.AddCommand(newPreviewCmd())
	root.AddCommand(newServeCmd())
	root.AddCommand(newAuditionCmd())
	root.AddCommand(newAnalyzeCmd())
	root.AddCommand(newCompileCourseCmd())
	root.AddCommand(newExportReviewCmd())
	root.AddCommand(newBuildSiteCmd())
	root.AddCommand(newBundleCmd())
	root.AddCommand(newEbookCmd())
	return root
}

// resolveCourse loads a course from an argument that is either a course slug
// (looked up under ./courses/) or a path to a course directory.
func resolveCourse(arg string) (*project.Course, error) {
	candidates := []string{
		filepath.Join(coursesDirName, arg),
		arg,
	}
	for _, dir := range candidates {
		if _, err := os.Stat(filepath.Join(dir, project.CourseFileName)); err == nil {
			return project.LoadCourse(dir)
		}
	}
	return nil, fmt.Errorf("no course %q: looked for %s in %s and %s",
		arg,
		project.CourseFileName,
		filepath.Join(coursesDirName, arg),
		arg,
	)
}
