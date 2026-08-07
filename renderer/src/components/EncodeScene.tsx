import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// EncodeScene: three stations and two arrows.
//
// The layout is a left-to-right journey because the clip is making a claim
// about ORDER — the character gets a number first, and only then does the
// number get written down as bytes. A stacked or centred arrangement would say
// "here are three facts about this character", which is the misunderstanding
// the template exists to correct. The arrows are the argument; the stations are
// what they connect.
//
// The glyph is set enormous, far larger than anything else on the stage. That
// is a deliberate imbalance: the character is the only thing here the viewer
// already knows, and the picture works by starting somewhere completely
// familiar and travelling somewhere completely unfamiliar. A modestly sized
// glyph would look like the first of three equal boxes.
//
// Each byte box shows its eight bits with the UTF-8 marker prefix tinted apart
// from the payload, and its hex value beneath in the accent. The tinting is the
// answer to the question the row provokes — how does anything know where one
// character ends — and it costs nothing to show, so it is always shown even
// though no beat is spent on it. The boxes land one at a time on a spring;
// which of them have landed by a given beat was decided in Go.
//
// One glow maximum, on the box that is landing. The stations reveal by opacity
// and a small horizontal drift in the direction of travel, so the whole frame
// keeps moving the way the journey does.

const LANE_W = Math.min(STAGE_W, 1700);
const GLYPH_BOX = 240;
const BYTE_W = 156;
const BYTE_GAP = 16;
const ARROW_W = 74;

type ByteBox = {hex: string; bits: string; marker: string; payload: string; lead: boolean};
type Step = {
  startMs: number;
  endMs: number;
  show: 'glyph' | 'codepoint' | 'bytes' | 'note';
  at?: number;
  landed: number;
};

const RANK: Record<string, number> = {glyph: 0, codepoint: 1, bytes: 2, note: 3};

