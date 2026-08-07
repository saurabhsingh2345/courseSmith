import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// ShellScene: a terminal that is actually being used, not a screenshot of one.
//
// A career-switcher's first hour with a command line is spent believing the
// terminal is a place where things are already true. It is not: it is a place
// where you TYPE, and then something answers. So the command types in,
// character by character, across the full span of its beat — not as an effect,
// but because the pause between "$" and the first output is the single most
// important beat of the interaction and a pre-filled line deletes it.
//
// The typing rate is derived from the beat's own duration rather than fixed,
// so a sixty-character command over a four-second beat and a twelve-character
// one over the same beat both finish with room to breathe before the output
// lands. The last fifth of the span is left clear on purpose: a command that
// finishes typing exactly as the beat ends reads as rushed.
//
// The window chrome — three traffic-light dots, the host in the title bar — is
// there to buy recognition, not realism. The dots are drawn in theme tokens,
// not in the operating system's colours, because this is a chart of a terminal
// and a chart should stay in the palette. Everything inside the card is
// fontMono, including the notes, because the moment one line of a terminal is
// set in a proportional face the whole illusion of a fixed grid goes.
//
// Output lines fade in staggered, top to bottom, at a stride slow enough to
// read as printing and fast enough not to become the subject. The note rides
// to the right of the entry it belongs to, joined by a short dashed leader in
// accent — an annotation on a transcript, which is exactly what a "the thing
// to notice here" line is.
//
// One glow maximum: the cursor block on the line currently being typed.

const CARD_W = Math.min(STAGE_W, 1420);
const NOTE_W = 300;
const BODY_W = CARD_W - NOTE_W - 24;
const CHROME_H = 52;

type Entry = {cmd: string; output: string[]; note: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'prompt' | 'type' | 'output' | 'recap';
  at?: number;
  typed: number[];
  shown: number[];
};

