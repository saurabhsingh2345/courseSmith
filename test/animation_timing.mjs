// Animation timing checks (workstream I).
//
// Guards that a composition's reveal animation actually plays and lands where
// the motion tokens predict — the concern being both "the animation never
// runs / is baked static" and "the deterministic layout flickers" (see the
// D3Diagram design notes). It renders a handful of frames of D3Viz, measures an
// "ink" signal per frame (pixels that differ from the background), and asserts:
//
//   1. the reveal builds up monotonically (no backward flicker, within a small
//      overshoot tolerance),
//   2. it is still visibly incomplete early on (the animation isn't instant),
//   3. it has settled by the frame the motion tokens predict, and
//   4. the settled frame matches the fully-built final frame.
//
// The predicted settle frame is derived from the SAME tokens the renderer uses
// (renderer/src/theme/motion.defaults.json) and the D3 fixture's node/edge
// counts, so a change to the tokens or the reveal math moves the expectation
// with it.
//
// == Why the reveal is measured with the camera parked ==
//
// Those four checks are about the REVEAL, and every one of them reads total ink.
// That was a complete description of the frame until `d64f816` gave every scene
// a camera: the frame now creeps closer for the whole of its first eighteen
// seconds, so content keeps growing long after the last node has landed and the
// ink signal keeps climbing with it. Check 4 — "nothing more appears after the
// reveal settles" — went red on a scene nobody had touched, and it was right
// about the pixels and wrong about the cause.
//
// Measured on D3Viz between the settle frame and eighty frames later: with the
// camera running, ink grows 3.35% and 4% of pixels move; with it parked, ink
// moves by 0.01% and *zero* pixels differ. The reveal settles exactly as the
// tokens predict. So the reveal is measured against a parked camera, which is
// the only way these four checks mean what they say.
//
//   5. and then, separately, the camera is asserted to be ALIVE — the same two
//      frames rendered with it running must differ from the parked ones.
//
// That check is not padding. The camera's own commit message records that its
// first implementation produced no motion at all (3% spread over a seventy-second
// segment is 28 pixels of travel), and nothing in the suite would have caught it.
// Parking the camera to fix check 4 without asserting it still runs would remove
// the only frames in the whole suite that would notice.
//
// Usage: node test/animation_timing.mjs
// Skips cleanly (exit 0) where a browser or the render deps aren't available.

