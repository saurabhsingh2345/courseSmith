// Rebuilds the snippet-template preview thumbnails the studio gallery shows.
//
// The previews are downscaled *real renders* rather than mock-ups, and they
// come from the visual-regression baselines — which means the picture on a
// template's card is, by construction, what that template actually produces.
// A hand-drawn thumbnail would drift from the template the first time anyone
// changed a layout, and nobody would notice because nothing compares them.
//
// Run after the baselines change:
//   node test/visual_regression.mjs --update
//   node test/template_previews.mjs
//
// A Go test (internal/studio) fails if a registered template has no preview,
// so adding a template to the catalog without coming back here is caught.

import { execFileSync } from "node:child_process";
import { existsSync, mkdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const BASELINES = join(__dirname, "baselines");
const OUT = join(__dirname, "..", "studio", "public", "template-previews");

// template name -> the baseline that shows it. Kept explicit rather than
// derived: the composition ids and the template names are deliberately
// different vocabularies and guessing between them would break silently.
const SOURCES = {
  objective: "ObjectiveViz-750",
  prereq: "PrereqViz-700",
  recap: "RecapViz-700",
  pitfall: "PitfallViz-840",
  checkpoint: "CheckpointViz-800",

  // The one preview that cannot show real content, and that is a property of
  // the template rather than a shortcut: what `footage` produces is *your*
  // recording, so a card showing somebody else's clip would misrepresent what
  // you get. This shows what the template actually contributes — the frame,
  // the origin, and the capture credit.
  footage: "FootageViz-90",
  vscode: "VSCodeViz-200",
  quiz: "QuizViz-560",
  compare: "CompareViz-570",
  anatomy: "AnatomyViz-420",
  timeline: "TimelineViz-350",
  canvas: "CanvasViz-755",
  promptloop: "PromptLoopViz-825",
  mockup: "MockupViz-800",
  stack: "StackViz-430",
  spec: "SpecViz-910",
  showcase: "ShowcaseViz-940",
  breakdown: "BreakdownViz-700",
  // The stated-figure frame rather than the recap: the card should show what
  // the template does most of the time, and four beats in five state a number.
  metric: "MetricViz-600",
  // The overrun frame: the card should show the moment the template exists for,
  // which is a bar crossing the line rather than one that comfortably clears it.
  gauge: "GaugeViz-600",
  // Both bands settled rather than mid-sweep: the card should show the finished
  // proportion, which is the thing somebody choosing this template is after.
  occupancy: "OccupancyViz-660",
  // Mid-re-sort rather than settled: a still of a finished board looks like any
  // ordered list, and the card has to show the one thing this template does.
  ranking: "RankingViz-540",
  // The replay frame rather than an append: a card showing a half-filled file
  // looks like any code panel, and the replay cursor is what names the template.
  journal: "JournalViz-1140",
  // The wide pass, for the same reason the baseline picks it: a card showing one
  // ready source advertises polling.
  multiplex: "MultiplexViz-560",
  // After the split, so the card carries the divergence and the pages that did
  // not diverge — a pre-write frame is just a row of cells.
  fork: "ForkViz-420",
  // After the grant, so the card shows the gap rather than a sealed box — a
  // fully denied frame advertises a wall.
  capabilities: "CapabilitiesViz-480",
  // The remainder frame: the card should show the number the template exists to
  // state, not a bar part-way through being eaten.
  budget: "BudgetViz-1020",
  // All three rows placed: a card with one bar on it looks like any chart, and
  // the gap between the rows is the whole subject.
  latency: "LatencyViz-960",
  // The whole sentence rather than a mid-build state: the card has to show the
  // small figure and the product together or it is just a number.
  multiply: "MultiplyViz-750",
  // With the phrase up: two bars alone are a chart, and the words are what makes
  // this template itself.
  ratio: "RatioViz-720",
  // Stripped back rather than straight: an evenly weighted sheet is just a table,
  // and the card has to show the move.
  table: "TableViz-480",
  // With the asterisks up: a switch on its own is a title card, and the
  // qualifiers are what make this a template rather than a graphic.
  toggle: "ToggleViz-660",
  // The two-column state rather than the call: the card should show the frame
  // the clip spends most of its time in, and a bare line of type would not read
  // as a distinct template on a gallery card.
  verdict: "VerdictViz-810",
  // Mid-walk rather than the closing rule: the card should show the marker on a
  // band with that band's instruction under it, which is the template's
  // characteristic frame.
  decision: "DecisionViz-600",
  // The settled frame rather than mid-strike: a card wants the claim cancelled
  // and the replacement legible, which the half-swept frame does not show.
  myth: "MythViz-500",
  rundown: "RundownViz-600",
  analogy: "AnalogyViz-600",
  trace: "TraceViz-1080",
  costing: "CostingViz-750",
  constellation: "ConstellationViz-1090",
  // The handover rather than a look-back: the card should show the state the
  // template is named for, which is the section that opens next.
  chapter: "ChapterViz-750",
  // The closed ring rather than a mid-walk: a card showing three quarters of a
  // circle would advertise the one shape this template exists to avoid.
  cycle: "CycleViz-1740",
  // The compressed ladder rather than a single rung: it is the only frame where
  // the nesting — the whole idea — is visible on a 480px card.
  scale: "ScaleViz-1440",
  workspace: "WorkspaceViz-690",
  whiteboard: "WhiteboardViz-390",
  flow: "FlowViz-400",
  illustration: "IllustrationViz-250",
  // The pair shot rather than the opening: the card should show that this
  // template arranges objects, which is the thing that distinguishes it from
  // every other big-type look in the catalog. Its `open` shot would show a
  // title card and sell it as one.
  spine: "SpineViz-330",
  cast: "CastViz-200",
  story: "StoryViz-340",
  data: "DataViz-2250",
};

// 480x270 is 2x a card at its widest, so it stays crisp on a retina display
// without shipping a megabyte per template.
const WIDTH = 480;
const HEIGHT = 270;

mkdirSync(OUT, { recursive: true });

let made = 0;
for (const [name, baseline] of Object.entries(SOURCES)) {
  const src = join(BASELINES, `${baseline}.png`);
  if (!existsSync(src)) {
    console.error(`MISSING baseline ${baseline}.png for template "${name}" — run visual_regression.mjs --update first`);
    process.exitCode = 1;
    continue;
  }
  execFileSync("ffmpeg", [
    "-y", "-i", src,
    "-vf", `scale=${WIDTH}:${HEIGHT}:flags=lanczos`,
    join(OUT, `${name}.png`),
    "-loglevel", "error",
  ]);
  console.log(`preview written: ${name}.png`);
  made++;
}
console.log(`${made} preview(s) in studio/public/template-previews/`);
