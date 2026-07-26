import { useEffect, useState } from "react";
import { AbsoluteFill } from "remotion";
import { Player } from "@remotion/player";
import { PythonExecutionViz } from "../../../renderer/src/components/PythonExecutionViz";
import type { CodeTrace } from "../../../renderer/src/types";
import { resolveTheme } from "../../../renderer/src/theme/theme";

// CodeTracePlayer drives the *real* renderer PythonExecutionViz (workstream C)
// over a lesson's actual code-trace artifact — the same component that renders
// the final video, played live at 30fps with scrub controls. It fetches the
// CodeTrace JSON from the artifact URL, then hands it to the Player.
//
// Default export so CoursePreview can React.lazy() it: this pulls in
// @remotion/player, so we keep it off the page's initial bundle (and out of the
// CoursePreview unit test's module graph unless the Code tab is opened).

const THEME = resolveTheme({
  primary: "#306998",
  accent: "#ffd43b",
  background: "#ffffff",
  courseName: "Coursesmith",
});

const FPS = 30;

export default function CodeTracePlayer({ url, title }: { url: string; title?: string }) {
  const [trace, setTrace] = useState<CodeTrace | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let alive = true;
    setTrace(null);
    setError(null);
    fetch(url)
      .then((r) => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return r.json();
      })
      .then((t: CodeTrace) => {
        if (alive) setTrace(t);
      })
      .catch((e) => {
        if (alive) setError(e instanceof Error ? e.message : String(e));
      });
    return () => {
      alive = false;
    };
  }, [url]);

  if (error) {
    return <div className="text-[12px] text-ink-500">Trace unavailable: {error}</div>;
  }
  if (!trace) {
    return <div className="text-[12px] text-ink-500">Loading trace…</div>;
  }

  const steps = Math.max(1, trace.steps?.length ?? 1);
  // One second per step so each line change is easy to follow while scrubbing.
  const durationInFrames = steps * FPS;

  const Composition: React.FC = () => (
    <AbsoluteFill style={{ backgroundColor: "#f6f7f9" }}>
      <PythonExecutionViz
        theme={THEME}
        durationInFrames={durationInFrames}
        props={{ title: title ?? "", trace }}
      />
    </AbsoluteFill>
  );

  return (
    <div className="overflow-hidden rounded-lg border border-ink-700">
      <Player
        component={Composition}
        durationInFrames={durationInFrames}
        compositionWidth={1920}
        compositionHeight={1080}
        fps={FPS}
        style={{ width: "100%" }}
        controls
        loop
      />
    </div>
  );
}