import { mkdtempSync, readFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const RENDERER_DIR = join(__dirname, "..", "renderer");
const NM = (...p) => join(RENDERER_DIR, "node_modules", ...p);

const FPS = 30;
const MOTION = JSON.parse(readFileSync(join(RENDERER_DIR, "src", "theme", "motion.defaults.json"), "utf8"));
const FIXTURE = JSON.parse(readFileSync(join(RENDERER_DIR, "public", "d3demo.json"), "utf8"));

// secondsToFrames mirrors renderer/src/theme/motion.ts: max(1, round(s*fps)).
const s2f = (seconds) => Math.max(1, Math.round(seconds * FPS));

// Reveal math mirrors D3Diagram.tsx: node `i` reveals over `revealDur` starting
// at `i*nodeStagger`; edge `i` starts at `i*edgeStagger + nodeStagger`. The
// per-item `order` is bounded by count-1, so this is the settle upper bound.
const nodeStagger = s2f(MOTION.stagger.items);
const edgeStagger = s2f(MOTION.stagger.connections);
const revealDur = s2f(MOTION.timing.normal);
const nodeCount = FIXTURE.nodes?.length ?? 0;
const edgeCount = FIXTURE.edges?.length ?? 0;
const lastNodeDone = Math.max(0, nodeCount - 1) * nodeStagger + revealDur;
const lastEdgeDone = Math.max(0, edgeCount - 1) * edgeStagger + nodeStagger + revealDur;
const settleFrame = Math.max(lastNodeDone, lastEdgeDone);

const FRAME_TOLERANCE = 5; // allow the settle to complete a few frames late
const INK_TOLERANCE = 0.03; // 3% of final ink absorbs entrance-overshoot wobble

// Parks the camera for the reveal measurements. `resolveMotion` merges this over
// the defaults, and SceneCamera renders its children untouched — not wrapped in
// an identity transform — when both values are zero, so a parked frame is the
// scene exactly as it would have been drawn before the camera existed.
const PARKED_CAMERA = { motion: { camera: { push: 0, drift: 0 } } };

// How much of the frame the camera must have displaced by the late sample for it
// to count as running. Measured at 3.7%; a token change that halved the push
// would still clear this, and one that zeroed it could not.
const CAMERA_MOTION_FLOOR = 0.01;

async function importOrSkip() {
  try {
    const bundler = await import(NM("@remotion", "bundler", "dist", "index.js"));
    const renderer = await import(NM("@remotion", "renderer", "dist", "index.js"));
    const { PNG } = await import(NM("pngjs", "lib", "png.js"));
    const pixelmatch = (await import(NM("pixelmatch", "index.js"))).default;
    return { bundler, renderer, PNG, pixelmatch };
  } catch (err) {
    console.error(`SKIP animation_timing: render deps not resolvable (${err.message})`);
    process.exit(0);
  }
}

// ink counts pixels that differ from the background, sampled from the
// top-left corner. The scene background is no longer uniform (gradient +
// slowly drifting accent glows + grain), so the threshold is set well above
// that ambient variation — only genuine content (text, node fills, strokes)
// clears it, keeping the signal stable while the background breathes.
function ink(PNG, buf) {
  const { data, width, height } = PNG.sync.read(buf);
  const br = data[0], bg = data[1], bb = data[2];
  let count = 0;
  for (let i = 0; i < data.length; i += 4) {
    const d = Math.abs(data[i] - br) + Math.abs(data[i + 1] - bg) + Math.abs(data[i + 2] - bb);
    if (d > 110) count++;
  }
  return count / (width * height);
}

async function main() {
  const { bundler, renderer, PNG, pixelmatch } = await importOrSkip();
  await renderer.ensureBrowser();
  const serveUrl = await bundler.bundle({ entryPoint: join(RENDERER_DIR, "src", "index.ts") });
  // Two composition handles, because input props are resolved at selection time
  // as well as at render time — passing them to renderStill alone leaves the
  // camera running and the "parked" frames come back byte-identical to the live
  // ones, which reads as the camera doing nothing rather than as the override
  // being dropped.
  const parked = await renderer.selectComposition({ serveUrl, id: "D3Viz", inputProps: PARKED_CAMERA });
  const live = await renderer.selectComposition({ serveUrl, id: "D3Viz", inputProps: {} });

  const outDir = mkdtempSync(join(tmpdir(), "anim-timing-"));
  const shoot = async (frame, camera) => {
    const output = join(outDir, `f${frame}-${camera ? "live" : "parked"}.png`);
    await renderer.renderStill({
      serveUrl,
      composition: camera ? live : parked,
      frame,
      output,
      inputProps: camera ? {} : PARKED_CAMERA,
    });
    return readFileSync(output);
  };
  const render = async (frame) => ink(PNG, await shoot(frame, false));

  const early = Math.min(5, Math.floor(settleFrame / 3));
  const mid = Math.floor(settleFrame / 2);
  const settle = settleFrame + FRAME_TOLERANCE;
  const final = Math.min(parked.durationInFrames - 1, settle + 80);

  const samples = {
    [early]: await render(early),
    [mid]: await render(mid),
    [settle]: await render(settle),
    [final]: await render(final),
  };
  for (const [f, v] of Object.entries(samples)) {
    console.log(`  frame ${f}: ink ${(v * 100).toFixed(2)}%  (camera parked)`);
  }

  // The same late frame with the camera running, for check 5.
  const liveFinal = PNG.sync.read(await shoot(final, true));
  const parkedFinal = PNG.sync.read(await shoot(final, false));
  const cameraMoved =
    pixelmatch(liveFinal.data, parkedFinal.data, null, liveFinal.width, liveFinal.height, {
      threshold: 0.1,
    }) /
    (liveFinal.width * liveFinal.height);
  console.log(`  frame ${final}: camera displaced ${(cameraMoved * 100).toFixed(2)}% of the frame`);

  const failures = [];
  const tol = INK_TOLERANCE * samples[final];

  // 1. Monotonic build (within overshoot tolerance): early <= mid <= settle.
  if (samples[early] > samples[mid] + tol) failures.push(`ink dropped ${early}->${mid} (flicker): ${samples[early].toFixed(4)} > ${samples[mid].toFixed(4)}`);
  if (samples[mid] > samples[settle] + tol) failures.push(`ink dropped ${mid}->${settle} (flicker): ${samples[mid].toFixed(4)} > ${samples[settle].toFixed(4)}`);

  // 2. Not instant: the reveal is still visibly incomplete at the early frame.
  if (samples[early] >= 0.85 * samples[settle]) failures.push(`animation looks instant: frame ${early} already ${(100 * samples[early] / samples[settle]).toFixed(0)}% of settled ink`);

  // 3. Settled on schedule: near-complete by the token-predicted frame.
  if (samples[settle] < 0.92 * samples[final]) failures.push(`not settled by frame ${settle}: ${(100 * samples[settle] / samples[final]).toFixed(0)}% of final ink (want >= 92%)`);

  // 4. Final frame agrees with settle (nothing more appears after).
  if (Math.abs(samples[settle] - samples[final]) > tol) failures.push(`settle (${samples[settle].toFixed(4)}) and final (${samples[final].toFixed(4)}) disagree by > ${INK_TOLERANCE * 100}%`);

  // 5. The camera is alive. Everything above deliberately renders with it parked,
  //    so without this the suite would pass a build in which it had stopped
  //    working — which is the exact failure its own first implementation had.
  if (cameraMoved < CAMERA_MOTION_FLOOR) failures.push(`the camera is not moving: frame ${final} with it running differs from the parked frame by only ${(cameraMoved * 100).toFixed(2)}% of pixels (want >= ${CAMERA_MOTION_FLOOR * 100}%)`);

  // 6. Sanity: there is actually content.
  if (samples[final] <= 0) failures.push(`final frame ${final} is blank`);

  if (failures.length > 0) {
    console.error(`FAIL animation_timing (D3Viz, predicted settle frame ${settleFrame})`);
    for (const f of failures) console.error(`  - ${f}`);
    process.exit(1);
  }
  console.log(`PASS animation_timing: D3Viz reveal builds and settles by frame ${settle} (predicted ${settleFrame}), and the camera is running`);
  process.exit(0);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
