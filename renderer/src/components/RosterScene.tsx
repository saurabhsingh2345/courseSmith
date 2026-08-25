import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';
import {CardLabel, CardName, CardPanel, CardTile, roleColour} from './showroomCard';

// RosterScene is two or three cards, one per kind of viewer, and every one of
// them stays readable for the whole clip.
//
// That last clause is the design, and it is a departure from every other card
// row in the catalog. `cards` and `rundown` keep an unlit card deliberately
// unreadable — the label is there, the line under it is not — because a voice is
// walking the row and a line already read has spent itself before the narration
// reaches it. This frame is built for the opposite case as well: a piece with no
// voice at all, where the viewer is the one doing the reading. A card whose
// content only exists on its own beat is, in silence, a card that is blank most
// of the time.
//
// So the two lines are on every card from the first frame, and what a beat moves
// is INK: the raised card's lines go to full text colour and its tile takes the
// accent, the others sit at muted. Nothing fades, nothing is hidden, and the row
// reads as a finished object at any frame you pause on — which is also what makes
// the closing frame worth screenshotting.
//
// The two lines are labelled rather than merely stacked. "an idea and a laptop"
// above "a working app before you learn a language" is two phrases with no stated
// relation, and the relation is the whole card: one is what you walk in with, the
// other is what you walk out with. Two small-caps labels cost eleven pixels of
// type and turn a pair of phrases into a before and after.

const BLOCK_W = Math.min(STAGE_W, 1500);

/** Card metrics: two cards get a wider measure and larger type than three. */
const metrics = (n: number) =>
  n <= 2
    ? {tile: 104, who: 52, line: 30, gap: 40, pad: 44, minH: 470}
    : {tile: 88, who: 42, line: 25, gap: 28, pad: 34, minH: 470};

type Person = {who: string; brings: string; gets: string; icon?: string; role?: string};

type Step = {startMs: number; endMs: number; show: 'row' | 'person' | 'all'; at?: number};

/** One labelled line inside a card — "BRINGS", then what they bring. */
const CardFact: React.FC<{
  theme: ResolvedTheme;
  label: string;
  value: string;
  size: number;
  lit: boolean;
  colour: string;
}> = ({theme, label, value, size, lit, colour}) => (
  <div style={{textAlign: 'left', width: '100%'}}>
    <CardLabel theme={theme} lit={lit} style={{marginBottom: 8, color: lit ? colour : withAlpha(theme.textMuted, 0.7)}}>
      {label}
    </CardLabel>
    <div
      style={{
        fontFamily: theme.fontBody,
        fontSize: size,
        fontWeight: lit ? 600 : 500,
        lineHeight: 1.32,
        // The dim state is a quieter ink, never a lower opacity: on paper a card
        // at reduced opacity fades into the page and the row loses the count it
        // exists to show. See showroomCard's CardPanel.
        color: lit ? theme.text : theme.textMuted,
      }}
    >
      {value}
    </div>
  </div>
);

export const RosterScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const people = (Array.isArray(props.people) ? props.people : []) as Person[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  const closer = String(props.closer ?? '');
  if (people.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;
  const all = step.show === 'all';
  const raised = step.show === 'person' ? (step.at ?? 0) : -1;
  const m = metrics(people.length);

  return (
    <Stage justify="center">
      <div style={{width: BLOCK_W, display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
        <SceneHeader
          theme={theme}
          title={String(props.title ?? '')}
          emphasis={props.emphasis as string | undefined}
          emphasisRole={props.emphasisRole as string | undefined}
          marginBottom={52}
        />

        <div style={{display: 'flex', gap: m.gap, width: '100%', alignItems: 'stretch'}}>
          {people.map((p, i) => {
            // Lit means "readable" and is true of every card once the row is up;
            // selected means "being spoken about now". The split is what lets the
            // opening frame show three finished cards without picking one.
            const selected = all || i === raised;
            const lit = all || raised < 0 || i === raised;
            const colour = roleColour(theme, p.role);
            // The stagger is the row assembling left to right on the opening
            // beat. It moves the card, not its contents' legibility.
            const on = interpolate(frame, [3 + i * 5, 20 + i * 5], [0, 1], {
              extrapolateLeft: 'clamp',
              extrapolateRight: 'clamp',
            });
            const rise = selected
              ? interpolate(sinceStep, [0, 14], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                })
              : 0;
            return (
              <CardPanel
                key={i}
                theme={theme}
                colour={colour}
                selected={selected}
                style={{
                  flex: 1,
                  minHeight: m.minH,
                  padding: m.pad,
                  alignItems: 'flex-start',
                  textAlign: 'left',
                  justifyContent: 'flex-start',
                  gap: 26,
                  // A raised card lifts four pixels. Small on purpose: the seat
                  // shadow under it does the work, and a card that jumps reads as
                  // an interface responding to a click.
                  transform: `translateY(${(1 - on) * 20 - rise * 4}px)`,
                }}
              >
                <div style={{display: 'flex', alignItems: 'center', gap: 20, width: '100%'}}>
                  <CardTile theme={theme} subject={{title: p.who, icon: p.icon, role: p.role}} size={m.tile} lit={selected} />
                  <CardName theme={theme} size={m.who} lit={lit}>
                    {p.who}
                  </CardName>
                </div>

                <div
                  style={{
                    height: 1,
                    width: '100%',
                    background: selected ? withAlpha(colour, 0.45) : theme.surfaceBorder,
                  }}
                />

                <CardFact theme={theme} label="arrives with" value={p.brings} size={m.line} lit={lit} colour={colour} />
                <CardFact theme={theme} label="leaves with" value={p.gets} size={m.line} lit={lit} colour={colour} />
              </CardPanel>
            );
          })}
        </div>

        {/* The closer sits under the row and only on the closing beat. It is the
            one line that belongs to nobody on the frame, so it cannot live in a
            card, and reserving its height keeps the row from moving when it
            lands. */}
        <div style={{minHeight: 96, marginTop: 40, display: 'flex', alignItems: 'center'}}>
          {all && closer ? (
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: 38,
                fontWeight: 700,
                letterSpacing: -0.5,
                lineHeight: 1.3,
                color: theme.text,
                textAlign: 'center',
                maxWidth: 1180,
                opacity: interpolate(sinceStep, [2, 16], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                }),
              }}
            >
              {closer}
            </div>
          ) : null}
        </div>
      </div>
    </Stage>
  );
};
