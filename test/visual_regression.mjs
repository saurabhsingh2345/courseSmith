// Visual regression for Remotion compositions (workstream I).
//
// Renders each composition to a PNG at a fixed frame and pixel-diffs it against
// a committed baseline under test/baselines/. If more than THRESHOLD of the
// pixels differ the check fails and a *.diff.png is written for review.
//
// Usage:
//   node test/visual_regression.mjs            compare against baselines
//   node test/visual_regression.mjs --update   (re)write baselines
//
// The render path (bundle -> selectComposition -> renderStill) is the same one
// proven for ad-hoc still renders. pixelmatch + pngjs are dev-deps of renderer/
// so they resolve from there; if any of the render deps are missing (e.g. a
// CI runner without a browser) the check skips cleanly with exit 0 rather than
// failing, so it can be wired into CI before every runner can render.

import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const RENDERER_DIR = join(__dirname, "..", "renderer");
const BASELINE_DIR = join(__dirname, "baselines");
const NM = (...p) => join(RENDERER_DIR, "node_modules", ...p);

// Compositions to snapshot and the frame to capture each at. Frames are chosen
// to land on a settled, content-rich moment (not mid-transition) so the diff is
// stable across runs.
const TARGETS = [
  { id: "ExecViz", frame: 60 },
  { id: "PointsViz", frame: 140 },
  { id: "VSCodeViz", frame: 200 },
  // The same editor in light mode, with the terminal drawer open. The editor
  // carries its own palette rather than the design system's, so it is the one
  // scene where light mode cannot be inferred from any other baseline — and
  // the drawer is the half of the template that had no coverage at all.
  { id: "VSCodeLightViz", frame: 460 },
  // The opening gesture: the pointer on the file it is about to click, the row
  // lit under it, the tab not yet open. Every snippet renders this and no
  // baseline covered any of it — the other two fixtures leave `intro` off, so
  // the window scale-up, the tree click and the tab slide were all switched
  // off in the only compositions that existed.
  { id: "VSCodeIntroViz", frame: 28 },
  // The quiz template, on an `explain` beat — the only state where every part
  // of the scene is on screen at once: the question, the revealed answer, a
  // dimmed distractor lifted back up, and its explanation underneath.
  { id: "QuizViz", frame: 560 },
  // The compare template on its verdict beat — the only state where the winner
  // is marked, the loser has receded and the verdict line is up, so a
  // regression in any of the three shows in one frame.
  { id: "CompareViz", frame: 570 },
  // The anatomy template on a part beat: the lit run, the dimmed remainder, the
  // callout and its note — the only state where all four are on screen.
  { id: "AnatomyViz", frame: 420 },
  // The timeline mid-walk. The stops still ahead are visible but faded, which
  // is the state that proves the future is drawn rather than revealed — a
  // frame on the closing beat could not tell the two designs apart.
  { id: "TimelineViz", frame: 350 },
  // The canvas mid-run: the token out on the wire between two cards, ticks
  // behind it, the wire ahead still dark. A frame on a build beat would prove
  // the layout and nothing about the payoff, which is the half of this template
  // that has moving parts.
  { id: "CanvasViz", frame: 755 },
  // The prompt loop on its second answer — the only state holding the whole
  // argument at once: three turns of history in the thread, the attempt counter
  // past one, and the goal bar reaching further than it did last round.
  { id: "PromptLoopViz", frame: 825 },
  // Mid-word, with the completion popup open. It only exists for the handful
  // of frames where a fragment has matches, so no other target can catch it —
  // and it is the detail that makes the scene read as an editor rather than as
  // text appearing.
  { id: "VSCodeViz", frame: 52 },
  // The workspace template's payoff: the code and the terminal framed
  // together, with the output the program really printed. It is the one shot
  // that exercises the camera, the multi-file tabs and the executed project at
  // once.
  { id: "WorkspaceViz", frame: 690 },
  // Late enough that every box is drawn and settled, so the frame catches the
  // finished board rather than a mid-stroke moment.
  { id: "WhiteboardViz", frame: 390 },
  // Inside the focus window, so the frame covers the dim/highlight path as
  // well as the graph itself.
  { id: "FlowViz", frame: 400 },
  // The second shot, past its headline and caption but with the figure's idle
  // still running — so the frame covers the alternated layout, the marker
  // stroke and an assembled figure at once.
  { id: "IllustrationViz", frame: 250 },
  // The same shot in light mode, with captions on. Light mode is the branch no
  // default config exercises, so without a baseline every scene that quietly
  // assumed a dark stage regresses silently.
  { id: "IllustrationLightViz", frame: 75 },
  // Mid-clip, so the frame catches a settled pose with its prop up rather than
  // the entrance.
  { id: "CastViz", frame: 200 },
  // Deep into the third shot, so the camera has travelled most of a push and
  // the frame proves the world-space staging composes under a moved camera —
  // which a frame-0 capture cannot.
  { id: "StoryViz", frame: 340 },
  // Inside the map shot's highlight window: the kind with an external
  // dependency (the world atlas), whose failure mode is a country that quietly
  // does not appear. It is the last of the thirteen kinds the demo walks, six
  // seconds each — so the frame moves whenever a kind is added ahead of it.
  { id: "DataViz", frame: 2250 },
  { id: "D3Viz", frame: 130 },
  { id: "MemoryViz", frame: 90 },
  { id: "LessonVideo", frame: 15 },
];

