import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, seat, withAlpha} from '../theme/theme';
import {Stage} from './Stage';

// DocScene holds a file open and lights one section of it at a time.
//
// The mechanic is dimming, and it is the opposite of how every other document
// scene in the catalog works. Those build up — a line arrives, then another, and
// the viewer watches a file being written. This one shows the WHOLE file from the
// first frame and then takes light away from everything except the part being
// discussed. That choice is what lets a viewer keep the shape of the document in
// their head while the narration works through it: they can always see that there
// are five sections and that this is the third, which is exactly the thing a
// scrolling camera destroys.
//
// Three details carry the illusion of an editor.
//
// THE LINE NUMBERS ADD UP. They are computed off the flattened document with a
// blank line between sections, so a section that starts at line 14 starts there
// because the thirteen lines above it exist. Numbers that do not reconcile are the
// tell that turns an editor into a drawing of one.
//
// MARKDOWN IS PICKED OUT BY ITS OWN SYNTAX. Headings, bullets, inline code and bold
// are styled from the characters the author actually wrote — no field for "this is a
// heading", no parser either. A config-file lesson rendered as flat grey text
// teaches nothing about how such a file is written, and the syntax is right there
// in the string.
//
// AND THE GUTTER IS NARROW. Line numbers at full strength compete with the text;
// at a third they read as furniture, which is what they are. Same for the tree: it
// answers "where does this file live" once and then gets out of the way.

type Block = {text: string; lines?: string[]; note?: string};
type TreeItem = {name: string; kind?: 'dir' | 'file' | 'mark'; depth?: number};

type Step = {
  startMs: number;
  endMs: number;
  show: 'open' | 'tree' | 'block' | 'whole';
  at?: number;
};

type Props = {
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: {
    file?: string;
    crumb?: string;
    tree?: TreeItem[];
    blocks?: Block[];
    steps?: Step[];
  };
};

const WIN_W = 1560;
const WIN_H = 830;
const TREE_W = 320;
const CHROME_H = 44;
const TAB_H = 42;
const GUTTER_W = 62;
const BODY_FS = 24;
const LINE_H = 38;

/**
 * Why the camera settles and then STOPS.
 *
 * A continuous scale on a layer with text in it re-rasterizes every glyph every
 * frame at a slightly different subpixel offset, and the eye reads that as the
 * type glittering. Measured on a held card: 2% of the frame's pixels changed
 * between consecutive frames with nothing supposed to be moving, 12,000 of them
 * by more than six levels. A 2% drift nobody consciously notices is not worth a
 * frame that will not sit still.
 *
 * So the move is: settle in, land on EXACTLY 1, and hold there. Everything that
 * makes the frame feel alive after that is content — a line landing, a row
 * lighting, a caret blinking — which is motion that means something anyway.
 * Where a held frame still wants a nudge, it is whole pixels of translate, which
 * resamples nothing.
 */

const LIGHTS = ['#ff5f57', '#febc2e', '#28c840'];

/**
 * Inline markdown, rendered from the characters the author wrote.
 *
 * Backticks and double asterisks only. Deliberately not a markdown parser: the
 * two things that carry meaning in a config file are "this is a literal you type"
 * and "this is the bit that matters", and every further construct would be styling
 * for its own sake in a frame that is already dimming most of itself.
 */
