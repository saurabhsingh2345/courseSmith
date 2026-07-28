// Scene graph schema — mirrors internal/pipeline/scenegraph.go exactly.

import type {MotionTokens} from './theme/motion';

export type Theme = {
  primary: string;
  accent: string;
  background: string;
  courseName: string;
  // Rich design tokens derived by the Go pipeline (videotheme.go). Optional:
  // scene graphs generated before the design system omit them, and
  // resolveTheme() (theme/theme.ts) fills in dark-editorial defaults.
  mode?: 'dark' | 'light';
  bgTop?: string;
  bgBottom?: string;
  surface?: string;
  surfaceBorder?: string;
  text?: string;
  textMuted?: string;
  mass?: string;
  ink?: string;
  accentText?: string;
  fontDisplay?: string;
  fontBody?: string;
  fontMono?: string;
  grain?: number;
};

export type Callout = {
  atMs: number;
  durMs: number;
  shape: 'arrow' | 'circle';
  x: number; // fraction of frame width, 0-1
  y: number; // fraction of frame height, 0-1
  label: string;
};

// Execution trace — mirrors internal/pipeline/trace.go. Embedded in a code
// scene's props (as `trace`) to drive the Python-Tutor execution viz.
export type TraceValue = {
  type: string;
  repr: string;
  items?: TraceValue[];
  entries?: {key: string; value: TraceValue}[];
  fields?: {key: string; value: TraceValue}[];
};
export type TraceVar = {name: string; value: TraceValue};
export type TraceStep = {
  step: number;
  line: number; // 1-based
  event: string;
  func: string;
  vars: TraceVar[];
  stack: string[];
  stdout: string;
};
export type CodeTrace = {
  code: string;
  lines: string[];
  steps: TraceStep[];
  truncated: boolean;
  // Go strips this when empty (omitempty); the raw tracer emits null.
  error?: string | null;
  stdout: string;
};

// D3 node-link diagram spec — mirrors internal/pipeline/visuals_multi.go.
export type D3Node = {id: string; label: string; group?: number};
export type D3Edge = {from: string; to: string; label?: string};
export type D3DiagramSpec = {
  kind: 'd3';
  layout: 'tree' | 'radial' | 'force';
  title?: string;
  nodes: D3Node[];
  edges: D3Edge[];
};

export type SceneType =
  | 'title'
  | 'code'
  | 'diagram'
  | 'terminal'
  | 'points'
  | 'walkthrough'
  | 'workspace'
  | 'whiteboard'
  | 'flow'
  | 'illustration'
  | 'cast'
  | 'story'
  | 'data'
  | 'quiz'
  | 'compare'
  | 'anatomy'
  | 'timeline'
  | 'canvas'
  | 'promptloop'
  | 'mockup';

export type Scene = {
  type: SceneType;
  startMs: number;
  endMs: number;
  props: Record<string, unknown>;
  callouts?: Callout[];
};

export type CaptionWord = {
  word: string;
  startMs: number;
  endMs: number;
};

export type LessonVideoProps = {
  theme: Theme;
  // Animation language for this lesson; absent on scene graphs generated
  // before the motion system — resolveMotion() falls back to defaults.
  motion?: MotionTokens;
  assetBase?: string;
  audioFile: string;
  /** Generated keystroke click track, played under the voice. Optional: only
   *  videos that type something have one. See internal/pipeline/keysound.go. */
  sfxFile?: string;
  durationMs: number;
  scenes: Scene[];
  captions: CaptionWord[];
  // Global caption-word indices the emphasis pass marked as keywords; the
  // caption track keeps them in the accent colour. Absent on older graphs.
  captionEmphasis?: number[];
};

export const FPS = 30;

export const msToFrame = (ms: number): number => Math.round((ms / 1000) * FPS);

/** Resolves a lesson-relative asset path against the staged asset base. */
export const assetPath = (base: string | undefined, rel: string): string =>
  base ? `${base}/${rel}` : rel;
