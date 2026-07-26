import { AbsoluteFill } from "remotion";
import { Player } from "@remotion/player";
import { PythonExecutionViz } from "../../../renderer/src/components/PythonExecutionViz";
import { D3Diagram } from "../../../renderer/src/components/D3Diagram";
import { resolveTheme } from "../../../renderer/src/theme/theme";
import execTrace from "../../../renderer/src/fixtures/execTrace.json";

// ShowcasePlayer embeds the real renderer components (workstreams C and A) in a
// @remotion/player so they actually play in the studio — the same code that
// renders the final video, driven live at 30fps with scrub controls.

const THEME = resolveTheme({
  primary: "#306998",
  accent: "#ffd43b",
  background: "#ffffff",
  courseName: "Coursesmith",
});

const FPS = 30;

// Each demo is wrapped in a tiny composition component: the renderer components
// expect to live inside a Remotion timeline (they call useCurrentFrame), which
// the Player provides.

const ExecVizComposition: React.FC = () => (
  <AbsoluteFill style={{ backgroundColor: "#f6f7f9" }}>
    <PythonExecutionViz
      theme={THEME}
      durationInFrames={FPS * 8}
      props={{ title: "Watching the code run", trace: execTrace }}
    />
  </AbsoluteFill>
);

const D3VizComposition: React.FC = () => (
  <AbsoluteFill style={{ backgroundColor: THEME.background }}>
    {/* D3Diagram fetches its spec from the app's public dir at runtime. */}
    <D3Diagram theme={THEME} props={{ src: "d3demo.json", kind: "d3", title: "" }} />
  </AbsoluteFill>
);

const DEMOS = {
  exec: { component: ExecVizComposition, durationInFrames: FPS * 8, label: "ExecViz — Python execution (workstream C)" },
  d3: { component: D3VizComposition, durationInFrames: FPS * 6, label: "D3Viz — animated diagram (workstream A)" },
} as const;

export function ShowcasePlayer({ demo }: { demo: keyof typeof DEMOS }) {
  const d = DEMOS[demo];
  return (
    <div className="overflow-hidden rounded-lg border border-ink-700">
      <Player
        component={d.component}
        durationInFrames={d.durationInFrames}
        compositionWidth={1920}
        compositionHeight={1080}
        fps={FPS}
        style={{ width: "100%" }}
        controls
        loop
        autoPlay
      />
      <div className="text-ink-500 border-ink-800 border-t px-3 py-1.5 text-[10px]">{d.label}</div>
    </div>
  );
}