export const ShellScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const host = String(props.host ?? '');
  const entries = (Array.isArray(props.entries) ? props.entries : []) as Entry[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (entries.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const typing = step.show === 'type' ? step.at ?? -1 : -1;
  const printing = step.show === 'output' ? step.at ?? -1 : -1;
  const typed = new Set(Array.isArray(step.typed) ? step.typed : []);
  const shown = new Set(Array.isArray(step.shown) ? step.shown : []);
  const recap = step.show === 'recap';

  // The beat's own length in frames, so the typing paces itself to the shot.
  const stepFrames = Math.max(1, ((step.endMs - step.startMs) / 1000) * FPS);
  const typeProgress = interpolate(sinceStep, [3, stepFrames * 0.8], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const caret = Math.floor(frame / 14) % 2 === 0;
  const settle = recap
    ? spring({frame: sinceStep - 2, fps, config: {damping: 14, mass: 0.6}, durationInFrames: 28})
    : 0;

  const dot = (colour: string): React.ReactNode => (
    <span style={{width: 13, height: 13, borderRadius: 7, background: colour, display: 'block'}} />
  );

  return (
    <Stage justify="center">
      <SceneHeader
        theme={theme}
        title={String(props.title ?? '')}
        emphasis={props.emphasis as string | undefined}
        emphasisRole={props.emphasisRole as string | undefined}
        size="compact"
        marginBottom={20}
      />

      <div style={{width: CARD_W, display: 'flex', alignItems: 'flex-start', gap: 24}}>
        <div
          style={{
            width: BODY_W,
            flexShrink: 0,
            borderRadius: 16,
            overflow: 'hidden',
            background: withAlpha(theme.ink, 0.85),
            border: `2px solid ${recap ? withAlpha(theme.accent, 0.28 + 0.25 * settle) : theme.surfaceBorder}`,
            boxShadow: `10px 12px 0 ${withAlpha(theme.ink, 0.4)}`,
          }}
        >
          {/* Chrome. Recognisable, but in the palette. */}
          <div
            style={{
              height: CHROME_H,
              display: 'flex',
              alignItems: 'center',
              gap: 9,
              paddingInline: 18,
              background: withAlpha(theme.surface, 0.9),
              borderBottom: `1.5px solid ${theme.surfaceBorder}`,
            }}
          >
            {dot(withAlpha(theme.accentRival, 0.85))}
            {dot(withAlpha(theme.accentLimit, 0.85))}
            {dot(withAlpha(theme.accentQuantity, 0.85))}
            <span
              style={{
                marginLeft: 'auto',
                marginRight: 'auto',
                fontFamily: theme.fontMono,
                fontSize: 16,
                letterSpacing: 1.6,
                color: theme.textMuted,
              }}
            >
              {host}@shell — bash
            </span>
          </div>

          <div style={{padding: 26, minHeight: 330, display: 'flex', flexDirection: 'column', gap: 16}}>
            {entries.map((e, i) => {
              const isTyping = i === typing;
              const hasTyped = typed.has(i) || shown.has(i);
              if (!hasTyped && !isTyping) return null;
              const cmd = String(e.cmd ?? '');
              const chars = isTyping ? Math.min(cmd.length, Math.floor(typeProgress * cmd.length)) : cmd.length;
              const out = Array.isArray(e.output) ? e.output : [];
              const outOn = shown.has(i);
              const isPrinting = i === printing;
              return (
                <div key={i} style={{display: 'flex', flexDirection: 'column', gap: 6}}>
                  <div style={{display: 'flex', alignItems: 'baseline', gap: 12}}>
                    <span style={{fontFamily: theme.fontMono, fontSize: 25, color: theme.accentQuantity, flexShrink: 0}}>
                      {host} $
                    </span>
                    <span style={{fontFamily: theme.fontMono, fontSize: 25, color: theme.text, letterSpacing: 0.3}}>
                      {cmd.slice(0, chars)}
                      {isTyping ? (
                        <span
                          style={{
                            display: 'inline-block',
                            width: 13,
                            height: 25,
                            marginLeft: 2,
                            verticalAlign: 'text-bottom',
                            background: theme.accent,
                            // The one glow, on the one thing that is live.
                            boxShadow: `0 0 16px ${withAlpha(theme.accent, 0.55)}`,
                          }}
                        />
                      ) : null}
                    </span>
                  </div>
                  {outOn && out.length > 0 ? (
                    <div style={{display: 'flex', flexDirection: 'column', paddingLeft: 8}}>
                      {out.map((line, j) => (
                        <span
                          key={j}
                          style={{
                            fontFamily: theme.fontMono,
                            fontSize: out.length > 5 ? 20 : 23,
                            lineHeight: 1.42,
                            color: theme.textMuted,
                            whiteSpace: 'pre-wrap',
                            opacity: isPrinting
                              ? interpolate(sinceStep, [4 + j * 4, 16 + j * 4], [0, 1], {
                                  extrapolateLeft: 'clamp',
                                  extrapolateRight: 'clamp',
                                })
                              : 1,
                          }}
                        >
                          {line}
                        </span>
                      ))}
                    </div>
                  ) : null}
                </div>
              );
            })}

            {/* The waiting prompt, before anything has been typed. */}
            {typed.size === 0 && typing < 0 ? (
              <div style={{display: 'flex', alignItems: 'baseline', gap: 12}}>
                <span style={{fontFamily: theme.fontMono, fontSize: 25, color: theme.accentQuantity}}>{host} $</span>
                <span
                  style={{
                    display: 'inline-block',
                    width: 13,
                    height: 25,
                    background: caret ? theme.accent : 'transparent',
                  }}
                />
              </div>
            ) : null}
          </div>
        </div>

        {/* Notes: annotations on the transcript, in accent, at the right. */}
        <div style={{width: NOTE_W, flexShrink: 0, paddingTop: CHROME_H + 26, display: 'flex', flexDirection: 'column', gap: 20}}>
          {entries.map((e, i) => {
            const note = String(e.note ?? '');
            const on = shown.has(i) && note !== '';
            const reveal = on
              ? i === printing
                ? interpolate(sinceStep, [10, 26], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
                : 1
              : 0;
            if (!note) return null;
            return (
              <div
                key={i}
                style={{
                  display: 'flex',
                  alignItems: 'flex-start',
                  gap: 10,
                  opacity: reveal,
                  transform: `translateX(${(1 - reveal) * 14}px)`,
                }}
              >
                <svg width={22} height={22} style={{flexShrink: 0, marginTop: 6}}>
                  <line
                    x1={0}
                    y1={11}
                    x2={22}
                    y2={11}
                    stroke={withAlpha(theme.accent, 0.75)}
                    strokeWidth={2}
                    strokeDasharray="2 5"
                    strokeLinecap="round"
                  />
                </svg>
                <span style={{fontFamily: theme.fontMono, fontSize: 19, lineHeight: 1.35, color: theme.accentText}}>
                  {note}
                </span>
              </div>
            );
          })}
        </div>
      </div>
    </Stage>
  );
};
