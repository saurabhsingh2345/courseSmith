import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_W} from './Stage';

// AnatomyScene is one artefact held still while its pieces are named in turn.
//
// The artefact is mounted once and never re-animates. That is the whole premise:
// the viewer reads the line, then finds each piece inside the shape they have
// already read. A subject that re-landed every beat would be a subject they had
// to re-read before every callout, which is the opposite of the point.
//
// Character positions come from Go as rune spans (snippet_anatomy.go resolves
// each part's quoted text against the subject). Nothing here searches for
// anything — a second implementation of "which characters did they mean" would
// eventually disagree with the first, and the failure would be a callout
// pointing at the wrong word.

const COL_W = Math.min(STAGE_W, 1560);
/** Monospace advance as a fraction of font size — matched to the theme's mono. */
const CHAR_RATIO = 0.6;

type Part = {label: string; note: string; start: number; end: number};
type Step = {startMs: number; endMs: number; part?: number; whole?: boolean};

export const AnatomyScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const title = String(props.title ?? '');
  const subject = String(props.subject ?? '');
  const parts = (Array.isArray(props.parts) ? props.parts : []) as Part[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];

  if (!subject || steps.length === 0) return null;

  const chars = [...subject];
  // The type size is chosen so the whole line fits the column. Every callout
  // below is positioned in these units, so this is the one number the layout
  // hangs off.
  const fontSize = Math.min(52, Math.floor((COL_W - 80) / Math.max(1, chars.length) / CHAR_RATIO));
  const charW = fontSize * CHAR_RATIO;
  const lineW = charW * chars.length;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const lit = step.whole ? -1 : (step.part ?? -1);
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const enter = spring({
    frame: sinceStep,
    fps,
    config: {damping: 200, mass: 0.5},
    durationInFrames: 14,
  });

  /** Which part, if any, a character index belongs to. */
  const partAt = (i: number): number => parts.findIndex((p) => i >= p.start && i < p.end);

  const active = lit >= 0 && lit < parts.length ? parts[lit] : null;
  // The callout hangs under the middle of the lit run.
  const anchorX = active ? ((active.start + active.end) / 2) * charW : 0;
  // The callout block's width, so it can be centred on the anchor and clamped
  // to the line rather than running off whichever edge the part sits near.
  const calloutW = Math.min(720, lineW);

  return (
    <Stage justify="center">
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={40} />

      <div style={{width: lineW, position: 'relative'}}>
        {/* The artefact. Rendered per character so a span can be lit without
            re-flowing the line — splitting it into three strings would let the
            browser re-measure and the callout would drift off its column. */}
        <div
          style={{
            fontFamily: theme.fontMono,
            fontSize,
            lineHeight: 1.35,
            whiteSpace: 'pre',
            letterSpacing: 0,
          }}
        >
          {chars.map((ch, i) => {
            const mine = partAt(i);
            const isLit = lit < 0 ? mine >= 0 : mine === lit;
            const dim = lit < 0 ? 1 : isLit ? 1 : 0.34;
            return (
              <span
                key={i}
                style={{
                  display: 'inline-block',
                  width: charW,
                  textAlign: 'center',
                  color: isLit && lit >= 0 ? theme.accentText : theme.text,
                  opacity: dim,
                  // Only the lit run is backed, so the highlight reads as a
                  // selection rather than as syntax colouring.
                  backgroundColor: isLit && lit >= 0 ? `${theme.accent}26` : 'transparent',
                }}
              >
                {ch === ' ' ? ' ' : ch}
              </span>
            );
          })}
        </div>

        {/* Every part keeps a tick under it for the whole clip, so the artefact
            reads as something that HAS parts even in the overview beats. */}
        <div style={{position: 'relative', height: 18, marginTop: 6}}>
          {parts.map((p, i) => (
            <div
              key={i}
              style={{
                position: 'absolute',
                left: p.start * charW,
                width: (p.end - p.start) * charW,
                top: 0,
                height: 3,
                borderRadius: 2,
                backgroundColor: lit === i ? theme.accent : theme.textMuted,
                opacity: lit === i ? 1 : 0.28,
              }}
            />
          ))}
        </div>

        {/* The callout: a line down from the lit run to its label and note. */}
        {active && (
          <div
            style={{
              position: 'absolute',
              left: 0,
              top: fontSize * 1.35 + 24,
              width: lineW,
              opacity: enter,
            }}
          >
            {/* A straight drop to a block centred under the lit run and
                clamped to the line. The first version curved the tail toward
                whichever half of the frame the anchor was in and aligned the
                text to match, which meant the label moved differently depending
                on which part was lit — motion that carried no information and
                made two neighbouring parts read as different kinds of thing. */}
            <svg width={lineW} height={46} style={{display: 'block', overflow: 'visible'}}>
              <path
                d={`M${anchorX} 0 L${anchorX} 38`}
                fill="none"
                stroke={theme.accent}
                strokeWidth={2.5}
                strokeLinecap="round"
                pathLength={1}
                strokeDasharray={1}
                strokeDashoffset={1 - enter}
              />
              <circle cx={anchorX} cy={0} r={4} fill={theme.accent} opacity={enter} />
            </svg>
            <div
              style={{
                width: calloutW,
                marginLeft: Math.max(0, Math.min(anchorX - calloutW / 2, lineW - calloutW)),
                textAlign: 'center',
              }}
            >
              <div
                style={{
                  fontFamily: theme.fontDisplay,
                  fontSize: 36,
                  fontWeight: 700,
                  color: theme.accentText,
                }}
              >
                {active.label}
              </div>
              <div
                style={{
                  marginTop: 8,
                  fontFamily: theme.fontBody,
                  fontSize: 28,
                  lineHeight: 1.35,
                  color: theme.textMuted,
                }}
              >
                {active.note}
              </div>
            </div>
          </div>
        )}

        {/* On an overview beat the labels all sit under their own runs at once,
            which is what makes the closing shot the finished diagram. */}
        {!active && (
          <div
            style={{
              position: 'relative',
              height: 90,
              marginTop: 18,
              opacity: interpolate(sinceStep, [0, 12], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
            }}
          >
            {parts.map((p, i) => {
              const mid = ((p.start + p.end) / 2) * charW;
              return (
                <div
                  key={i}
                  style={{
                    position: 'absolute',
                    left: mid,
                    transform: 'translateX(-50%)',
                    // Alternating rows, so neighbouring labels on a dense line
                    // cannot collide.
                    top: i % 2 === 0 ? 0 : 44,
                    fontFamily: theme.fontBody,
                    fontSize: 22,
                    color: theme.textMuted,
                    whiteSpace: 'nowrap',
                  }}
                >
                  {p.label}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </Stage>
  );
};
