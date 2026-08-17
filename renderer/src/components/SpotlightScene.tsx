import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, seat, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';
import {iconFor} from './icons';
import {CardName, CardPanel, CardSubject, CardTile, cardColour} from './showroomCard';

// SpotlightScene is one card on the left and its claims stacked on the right.
//
// This is the family's asymmetric frame, and the asymmetry is the content. A
// centred composition says "here is a diagram"; a thing on one side with claims
// on the other says "here is a thing, and here is what I say about it". The
// catalog had no frame that said the second sentence — measured across it, almost
// every scene centres — which is most of why templates as different as a
// leaderboard and a memory budget came out looking like the same slide.
//
// The rows LAND rather than brighten, and that is the one place this component
// departs from the rest of the family. A card in a row is on screen from the first
// frame because the count is the point — the viewer should know there are three
// things before a word is spoken. A claim is not a count. There is no value in
// knowing that three unreadable claims are coming, and a ghosted stack of them
// invites the eye to read ahead of the voice. So the stack grows.
//
// What holds still is the CARD, and it has to: the rows are being added below the
// last one, so if the block were centred as a whole, every row landing would shift
// everything already on screen. The card is pinned to the top of the block and the
// stack grows downward from the same line, which is why the frame's lower band is
// deliberately empty on the opening beat.

const BLOCK_W = Math.min(STAGE_W, 1620);

// The hero card's width. Fixed rather than a fraction, because the rows beside it
// want a measure that does not change with the card — and the card is a product
// shot, which has a natural size.
const CARD_W = 520;
const GAP = 66;
const TILE = 190;
const CARD_H = 520;

/** Row metrics, tightened as the stack grows. */
const rowMetrics = (n: number) => {
  if (n <= 2) return {h: 118, chip: 74, icon: 34, font: 40, gap: 26};
  if (n === 3) return {h: 104, chip: 66, icon: 30, font: 36, gap: 22};
  return {h: 92, chip: 58, icon: 27, font: 32, gap: 18};
};

type Point = {text: string; icon?: string};

type Step = {
  startMs: number;
  endMs: number;
  show: 'card' | 'point' | 'all';
  at?: number;
  shown?: number;
};

export const SpotlightScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const card = (props.card ?? {}) as CardSubject;
  const points = (Array.isArray(props.points) ? props.points : []) as Point[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (!card.title || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const shown = step.shown ?? 0;
  const newest = step.show === 'point' ? (step.at ?? -1) : -1;
  const colour = cardColour(theme, card);
  const m = rowMetrics(points.length);

  const settle = interpolate(sinceStep, [0, 12], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const enter = interpolate(frame, [2, 22], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  return (
    <Stage justify="center">
      <div style={{width: BLOCK_W}}>
        <SceneHeader
          theme={theme}
          title={String(props.title ?? '')}
          emphasis={props.emphasis as string | undefined}
          emphasisRole={props.emphasisRole as string | undefined}
          size="compact"
          marginBottom={48}
        />

        {/* Top-aligned, not centred. The stack grows downward from the card's own
            top edge, so anything that centred the two columns against each other
            would move the card every time a row landed. */}
        <div style={{display: 'flex', alignItems: 'flex-start', gap: GAP}}>
          <CardPanel
            theme={theme}
            colour={colour}
            // NOT selected, and the first draft had it selected on the reasoning
            // that the hero card is always the subject. That was wrong on the
            // frame: a brand-coloured rim is a comparative mark — it means "this
            // one rather than that one" — and drawn around the only card on screen
            // it has nothing to compare against, so the eye reads it as a state
            // instead. In peach it read as a warning. What the card gets instead
            // is the deeper seating, below: the hero is the thing held closest to
            // the viewer, which is elevation rather than colour.
            selected={false}
            style={{
              width: CARD_W,
              height: CARD_H,
              flex: 'none',
              padding: '46px 34px',
              justifyContent: 'center',
              boxShadow: seat(theme, 'lifted'),
              opacity: enter,
              transform: `translateY(${(1 - enter) * 18}px)`,
            }}
          >
            <div style={{marginBottom: 30}}>
              <CardTile theme={theme} subject={card} size={TILE} lit />
            </div>
            <CardName theme={theme} size={56} lit>
              {card.title}
            </CardName>
            {card.note ? (
              <div
                style={{
                  marginTop: 18,
                  fontFamily: theme.fontBody,
                  fontSize: 28,
                  lineHeight: 1.35,
                  fontWeight: 500,
                  color: theme.textMuted,
                }}
              >
                {card.note}
              </div>
            ) : null}
          </CardPanel>

          <div style={{flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column', gap: m.gap}}>
            {points.map((pt, i) => {
              if (i >= shown) return null;
              const Icon = iconFor(pt.icon);
              const isNewest = i === newest;
              // Each row's own landing, so a row that arrived two beats ago is not
              // re-animated when the next one lands.
              const land = isNewest ? settle : 1;
              return (
                <div
                  key={i}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 24,
                    minHeight: m.h,
                    padding: `0 32px`,
                    borderRadius: 20,
                    background: theme.surface,
                    border: `1px solid ${theme.surfaceBorder}`,
                    boxShadow: seat(theme, 'resting'),
                    opacity: land,
                    // Sideways rather than down. A row that drops in from above
                    // travels through the row already there; one that comes in
                    // from the right arrives in empty space, which is where it is
                    // going to live.
                    transform: `translateX(${(1 - land) * 34}px)`,
                  }}
                >
                  <div
                    style={{
                      width: m.chip,
                      height: m.chip,
                      flex: 'none',
                      borderRadius: m.chip * 0.28,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      // The chip is neutral, not the card's colour. Every row
                      // belongs to the same subject, so colouring them in its
                      // brand would put five copies of one logo's colour down the
                      // frame and leave nothing for the card itself to be.
                      background: withAlpha(theme.text, 0.055),
                    }}
                  >
                    <Icon size={m.icon} strokeWidth={2} color={theme.text} />
                  </div>
                  <div
                    style={{
                      fontFamily: theme.fontDisplay,
                      fontSize: m.font,
                      fontWeight: 700,
                      letterSpacing: -0.6,
                      lineHeight: 1.2,
                      color: theme.text,
                    }}
                  >
                    {pt.text}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </Stage>
  );
};
