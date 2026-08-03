import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// RatioScene is two bars on one scale, with the fraction named between them.
//
// Both bars are measured against the SAME full width — the reference fills it and
// the subject fills its share. That is the only honest way to draw a proportion:
// two bars each scaled to their own maximum would both be full, and the picture
// would say the two things are equal while the narration says one is a third of
// the other.
//
// The subject's bar sits directly under the reference's, left edges aligned, with
// no gap between the rows. Aligned and adjacent is what lets the eye do the
// division on its own — a viewer sees the shortfall as a length before the phrase
// arrives, which is what makes the phrase land rather than inform.
//
// The phrase is set at the largest size on the frame, larger than either number.
// The numbers are evidence; the phrase is the thing the clip exists to be
// remembered by, and the typography says which is which.

const BAR_W = Math.min(STAGE_W, 1240);
const BAR_H = 64;

type Side = {label: string; value: number; role: string; frac?: number};
type Step = {startMs: number; endMs: number; show: 'reference' | 'subject' | 'fraction' | 'read'};

const roleColour = (theme: ResolvedTheme, role: string): string => {
  switch (role) {
    case 'quantity':
      return theme.accentQuantity;
    case 'limit':
      return theme.accentLimit;
    case 'rival':
      return theme.accentRival;
    default:
      return theme.textMuted;
  }
};

const fmt = (n: number): string => n.toLocaleString();

export const RatioScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const unit = String(props.unit ?? '');
  const reference = props.reference as Side | undefined;
  const subject = props.subject as Side | undefined;
  const phrase = String(props.phrase ?? '');
  const note = String(props.note ?? '');
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (!reference || !subject || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const reached = (show: Step['show']) =>
    steps.slice(0, idx + 1).some((s) => s.show === show);
  const showSubject = reached('subject');
  const showFraction = reached('fraction');

  const refColour = roleColour(theme, reference.role);
  const subColour = roleColour(theme, subject.role);
  const frac = subject.frac ?? 0;

  const enter = interpolate(sinceStep, [0, 16], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const refRun = idx === 0
    ? interpolate(sinceStep, [2, 24], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
    : 1;
  const subRun = step.show === 'subject'
    ? interpolate(sinceStep, [2, 24], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
    : 1;
  const phraseIn = step.show === 'fraction'
    ? interpolate(sinceStep, [2, 20], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
    : 1;

  const row = (side: Side, colour: string, width: number, value: number) => (
    <div style={{marginBottom: 14}}>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'baseline',
          marginBottom: 8,
        }}
      >
        <span
          style={{
            fontFamily: theme.fontMono,
            fontSize: 15,
            letterSpacing: 2.6,
            textTransform: 'uppercase',
            color: colour,
          }}
        >
          {side.label}
        </span>
        <span
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 34,
            fontWeight: 800,
            color: theme.text,
            fontVariantNumeric: 'tabular-nums',
          }}
        >
          {fmt(Math.round(value))}
          <span
            style={{
              fontFamily: theme.fontMono,
              fontSize: 17,
              fontWeight: 400,
              color: theme.textMuted,
              marginLeft: 6,
            }}
          >
            {unit}
          </span>
        </span>
      </div>
      {/* The track is the reference's full extent for BOTH rows, so a share is a
          share. Two bars each scaled to their own maximum would both be full. */}
      <div
        style={{
          width: '100%',
          height: BAR_H,
          borderRadius: 8,
          background: theme.surface,
          border: `1px solid ${theme.surfaceBorder}`,
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            width: `${width * 100}%`,
            height: '100%',
            background: withAlpha(colour, 0.85),
          }}
        />
      </div>
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
        marginBottom={28}
      />

      <div style={{width: BAR_W, opacity: enter}}>
        {row(reference, refColour, refRun, reference.value * refRun)}
        {showSubject
          ? row(subject, subColour, frac * subRun, subject.value * subRun)
          : null}
      </div>

      {/* The phrase, larger than either number: the figures are the evidence and
          this is what the clip is for. */}
      {showFraction ? (
        <div
          style={{
            marginTop: 30,
            textAlign: 'center',
            opacity: phraseIn,
            transform: `translateY(${(1 - phraseIn) * 10}px)`,
          }}
        >
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 104,
              fontWeight: 800,
              lineHeight: 1,
              letterSpacing: -2,
              textTransform: 'uppercase',
              color: subColour,
            }}
          >
            {phrase}
          </div>
          {note ? (
            <div
              style={{
                fontFamily: theme.fontBody,
                fontSize: 25,
                color: theme.textMuted,
                marginTop: 18,
                maxWidth: 1040,
                marginInline: 'auto',
              }}
            >
              {note}
            </div>
          ) : null}
        </div>
      ) : null}
    </Stage>
  );
};