// Fraction of pixels allowed to differ before a target fails. Small non-zero
// budget absorbs sub-pixel antialiasing jitter between headless renders.
const THRESHOLD = 0.01;
// Per-pixel colour-distance sensitivity handed to pixelmatch (0 strict, 1 lax).
const PIXEL_THRESHOLD = 0.1;

async function importOrSkip() {
  try {
    const bundler = await import(NM("@remotion", "bundler", "dist", "index.js"));
    const renderer = await import(NM("@remotion", "renderer", "dist", "index.js"));
    const pixelmatch = (await import(NM("pixelmatch", "index.js"))).default;
    const { PNG } = await import(NM("pngjs", "lib", "png.js"));
    return { bundler, renderer, pixelmatch, PNG };
  } catch (err) {
    console.error(`SKIP visual_regression: render deps not resolvable (${err.message})`);
    process.exit(0); // graceful skip — not a failure where the deps/browser are absent
  }
}

// compare returns { ok, reason } for one already-rendered pair of PNG buffers.
function compare(pixelmatch, PNG, baselineBuf, actualBuf, diffPath) {
  const base = PNG.sync.read(baselineBuf);
  const actual = PNG.sync.read(actualBuf);
  if (base.width !== actual.width || base.height !== actual.height) {
    return {
      ok: false,
      reason: `dimensions ${actual.width}x${actual.height} != baseline ${base.width}x${base.height}`,
    };
  }
  const { width, height } = base;
  const diff = new PNG({ width, height });
  const changed = pixelmatch(base.data, actual.data, diff.data, width, height, {
    threshold: PIXEL_THRESHOLD,
  });
  const total = width * height;
  const ratio = changed / total;
  if (ratio > THRESHOLD) {
    writeFileSync(diffPath, PNG.sync.write(diff));
    return {
      ok: false,
      reason: `${changed}/${total} px differ (${(ratio * 100).toFixed(2)}% > ${(THRESHOLD * 100).toFixed(1)}%) — see ${diffPath}`,
    };
  }
  return { ok: true, reason: `${changed}/${total} px differ (${(ratio * 100).toFixed(2)}%)` };
}

async function main() {
  const update = process.argv.includes("--update");
  const { bundler, renderer, pixelmatch, PNG } = await importOrSkip();

  mkdirSync(BASELINE_DIR, { recursive: true });
  await renderer.ensureBrowser();
  const serveUrl = await bundler.bundle({ entryPoint: join(RENDERER_DIR, "src", "index.ts") });

  let failures = 0;
  for (const { id, frame } of TARGETS) {
    const stem = `${id}-${frame}`;
    const baseline = join(BASELINE_DIR, `${stem}.png`);
    const actual = join(BASELINE_DIR, `${stem}.actual.png`);
    const diffPath = join(BASELINE_DIR, `${stem}.diff.png`);

    const composition = await renderer.selectComposition({ serveUrl, id });
    await renderer.renderStill({ serveUrl, composition, frame, output: update ? baseline : actual });

    if (update) {
      console.log(`baseline written: ${stem}.png`);
      continue;
    }
    if (!existsSync(baseline)) {
      console.error(`MISSING baseline for ${stem} — run with --update`);
      failures++;
      continue;
    }
    const { ok, reason } = compare(pixelmatch, PNG, readFileSync(baseline), readFileSync(actual), diffPath);
    if (ok) {
      console.log(`PASS ${stem}: ${reason}`);
    } else {
      console.error(`FAIL ${stem}: ${reason}`);
      failures++;
    }
  }

  if (!update && failures > 0) {
    console.error(`\nvisual_regression: ${failures} target(s) failed`);
  }
  process.exit(failures > 0 ? 1 : 0);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
