// The route table, and nothing else. The shell it renders into lives in
// layout/StudioLayout.tsx.

import { useCallback, useState } from "react";
import { Link, Route, Routes } from "react-router-dom";
import { StudioLayout, useShellShortcuts } from "./layout/StudioLayout";
import { useShortcutContext } from "./state/ShortcutContext";
import { ComposePage } from "./pages/ComposePage";
import { SnippetsPage } from "./pages/SnippetsPage";
import { ReelsPage } from "./pages/ReelsPage";
import { NoCodePage } from "./pages/NoCodePage";
import { CoursesPage } from "./pages/CoursesPage";
import { CoursePage } from "./pages/CoursePage";
import { CourseEditorPage } from "./pages/CourseEditorPage";
import { LessonPage } from "./pages/LessonPage";
import { LessonEditorPage } from "./pages/LessonEditorPage";
import { CoursePreview } from "./pages/CoursePreview";
import { QuizPage } from "./pages/QuizPage";
import { QuizStrategyPage } from "./pages/QuizStrategyPage";
import { GenerationPage } from "./pages/GenerationPage";
import { ResultsGalleryPage } from "./pages/ResultsGalleryPage";
import { TemplatesPage } from "./pages/TemplatesPage";
import { AdaptiveConfigPage } from "./pages/AdaptiveConfigPage";
import { LibraryPage } from "./pages/LibraryPage";
import { LedgerPage } from "./pages/LedgerPage";
import { ShowcasePage } from "./pages/ShowcasePage";

function NotFound() {
  return (
    <div className="p-6 text-muted">
      Not found.{" "}
      <Link className="text-brand hover:underline" to="/">
        Back to compose
      </Link>
    </div>
  );
}

export function App() {
  const { overlayOpen, setOverlayOpen } = useShortcutContext();
  const [logOpen, setLogOpen] = useState(false);
  const toggleLog = useCallback(() => setLogOpen((v) => !v), []);

  useShellShortcuts({ overlayOpen, setOverlayOpen, toggleLog });

  return (
    <StudioLayout logOpen={logOpen} onCloseLog={() => setLogOpen(false)}>
      <Routes>
        <Route path="/" element={<ComposePage />} />
        <Route path="/snippets" element={<SnippetsPage />} />
        <Route path="/reels" element={<ReelsPage />} />
        <Route path="/nocode" element={<NoCodePage />} />
        <Route path="/courses" element={<CoursesPage />} />
        <Route path="/generation" element={<GenerationPage />} />
        <Route path="/templates" element={<TemplatesPage />} />
        <Route path="/adaptive" element={<AdaptiveConfigPage />} />
        <Route path="/library" element={<LibraryPage />} />
        <Route path="/c/:slug" element={<CoursePage />} />
        <Route path="/c/:slug/edit" element={<CourseEditorPage />} />
        <Route path="/c/:slug/l/:id" element={<LessonPage />} />
        <Route path="/c/:slug/l/:id/edit" element={<LessonEditorPage />} />
        <Route path="/c/:slug/l/:id/preview" element={<CoursePreview />} />
        <Route path="/c/:slug/l/:id/quiz" element={<QuizPage />} />
        <Route path="/c/:slug/l/:id/strategy" element={<QuizStrategyPage />} />
        <Route path="/c/:slug/l/:id/results" element={<ResultsGalleryPage />} />
        <Route path="/ledger" element={<LedgerPage />} />
        <Route path="/showcase" element={<ShowcasePage />} />
        <Route path="*" element={<NotFound />} />
      </Routes>
    </StudioLayout>
  );
}