export const EncodeScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const glyph = String(props.glyph ?? '');
  const codepoint = String(props.codepoint ?? '');
  const note = String(props.note ?? '');
  const bytes = (Array.isArray(props.bytes) ? props.bytes : []) as ByteBox[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (bytes.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const reached = Math.max(...steps.slice(0, idx + 1).map((s) => RANK[s.show] ?? 0));
  const cpUp = reached >= RANK.codepoint;
  const bytesUp = reached >= RANK.bytes;
  const landing = step.show === 'bytes';

  // How many boxes are down. Within a byte beat they arrive one at a time; the
  // count each beat ends on came from Go.
  const before = landing ? (steps[idx - 1]?.landed ?? 0) : step.landed;
  const target = step.landed;
  const dealInterval = 12;
  const landed = landing
    ? Math.min(target, before + Math.max(0, Math.floor((sinceStep - 4) / dealInterval) + 1))
    : target;
  const current = landing && landed > before ? landed - 1 : -1;

  const enter = (on: boolean, delay: number) =>
    on
      ? interpolate(sinceStep, [delay, delay + 16], [0, 1], {
          extrapolateLeft: 'clamp',
          extrapolateRight: 'clamp',
        })
      : 0;

  const cpReveal = step.show === 'codepoint' ? enter(true, 2) : cpUp ? 1 : 0;

  const arrow = (on: number): React.ReactNode => (
    <svg width={ARROW_W} height={28} viewBox="0 0 74 28" style={{flexShrink: 0, opacity: on}}>
      <path
        d={`M2 14 H${8 + 54 * on}`}
        stroke={theme.line}
        strokeWidth={2.5}
        strokeLinecap="round"
        fill="none"
      />
      <path
        d="M60 6 L70 14 L60 22"
        stroke={theme.accent}
        strokeWidth={3}
        strokeLinecap="round"
        strokeLinejoin="round"
        fill="none"
        opacity={on}
      />
    </svg>
  );

  const stationLabel = (text: string, colour: string): React.ReactNode => (
    <div
      style={{
        fontFamily: theme.fontMono,
        fontSize: 15,
        letterSpacing: 2.6,
        textTransform: 'uppercase',
        color: colour,
        marginBottom: 14,
      }}
    >
      {text}
    </div>
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

      <div style={{width: LANE_W, display: 'flex', alignItems: 'center', gap: 22}}>
        {/* Station one: the only thing here the viewer already knows. */}
        <div style={{flexShrink: 0}}>
          {stationLabel('character', theme.textMuted)}
          <div
            style={{
              width: GLYPH_BOX,
              height: GLYPH_BOX,
              borderRadius: 24,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background: withAlpha(theme.surface, 0.9),
              border: `2px solid ${theme.surfaceBorder}`,
              fontFamily: theme.fontDisplay,
              fontSize: 150,
              lineHeight: 1,
              color: theme.text,
            }}
          >
            {glyph}
          </div>
        </div>

        {arrow(cpReveal)}

        {/* Station two: Unicode's fact about the character. */}
        <div style={{flexShrink: 0, opacity: cpReveal, transform: `translateX(${(1 - cpReveal) * -22}px)`}}>
          {stationLabel('codepoint', theme.textMuted)}
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 62,
              fontWeight: 700,
              letterSpacing: -1,
              color: theme.accentText,
            }}
          >
            {codepoint}
          </div>
          <div style={{fontFamily: theme.fontBody, fontSize: 21, color: theme.textMuted, marginTop: 8}}>
            assigned, not chosen
          </div>
        </div>

        {arrow(bytesUp ? 1 : 0)}

        {/* Station three: the choice. */}
        <div style={{flexShrink: 0, opacity: bytesUp ? 1 : 0}}>
          {stationLabel(`utf-8 · ${bytes.length} byte${bytes.length === 1 ? '' : 's'}`, theme.textMuted)}
          <div style={{display: 'flex', gap: BYTE_GAP}}>
            {bytes.map((b, i) => {
              const down = i < landed;
              const isCurrent = i === current;
              const pop = isCurrent
                ? spring({
                    frame: sinceStep - 4 - (landed - 1 - before) * dealInterval,
                    fps,
                    config: {damping: 13, mass: 0.55},
                    durationInFrames: 24,
                  })
                : 1;
              return (
                <div
                  key={i}
                  style={{
                    width: BYTE_W,
                    borderRadius: 14,
                    padding: '16px 10px 12px',
                    textAlign: 'center',
                    background: withAlpha(theme.surface, 0.9),
                    border: `2px solid ${isCurrent ? theme.accentQuantity : theme.surfaceBorder}`,
                    opacity: down ? (isCurrent ? pop : 1) : 0,
                    transform: `translateY(${(1 - (down ? pop : 0)) * 18}px)`,
                    // The one glow: the byte landing right now.
                    boxShadow: isCurrent ? `0 0 26px ${withAlpha(theme.accentQuantity, 0.5)}` : undefined,
                  }}
                >
                  <div style={{fontFamily: theme.fontMono, fontSize: 22, letterSpacing: 1.5}}>
                    <span style={{color: theme.accentRival}}>{b.marker}</span>
                    <span style={{color: theme.text}}>{b.payload}</span>
                  </div>
                  <div
                    style={{
                      fontFamily: theme.fontMono,
                      fontSize: 34,
                      fontWeight: 700,
                      color: theme.accentQuantity,
                      marginTop: 10,
                    }}
                  >
                    {b.hex}
                  </div>
                  <div
                    style={{
                      fontFamily: theme.fontBody,
                      fontSize: 16,
                      letterSpacing: 1.2,
                      textTransform: 'uppercase',
                      color: theme.textMuted,
                      marginTop: 6,
                    }}
                  >
                    {b.lead ? 'lead' : 'cont'}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* The closer: one line under the whole journey. */}
      <div
        style={{
          width: LANE_W,
          marginTop: 44,
          minHeight: 54,
          fontFamily: theme.fontBody,
          fontSize: 34,
          color: theme.text,
          opacity: step.show === 'note' ? enter(true, 3) : 0,
          transform: `translateY(${(1 - (step.show === 'note' ? enter(true, 3) : 0)) * 12}px)`,
        }}
      >
        {note}
      </div>
    </Stage>
  );
};
