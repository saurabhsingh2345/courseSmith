import { useEffect, useState } from "react";
import { Link, NavLink, Route, Routes } from "react-router-dom";
import { isTypingTarget, useShortcutContext } from "./state/ShortcutContext";
import { RunBar } from "./components/RunBar";
import { LogPanel } from "./components/LogPanel";
import { ShortcutOverlay } from "./components/ShortcutOverlay";
import { ComposePage } from "./pages/ComposePage";
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

function NavItem({ to, label }: { to: string; label: string }) {
  return (
    <NavLink
      to={to}
      end
      className={({ isActive }) =>
        isActive ? "text-ink-100" : "text-ink-400 hover:text-ink-200"
      }
    >
      {label}
    </NavLink>
  );
}

export function App() {
  const { hints, overlayOpen, setOverlayOpen } = useShortcutContext();
  const [logOpen, setLogOpen] = useState(false);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      if (e.key === "Escape") {
        if (overlayOpen) setOverlayOpen(false);
        return;
      }
      if (isTypingTarget(e.target)) return;
      if (e.key === "?") {
        e.preventDefault();
        setOverlayOpen(!overlayOpen);
      } else if (e.key === "l" || e.key === "L") {
        setLogOpen((v) => !v);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [overlayOpen, setOverlayOpen]);

  const footerHints = [...hints, { keys: "L", label: "logs" }, { keys: "?", label: "shortcuts" }];

  return (
    <div className="flex h-full flex-col">
      <header className="flex items-center gap-4 border-b border-ink-800 bg-ink-900 px-4 py-2">
        <Link to="/" className="font-semibold text-ink-100">
          coursesmith <span className="text-ink-500">studio</span>
        </Link>
        <nav className="flex gap-3">
          <NavItem to="/" label="Compose" />
          <NavItem to="/courses" label="Courses" />
          <NavItem to="/generation" label="Generation" />
          <NavItem to="/templates" label="Templates" />
          <NavItem to="/adaptive" label="Adaptive" />
          <NavItem to="/library" label="Library" />
          <NavItem to="/ledger" label="Ledger" />
          <NavItem to="/showcase" label="Showcase" />
        </nav>
        <div className="ml-auto">
          <RunBar />
        </div>
      </header>

      <main className="min-h-0 flex-1 overflow-auto">
        <Routes>
          <Route path="/" element={<ComposePage />} />
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
          <Route
            path="*"
            element={
              <div className="p-6 text-ink-400">
                Not found.{" "}
                <Link className="text-sky-400 hover:underline" to="/">
                  Back to courses
                </Link>
              </div>
            }
          />
        </Routes>
      </main>

      <LogPanel open={logOpen} onClose={() => setLogOpen(false)} />

      <footer className="flex flex-wrap items-center gap-3 border-t border-ink-800 bg-ink-900 px-4 py-1.5 text-[11px] text-ink-500">
        {footerHints.map((h) => (
          <span key={h.keys + h.label} className="flex items-center gap-1">
            <kbd>{h.keys}</kbd> {h.label}
          </span>
        ))}
      </footer>

      {overlayOpen && <ShortcutOverlay hints={hints} onClose={() => setOverlayOpen(false)} />}
    </div>
  );
}
