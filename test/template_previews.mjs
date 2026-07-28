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
  workspace: "WorkspaceViz-690",
  whiteboard: "WhiteboardViz-390",
  flow: "FlowViz-400",
  illustration: "IllustrationViz-250",
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
