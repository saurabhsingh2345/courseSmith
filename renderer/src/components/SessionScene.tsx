import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, seat, withAlpha} from '../theme/theme';
import {Stage} from './Stage';

// SessionScene is an agent session, filling the frame, being read.
//
// Four things make this different from drawing a terminal, and each one took a
// deliberate decision rather than falling out of the layout.
//
// THE CAMERA MOVES. Every other scene in the catalog is a fixed frame with things
// appearing in it. Here the window arrives slightly small and pushes in over the
// whole clip, and pushes a little further when the agent asks a question. A
// session held for ninety seconds at a dead-still 100% scale reads as a
// screenshot, and nobody watches a screenshot. The push is under 5% end to end.
//
// THE TRANSCRIPT SCROLLS. Content taller than the window pulls up so the newest
// work sits just above the composer. The height is estimated in code rather than
// measured because Remotion renders each frame independently in a headless
// browser: a scroll driven by measurement would jitter between frames.
//
// THE PANEL HAS ITS OWN INKS. Text inside the window comes from panelText and
// panelMuted, never from text/textMuted. Those are derived against the PAGE, and
// on a light skin the page is bright — using them here is what made a terminal's
// output mid-grey on near-black in an earlier template.
//
// AND THE SMALL PRINT IS THE POINT. A terminal is recognised by its furniture,
// not its font: the bullet glyph on a tool call, the elbow under it, the line
// saying how much output was folded away, the spinner with its elapsed time and
// token count, the mode line along the bottom with the branch beside it. Every
// one of those is three words long and each is worth more to the illusion than
// another paragraph of body text. A window without them reads as a slide about a
// terminal; with them it reads as one somebody is using. That is why they are
// modelled as first-class fields rather than left for the author to fake inside a
// line of output.

type Event = {
  kind: 'ask' | 'say' | 'tool' | 'menu' | 'welcome' | 'spin' | 'agents';
  text: string;
  lines?: string[];
  more?: string;
  aside?: string[];
  options?: string[];
  pick?: number;
  mark?: string;
  foot?: string;
  track?: boolean;
  note?: string;
};

type Step = {
  startMs: number;
  endMs: number;
  show: 'open' | 'event' | 'chip' | 'whole';
  at?: number;
  shown?: number;
  chip?: boolean;
};

type Props = {
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: {
    app?: string;
    header?: string[];
    hint?: string;
    chip?: string;
    branch?: string;
    pr?: string;
    status?: string;
    events?: Event[];
    steps?: Step[];
  };
};

const WIN_W = 1360;
/** Width when nothing is in the margin: the window takes the room back. */
const WIN_W_WIDE = 1600;
const WIN_H = 812;
const NOTE_W = 300;
const CHROME_H = 46;
const PAD = 38;
const BODY_FS = 23;
const LINE_H = 36;
/** The composer plus its mode line, reserved at the foot of the window. */
const COMPOSER_H = 138;

/**
 * The window's traffic lights, in their real colours.
 *
 * Not theme tokens, and that is the one place in this file where a literal is
 * correct. Everything else on the frame belongs to the design system and has to
 * flip with it; these three dots belong to the operating system. Tinting them
 * toward the brand hue is what makes a drawn window look drawn — the eye knows
 * these colours the way it knows a traffic light, and any other red reads as an
 * illustration of a window rather than a window.
 */
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
 * Diff colours for lines that start with + or -.
 *
 * Literals, for the same reason the traffic lights are: green-means-added is a
 * convention of the tool world rather than a decision of this design system, and
 * running it through the brand hue would produce a diff whose additions are
 * whatever colour the course happens to be. Placed against the panel, not the
 * page, so they hold in either polarity.
 */
const DIFF_ADD = '#5ad07a';
const DIFF_DEL = '#f2777a';

/** A transcript line's colour, from how it starts. */
const lineInk = (line: string, fallback: string): string => {
  const t = line.trimStart();
  if (t.startsWith('+')) return DIFF_ADD;
  if (t.startsWith('-') && !t.startsWith('--')) return DIFF_DEL;
  return fallback;
};

