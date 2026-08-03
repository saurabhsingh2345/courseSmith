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
  // The footage template's own contribution — chrome, origin, capture credit —
  // around a neutral placeholder. It is the one scene whose real content is a
  // recording, so the fixture supplies a deterministic frame rather than a clip.
  { id: "FootageViz", frame: 90 },
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
  // The mockup on its last build beat: every block landed, the active one
  // outlined, its layer row lit and its note up. It is also the frame that
  // proves the fit — five blocks are taller than the viewport, so a regression
  // in the scaling shows here as a footer that has fallen off the page.
  { id: "MockupViz", frame: 800 },
  // The stack on its third tier: two bands walked and lit, the current one
  // raised, the one below still dim. It is the only state that shows reached,
  // current and unreached at once, which is the whole grammar of the scene.
  { id: "StackViz", frame: 430 },
  // The spec after its cascade, with one line missed — the only state holding a
  // filled tick, a crossed box, struck text and a tally short of the total at
  // once. A clean sweep would prove none of the miss path.
  { id: "SpecViz", frame: 910 },
  // The showcase on its limits beat — the enforced half lit, all four decision
  // cells filled, the identity block intact. The hand-off frame is the other
  // state worth watching, but it hides the card the template is actually about.
  { id: "ShowcaseViz", frame: 940 },
  // The metric template on its third figure, past the count-up and with the
  // note up — the only state holding a settled number, its unit, its label and
  // its argument at once. It is also the catalog's one skinned baseline, so it
  // is the guard on the broadcast chrome, the semantic accents and the stage
  // scale as much as on the template.
  { id: "MetricViz", frame: 600 },
  // The same template on its recap: every figure back at once in its own role
  // colour. The row assembles on a stagger no other frame catches.
  { id: "MetricViz", frame: 1020 },
  // The gauge mid-clip, on the bar that does not fit: the dashed ceiling across
  // the track, one bar settled and dimmed above, the current one overrunning the
  // rule in the limit colour, and one still unrun below. It is the only state
  // that holds fits, does-not-fit and not-yet at once.
  { id: "GaugeViz", frame: 600 },
  // The occupancy grid with both bands settled: sixteen cells in the quantity
  // colour, a hundred and twenty neutral behind them, and the remaining seven
  // hundred and sixty unclaimed. It is the only state that holds all three at
  // once, which is the whole claim the template makes.
  { id: "OccupancyViz", frame: 660 },
  // The ranking board mid-re-sort on the second arrival: the new row travelling
  // in from the right, three rows sliding down under it, and the bottom row on
  // its way off. It is the only state that holds arriving, moving and leaving at
  // once, which is the whole reason the rows travel rather than being redrawn.
  { id: "RankingViz", frame: 540 },
  // The journal on the DELETE replaying: four lines written, the fifth still a
  // dash, the cursor bar on line four in the limit colour and the lines it has
  // passed dimmed back. It is the only state holding written, unwritten, passed
  // and current at once.
  { id: "JournalViz", frame: 1140 },
  // The multiplex on its wide pass: three chips lit as ready, three wires into
  // the worker, and the count reading 3 beside a box marked one thread. It is
  // the only state that makes the template's actual claim — a single-ready frame
  // draws the same picture as polling.
  { id: "MultiplexViz", frame: 560 },
  // The fork after its one write: the copied page lit beneath the row on the
  // parent's side, and five pages still shared at full strength beside it. Both
  // halves at once is the only state that says what copy-on-write is.
  { id: "ForkViz", frame: 420 },
  // The boundary after its one grant: files across the line in the quantity
  // colour, three capabilities still outside it carrying their crosses, and the
  // count reading one of four. Both halves of the rule in one frame.
  { id: "CapabilitiesViz", frame: 480 },
  // The budget on its closing beat: three segments filling most of the bar, the
  // remainder as the gap that is left, and the figure counted out under it. It is
  // the frame the whole template exists for.
  { id: "BudgetViz", frame: 1020 },
  // The latency axis with all three placed: a sliver at 0.1ms, a bar to 12ms,
  // and one running almost the full width to 6.5s, over five named decades. It is
  // the only state where the categorical gap the template exists for is visible.
  { id: "LatencyViz", frame: 960 },
  // The multiply on its caveat beat: the per-unit figure still where it was set,
  // eight glyphs, the product beneath it, and the caveat chip. It is the only
  // state that holds the whole sentence at once.
  { id: "MultiplyViz", frame: 750 },
  // And mid-bite on the third claim, which is the only state that shows a segment
  // growing into a bar the earlier segments are not re-scaling inside.
  { id: "BudgetViz", frame: 660 },
  // Mid-append, so the baseline also covers the first half — a line arriving at
  // the end of a file whose later slots are still empty.
  { id: "JournalViz", frame: 600 },
  // And settled, so the baseline also covers the resting board the clip ends on.
  { id: "RankingViz", frame: 840 },
  // Mid-sweep on the second band, to prove the cells light in a run rather than
  // snapping on together — the gesture is what makes a block read as a quantity.
  { id: "OccupancyViz", frame: 518 },
  // The verdict on its first asterisk beat: the holds column receded as a
  // block, one break lit on its tinted plate, the other still muted. It is the
  // only state that proves the asymmetry the template is built on.
  // Mid-strike: the rule part-way across the claim, the claim greying out
  // under it, the truth not yet up. It is the one frame that proves the
  // gesture travels rather than being toggled on, and it exists for about
  // half a second.
  // The rundown on its second card: one lit, two dimmed but still legible,
  // and the detail line under the row. It is the only state that proves the
  // row is fixed furniture and only brightness moves.
  { id: "RundownViz", frame: 600 },
  // The analogy on its second correspondence: both columns complete, one row
  // lit end to end with its connector drawn, the others dimmed, and the note
  // under it. The only state that holds the mapping and the walk at once.
  // The trace on its last operation: three drained and struck above, the
  // fourth lit, one marked "no change", and the value already at zero. It is
  // the only frame that holds history, contention and the bug at once.
  { id: "TraceViz", frame: 1080 },
  // The costing on its first hidden line: two ordinary costs settled above, the
  // surprising one lit with its badge, one still unlanded, and the running
  // total part-way up. The only state holding all four line states at once.
  { id: "CostingViz", frame: 750 },
  // The constellation closing on the whole picture: every spoke lit, every
  // relation word up, nothing moved since the walk. The frame the template
  // exists to produce.
  { id: "ConstellationViz", frame: 1090 },
  // And mid-walk, where two spokes are lit, one is drawing and one is still
  // faint — the state that proves the map accumulates rather than reveals.
  { id: "ConstellationViz", frame: 660 },
  // The chapter break on its handover beat: two stops ticked, the current one
  // haloed, two ahead still faint, and the ordinal at full stroke. It is the
  // only state that holds all three stop treatments and the hand-off line at
  // once.
  { id: "ChapterViz", frame: 750 },
  // And a look-back beat, where the card names a section already behind the
  // viewer rather than the one starting — the half of the template a frame on
  // the handover cannot show.
  { id: "ChapterViz", frame: 540 },
  // The same break in light mode. Every one of these three leans on
  // accentQuantity for its entire structure, so light mode is where a fixture
  // that forgot the semantic accents would show up.
  { id: "ChapterLightViz", frame: 750 },
  // The cycle closing on the return: the ring complete, every stage lit, and
  // the line about what changes next lap in the hub. The frame the template
  // exists to produce — and the one where a full-circle arc has to render,
  // which a single SVG arc command cannot do.
  { id: "CycleViz", frame: 1740 },
  // And mid-walk, with the comet parked at a stage's edge, two arcs lit and two
  // stages still dark — the state that proves the ring accumulates.
  { id: "CycleViz", frame: 1080 },
  { id: "CycleLightViz", frame: 1740 },
  // The scale on its last rung: three worlds nested inside the current one, the
  // viewfinder brackets on the subject, and the step multiplier beside the
  // figure.
  { id: "ScaleViz", frame: 1110 },
  // And the closing frame, where the ladder compresses so all four worlds are
  // legible at once. Nothing else in the clip renders that geometry.
  { id: "ScaleViz", frame: 1440 },
  { id: "ScaleLightViz", frame: 1110 },
  { id: "AnalogyViz", frame: 600 },
  // The break: the whole mapping receded and the admission over it. Nothing
  // else in the catalog dims its own content to make a point.
  { id: "AnalogyViz", frame: 1080 },
  { id: "MythViz", frame: 225 },
  // Two beats later: the claim struck and settled, the truth up, and one piece
  // of evidence carded under it.
  { id: "MythViz", frame: 500 },
  { id: "DecisionViz", frame: 600 },
  // The closing rule: every band lit at once with its own answer carded under
  // it, which is the only frame that proves the partition is total.
  { id: "DecisionViz", frame: 1080 },
  { id: "VerdictViz", frame: 810 },
  // The closing frame: the columns gone and the ruling alone at headline size.
  // Nothing else in the catalog renders this, so it regresses here or nowhere.
  { id: "VerdictViz", frame: 1150 },
  // The breakdown on an item beat: two phases collapsed above, one open with an
  // item spotlit and its neighbours dimmed, one unreached below, and the
  // progress read under it. Every state the accordion has, in one frame.
  { id: "BreakdownViz", frame: 700 },
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