const inline = (text: string, theme: ResolvedTheme, key: string) => {
  const out: React.ReactNode[] = [];
  // Split on both markers at once so a line can carry either.
  const parts = text.split(/(`[^`]+`|\*\*[^*]+\*\*)/g);
  parts.forEach((part, i) => {
    if (part.startsWith('`') && part.endsWith('`') && part.length > 2) {
      out.push(
        <span key={`${key}-${i}`} style={{color: theme.accent}}>
          {part.slice(1, -1)}
        </span>,
      );
    } else if (part.startsWith('**') && part.endsWith('**') && part.length > 4) {
      out.push(
        <span key={`${key}-${i}`} style={{color: theme.panelText, fontWeight: 700}}>
          {part.slice(2, -2)}
        </span>,
      );
    } else if (part) {
      out.push(<span key={`${key}-${i}`}>{part}</span>);
    }
  });
  return out;
};

/** How a line is set, from how it starts. */
const lineStyle = (raw: string, theme: ResolvedTheme): React.CSSProperties => {
  const t = raw.trimStart();
  if (t.startsWith('#')) {
    const depth = t.match(/^#+/)?.[0].length ?? 1;
    return {
      color: theme.panelText,
      fontWeight: 700,
      fontSize: depth === 1 ? BODY_FS + 3 : BODY_FS,
    };
  }
  if (t.startsWith('-') || t.startsWith('*')) return {color: withAlpha(theme.panelText, 0.82)};
  return {color: withAlpha(theme.panelText, 0.74)};
};

export const DocScene = ({theme, sceneStartMs, props}: Props) => {
  const frame = useCurrentFrame();
  const ms = (frame / FPS) * 1000 + sceneStartMs;
  const blocks = props.blocks ?? [];
  const tree = props.tree ?? [];
  const steps = props.steps ?? [];

  const idx = Math.max(
    0,
    steps.findIndex((s) => ms >= s.startMs && ms < s.endMs),
  );
  const step = steps[idx] ?? steps[steps.length - 1];
  const sceneStart = steps[0]?.startMs ?? sceneStartMs;
  const sceneEnd = steps[steps.length - 1]?.endMs ?? sceneStart + 8000;
  const lit = step?.show === 'block' ? (step.at ?? -1) : -1;
  const treeOn = step?.show === 'tree';

  const settle = interpolate(ms, [sceneStart, sceneStart + 900], [0.964, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  // No continuous drift — it made the document glitter. See the note above.
  const scale = settle;

  const arrive = (from: number, dur = 380) =>
    interpolate(ms, [from, from + dur], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'});

  // Line numbers, reconciled across the whole file: every section's lines, plus a
  // blank line between sections, exactly as the file would be written.
  let n = 0;
  const numbered = blocks.map((b) => {
    const lines = [b.text, ...(b.lines ?? [])].map((text) => ({text, no: ++n}));
    n += 1; // the blank line after the section
    return lines;
  });

  const note = lit >= 0 ? blocks[lit]?.note : undefined;
  const showTree = tree.length > 0;
  const docW = WIN_W - (showTree ? TREE_W : 0);

  return (
    <Stage>
      <div
        style={{
          display: 'flex',
          alignItems: 'flex-start',
          gap: 36,
          transform: `scale(${scale})`,
        }}
      >
        <div
          style={{
            width: WIN_W,
            height: WIN_H,
            borderRadius: 16,
            background: theme.panel,
            border: `1px solid ${withAlpha(theme.panelText, 0.1)}`,
            boxShadow: seat(theme, 'lifted'),
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
            opacity: arrive(sceneStart, 500),
          }}
        >
          {/* Window chrome. */}
          <div
            style={{
              height: CHROME_H,
              flexShrink: 0,
              display: 'flex',
              alignItems: 'center',
              padding: '0 18px',
              gap: 9,
              background: withAlpha(theme.panelText, 0.06),
              borderBottom: `1px solid ${withAlpha(theme.panelText, 0.08)}`,
            }}
          >
            {LIGHTS.map((c) => (
              <span key={c} style={{width: 12, height: 12, borderRadius: 6, background: c}} />
            ))}
            <span
              style={{
                flex: 1,
                textAlign: 'center',
                fontFamily: theme.fontMono,
                fontSize: 18,
                color: withAlpha(theme.panelMuted, 0.85),
                marginRight: 45,
              }}
            >
              {props.file}
            </span>
          </div>

          <div style={{flex: 1, display: 'flex', minHeight: 0}}>
            {/* Where the file lives. */}
            {showTree ? (
              <div
                style={{
                  width: TREE_W,
                  flexShrink: 0,
                  borderRight: `1px solid ${withAlpha(theme.panelText, 0.09)}`,
                  padding: '18px 0',
                  fontFamily: theme.fontMono,
                  fontSize: 19,
                  lineHeight: '34px',
                  // The tree is furniture until the beat that is about it.
                  opacity: treeOn ? 1 : 0.5,
                }}
              >
                {tree.map((it, i) => {
                  const marked = it.kind === 'mark';
                  return (
                    <div
                      key={i}
                      style={{
                        display: 'flex',
                        gap: 10,
                        alignItems: 'center',
                        paddingLeft: 20 + (it.depth ?? 0) * 20,
                        paddingRight: 14,
                        background: marked && treeOn ? withAlpha(theme.accent, 0.14) : 'transparent',
                        borderLeft: `2px solid ${marked && treeOn ? theme.accent : 'transparent'}`,
                      }}
                    >
                      <span style={{color: withAlpha(theme.panelMuted, 0.6)}}>
                        {it.kind === 'dir' ? '▸' : '·'}
                      </span>
                      <span
                        style={{
                          color: marked
                            ? theme.accent
                            : it.kind === 'dir'
                              ? withAlpha(theme.panelText, 0.72)
                              : theme.panelMuted,
                        }}
                      >
                        {it.name}
                      </span>
                    </div>
                  );
                })}
              </div>
            ) : null}

            <div style={{flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0}}>
              {/* Tab, then breadcrumb — the two lines that say which file and where in it. */}
              <div
                style={{
                  height: TAB_H,
                  flexShrink: 0,
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  padding: '0 22px',
                  fontFamily: theme.fontMono,
                  fontSize: 18,
                  borderBottom: `1px solid ${withAlpha(theme.panelText, 0.07)}`,
                }}
              >
                <span style={{color: theme.accent}}>✳</span>
                <span style={{color: withAlpha(theme.panelText, 0.9)}}>{props.file}</span>
                <span style={{color: theme.accent, fontSize: 22, lineHeight: '18px'}}>●</span>
                {props.crumb ? (
                  <span style={{color: withAlpha(theme.panelMuted, 0.6), marginLeft: 14}}>
                    ❯ {props.crumb}
                  </span>
                ) : null}
              </div>

              {/* The document. */}
              <div
                style={{
                  flex: 1,
                  padding: '22px 0',
                  fontFamily: theme.fontMono,
                  fontSize: BODY_FS,
                  lineHeight: `${LINE_H}px`,
                  overflow: 'hidden',
                }}
              >
                {numbered.map((lines, bi) => {
                  // Dimming is the mechanic. Nothing is hidden — the shape of the
                  // file stays readable at a third strength, which is what keeps
                  // the viewer's place.
                  const on = lit < 0 || lit === bi;
                  const o = interpolate(
                    ms,
                    [step?.startMs ?? 0, (step?.startMs ?? 0) + 320],
                    [on ? 0.6 : 0.5, on ? 1 : 0.3],
                    {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'},
                  );
                  return (
                    <div key={bi} style={{opacity: o, marginBottom: LINE_H * 0.55}}>
                      {lines.map((l, li) => (
                        <div key={li} style={{display: 'flex'}}>
                          <span
                            style={{
                              width: GUTTER_W,
                              flexShrink: 0,
                              textAlign: 'right',
                              paddingRight: 20,
                              color: withAlpha(theme.panelMuted, 0.34),
                              fontSize: BODY_FS - 4,
                            }}
                          >
                            {l.no}
                          </span>
                          <span
                            style={{
                              whiteSpace: 'pre-wrap',
                              paddingRight: 26,
                              ...lineStyle(l.text, theme),
                            }}
                          >
                            {inline(l.text, theme, `${bi}-${li}`)}
                          </span>
                        </div>
                      ))}
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        </div>

        {/* The margin: what to notice about the section that is lit. */}
        <div style={{width: 250, paddingTop: CHROME_H + TAB_H + 26}}>
          {note ? (
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 19,
                lineHeight: 1.45,
                color: theme.accentText,
                opacity: arrive((step?.startMs ?? sceneStart) + 260, 420),
              }}
            >
              <span style={{opacity: 0.5}}>····&nbsp;</span>
              {note}
            </div>
          ) : null}
        </div>
      </div>
    </Stage>
  );
};