/**
 * How tall an event draws, in pixels.
 *
 * Estimated rather than measured, for the reason in the file header. The numbers
 * are the layout below expressed as arithmetic, so the two have to be kept in
 * step by hand — the alternative is a scroll position that disagrees with itself
 * from one frame to the next.
 */
const eventHeight = (e: Event): number => {
  const lines = e.lines?.length ?? 0;
  switch (e.kind) {
    case 'menu':
      return (
        LINE_H +
        (e.options?.length ?? 0) * (LINE_H + 8) +
        (e.mark ? LINE_H : 0) +
        (e.track ? 54 : 0) +
        (e.foot ? LINE_H : 0) +
        34
      );
    case 'welcome':
      return Math.max(lines, e.aside?.length ?? 0) * (LINE_H - 8) + 96;
    case 'ask':
      // The tinted block has padding of its own.
      return LINE_H * (1 + lines) + 46;
    case 'spin':
      return LINE_H + 18;
    default:
      return LINE_H * (1 + lines) + (e.more ? LINE_H : 0) + 22;
  }
};

export const SessionScene = ({theme, sceneStartMs, props}: Props) => {
  const frame = useCurrentFrame();
  const ms = (frame / FPS) * 1000 + sceneStartMs;
  const events = props.events ?? [];
  const steps = props.steps ?? [];
  const header = props.header ?? [];

  const idx = Math.max(
    0,
    steps.findIndex((s) => ms >= s.startMs && ms < s.endMs),
  );
  const step = steps[idx] ?? steps[steps.length - 1];
  const sceneStart = steps[0]?.startMs ?? sceneStartMs;
  const sceneEnd = steps[steps.length - 1]?.endMs ?? sceneStart + 8000;
  const shown = step?.shown ?? 0;
  const visible = events.slice(0, shown);

  // The camera: the window settling in on arrival, and a slow push across the
  // whole clip that never stops.
  const settle = interpolate(ms, [sceneStart, sceneStart + 900], [0.962, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  // No continuous drift: it made the transcript glitter. See the note above.
  // And the lean: when the newest thing on screen is a question, push a little
  // further over a second. The viewer is being asked to decide; the frame should
  // feel like it moved closer.
  const newest = visible[visible.length - 1];
  const leaning = step?.show === 'event' && newest?.kind === 'menu';
  // The lean is now whole pixels of rise rather than scale — same "moved closer"
  // read on a question beat, and it resamples nothing.
  const lean = Math.round(
    interpolate(ms, [step?.startMs ?? 0, (step?.startMs ?? 0) + 1000], [0, leaning ? -10 : 0], {
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp',
    }),
  );
  const scale = settle;

  const viewportH = WIN_H - CHROME_H - COMPOSER_H - PAD;
  const headerH = header.length ? header.length * (LINE_H - 6) + 30 : 0;
  const scrollFor = (n: number): number => {
    const content = headerH + events.slice(0, n).reduce((sum, e) => sum + eventHeight(e), 0);
    return Math.max(0, content - viewportH);
  };
  const prevShown = steps[idx - 1]?.shown ?? 0;
  const scroll = interpolate(
    ms,
    [step?.startMs ?? 0, (step?.startMs ?? 0) + 480],
    [scrollFor(prevShown), scrollFor(shown)],
    {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'},
  );

  const arrive = (from: number, dur = 380) =>
    interpolate(ms, [from, from + dur], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'});

  const caretOn = Math.floor(((ms - sceneStart) / 1000) * 1.5) % 2 === 0;
  const lastAsk = [...visible].reverse().find((e) => e.kind === 'ask');
  const notes = visible.filter((e) => e.note);
  // Reserved on whether the SESSION has notes, not on what is showing yet:
  // sizing off current content would resize the window the moment the first
  // annotation lands, and a window that jumps sideways mid-push is worse than no
  // margin at all.
  const hasNotes = events.some((e) => e.note);
  const winW = hasNotes ? WIN_W : WIN_W_WIDE;

  const mono = theme.fontMono;
  const dim = withAlpha(theme.panelMuted, 0.72);

  return (
    <Stage>
      <div
        style={{
          display: 'flex',
          alignItems: 'flex-start',
          gap: 40,
          transform: `scale(${scale}) translateY(${lean}px)`,
        }}
      >
        <div
          style={{
            width: winW,
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
          {/* Chrome. */}
          <div
            style={{
              height: CHROME_H,
              flexShrink: 0,
              display: 'flex',
              alignItems: 'center',
              padding: '0 18px',
              gap: 9,
              background: withAlpha(theme.panelText, 0.055),
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
                fontFamily: mono,
                fontSize: 19,
                color: withAlpha(theme.panelMuted, 0.9),
                marginRight: 45,
              }}
            >
              {props.app ? `✳ ${props.app}` : ''}
            </span>
          </div>

          {/* Transcript. */}
          <div style={{flex: 1, position: 'relative', overflow: 'hidden'}}>
            <div
              style={{
                position: 'absolute',
                left: PAD,
                right: PAD,
                top: PAD - scroll,
                fontFamily: mono,
                fontSize: BODY_FS,
                lineHeight: `${LINE_H}px`,
              }}
            >
              {header.length ? (
                <div style={{marginBottom: 30, opacity: arrive(sceneStart + 200)}}>
                  {header.map((h, i) => (
                    <div key={i} style={{color: dim, lineHeight: `${LINE_H - 6}px`}}>
                      {h}
                    </div>
                  ))}
                </div>
              ) : null}

              {visible.map((e, i) => {
                const isNew = i === visible.length - 1;
                const from = isNew ? (step?.startMs ?? sceneStart) : sceneStart;
                const a = isNew ? arrive(from) : 1;
                return (
                  <div
                    key={i}
                    style={{marginBottom: 22, opacity: a, transform: `translateY(${(1 - a) * 14}px)`}}
                  >
                    {/* The greeting box a session prints on start. */}
                    {e.kind === 'welcome' ? (
                      <div
                        style={{
                          border: `1px solid ${withAlpha(theme.accent, 0.35)}`,
                          borderRadius: 10,
                          padding: '20px 24px',
                          display: 'flex',
                          gap: 34,
                        }}
                      >
                        <div style={{flex: 1}}>
                          <div style={{color: theme.panelText, marginBottom: 8}}>{e.text}</div>
                          {(e.lines ?? []).map((l, j) => (
                            <div key={j} style={{color: dim, lineHeight: `${LINE_H - 8}px`}}>
                              {l}
                            </div>
                          ))}
                        </div>
                        {e.aside?.length ? (
                          <div style={{width: 380}}>
                            {e.aside.map((l, j) => (
                              <div
                                key={j}
                                style={{
                                  color: j === 0 ? withAlpha(theme.panelText, 0.8) : dim,
                                  lineHeight: `${LINE_H - 8}px`,
                                }}
                              >
                                {l}
                              </div>
                            ))}
                          </div>
                        ) : null}
                      </div>
                    ) : null}

                    {/* What the person typed. A tinted block, because in a real
                        session your own words are set apart from the tool's. */}
                    {e.kind === 'ask' ? (
                      <div
                        style={{
                          background: withAlpha(theme.panelText, 0.07),
                          borderRadius: 8,
                          padding: '12px 16px',
                          color: theme.panelText,
                        }}
                      >
                        {e.text}
                        {(e.lines ?? []).map((l, j) => (
                          <div key={j}>{l}</div>
                        ))}
                      </div>
                    ) : null}

                    {e.kind === 'say' ? (
                      <div style={{color: withAlpha(theme.panelText, 0.88)}}>
                        <div>{e.text}</div>
                        {(e.lines ?? []).map((l, j) => (
                          <div key={j} style={{color: lineInk(l, theme.panelMuted), whiteSpace: 'pre-wrap'}}>
                            {l}
                          </div>
                        ))}
                      </div>
                    ) : null}

                    {/* A tool call: bullet, then the elbow under it. */}
                    {e.kind === 'tool' ? (
                      <div>
                        <div style={{display: 'flex', gap: 12}}>
                          <span style={{color: theme.accent}}>●</span>
                          <span style={{color: theme.panelText}}>{e.text}</span>
                        </div>
                        {(e.lines ?? []).map((l, j) => (
                          <div key={j} style={{display: 'flex', gap: 10, paddingLeft: 12}}>
                            <span style={{color: withAlpha(theme.panelMuted, 0.5)}}>{j === 0 ? '⎿' : ' '}</span>
                            <span style={{color: lineInk(l, theme.panelMuted)}}>{l}</span>
                          </div>
                        ))}
                        {e.more ? (
                          <div style={{color: withAlpha(theme.panelMuted, 0.62), paddingLeft: 34}}>
                            … {e.more}
                          </div>
                        ) : null}
                      </div>
                    ) : null}

                    {/* Subagents, working at once. */}
                    {e.kind === 'agents' ? (
                      <div>
                        <div style={{display: 'flex', gap: 12}}>
                          <span style={{color: theme.accent}}>●</span>
                          <span style={{color: theme.panelText}}>{e.text}</span>
                        </div>
                        {(e.lines ?? []).map((l, j) => (
                          <div key={j} style={{display: 'flex', gap: 10, paddingLeft: 12}}>
                            <span style={{color: withAlpha(theme.panelMuted, 0.5)}}>⎿</span>
                            <span style={{color: theme.panelMuted}}>{l}</span>
                          </div>
                        ))}
                      </div>
                    ) : null}

                    {/* The working line. Its own glyph, and the elapsed time and
                        token count that make a transcript feel live. */}
                    {e.kind === 'spin' ? (
                      <div style={{display: 'flex', gap: 12, alignItems: 'baseline'}}>
                        <span style={{color: theme.accent}}>✻</span>
                        <span style={{color: withAlpha(theme.panelText, 0.9)}}>{e.text}</span>
                        {e.mark ? <span style={{color: dim}}>{e.mark}</span> : null}
                      </div>
                    ) : null}

                    {e.kind === 'menu' ? (
                      <div>
                        <div style={{color: theme.panelText, marginBottom: 10}}>{e.text}</div>
                        {(e.options ?? []).map((o, j) => {
                          const on = j === (e.pick ?? 0);
                          const oa = isNew ? arrive(from + 260 + j * 140, 260) : 1;
                          return (
                            <div
                              key={j}
                              style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 12,
                                height: LINE_H + 8,
                                paddingLeft: 10,
                                marginLeft: -10,
                                borderRadius: 6,
                                background: on ? withAlpha(theme.accent, 0.13) : 'transparent',
                                borderLeft: `2px solid ${on ? theme.accent : 'transparent'}`,
                                opacity: oa,
                              }}
                            >
                              <span style={{color: on ? theme.accent : 'transparent', width: 16}}>❯</span>
                              <span style={{color: dim}}>{j + 1}.</span>
                              <span style={{color: on ? theme.panelText : theme.panelMuted}}>{o}</span>
                            </div>
                          );
                        })}
                        {e.track ? (
                          <div style={{position: 'relative', height: 54, marginTop: 16}}>
                            <div
                              style={{
                                position: 'absolute',
                                top: 26,
                                left: 0,
                                right: 0,
                                height: 2,
                                background: withAlpha(theme.panelMuted, 0.35),
                              }}
                            />
                            <div
                              style={{
                                position: 'absolute',
                                top: 18,
                                left: `${interpolate(ms, [from + 500, from + 1400], [82, 34], {
                                  extrapolateLeft: 'clamp',
                                  extrapolateRight: 'clamp',
                                })}%`,
                                width: 3,
                                height: 18,
                                background: theme.accent,
                                boxShadow: `0 0 14px ${withAlpha(theme.accent, 0.6)}`,
                              }}
                            />
                          </div>
                        ) : null}
                        {e.mark ? (
                          <div
                            style={{
                              color: theme.accent,
                              fontSize: BODY_FS + 1,
                              marginTop: 14,
                              opacity: isNew ? arrive(from + 700, 300) : 1,
                            }}
                          >
                            {e.mark}
                          </div>
                        ) : null}
                        {e.foot ? (
                          <div style={{color: withAlpha(theme.panelMuted, 0.6), marginTop: 12}}>{e.foot}</div>
                        ) : null}
                      </div>
                    ) : null}
                  </div>
                );
              })}
            </div>

            <div
              style={{
                position: 'absolute',
                left: 0,
                right: 0,
                top: 0,
                height: 54,
                background: `linear-gradient(${theme.panel}, ${withAlpha(theme.panel, 0)})`,
                opacity: scroll > 2 ? 1 : 0,
              }}
            />
          </div>

          {/* The composer, and the mode line: the furniture that says "in use". */}
          <div style={{flexShrink: 0, padding: `0 ${PAD}px ${PAD - 12}px`, position: 'relative'}}>
            {props.branch ? (
              <div
                style={{
                  position: 'absolute',
                  right: PAD,
                  top: -14,
                  fontFamily: mono,
                  fontSize: 16,
                  color: theme.panel,
                  background: theme.accent,
                  padding: '2px 10px',
                  borderRadius: 4,
                }}
              >
                {props.branch}
              </div>
            ) : null}
            <div
              style={{
                border: `1px solid ${withAlpha(theme.panelText, 0.16)}`,
                borderRadius: 10,
                padding: '14px 18px',
                display: 'flex',
                alignItems: 'center',
                gap: 12,
                fontFamily: mono,
                fontSize: BODY_FS,
                background: withAlpha(theme.panelText, 0.03),
              }}
            >
              <span style={{color: theme.accent}}>&gt;</span>
              <span style={{color: withAlpha(theme.panelMuted, 0.75), flex: 1}}>{lastAsk ? lastAsk.text : ''}</span>
              <span style={{width: 11, height: 22, background: caretOn ? theme.accent : 'transparent'}} />
            </div>
            <div
              style={{
                fontFamily: mono,
                fontSize: 17,
                color: withAlpha(theme.panelMuted, 0.62),
                marginTop: 12,
                display: 'flex',
                gap: 14,
              }}
            >
              {props.hint ? <span style={{color: withAlpha(theme.accent, 0.85)}}>▸▸ {props.hint}</span> : null}
              {props.pr ? <span>· {props.pr} ·</span> : null}
              <span style={{flex: 1}} />
              {step?.chip && props.chip ? (
                <span
                  style={{
                    color: theme.accent,
                    opacity: arrive(steps.find((s) => s.show === 'chip')?.startMs ?? 0, 400),
                  }}
                >
                  {props.chip}
                </span>
              ) : null}
              {props.status ? <span>{props.status}</span> : null}
            </div>
          </div>
        </div>

        {/* The margin. Notes live outside the window: inside, a note is something
            the tool printed, which is a lie about what it is. */}
        <div style={{width: hasNotes ? NOTE_W : 0, paddingTop: CHROME_H + PAD}}>
          {notes.map((e, i) => {
            const at = events.indexOf(e);
            const s = steps.find((x) => x.show === 'event' && x.at === at);
            return (
              <div
                key={i}
                style={{
                  fontFamily: mono,
                  fontSize: 19,
                  lineHeight: 1.45,
                  color: theme.accentText,
                  marginBottom: 26,
                  opacity: arrive((s?.startMs ?? sceneStart) + 400, 420),
                }}
              >
                <span style={{opacity: 0.5}}>····&nbsp;</span>
                {e.note}
              </div>
            );
          })}
        </div>
      </div>
    </Stage>
  );
};
