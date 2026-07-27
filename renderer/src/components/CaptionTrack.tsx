import {useMemo} from 'react';
import {AbsoluteFill, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {CaptionWord, FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';

// CaptionTrack renders karaoke captions as short pages (3–5 words), the
// education-tuned middle ground between full-line subtitles and TikTok
// word-flashing: each page pops in once, the active word springs up in the
// accent colour, and LLM-marked keywords keep the accent even after being
// spoken. Word indices are global (into the captions array) so the emphasis
// pass can address words stably.

const PAGE_MAX_WORDS = 5;
const PAGE_MAX_CHARS = 28;
const PAGE_BREAK_GAP_MS = 800;

type PageWord = CaptionWord & {index: number};
type Page = {words: PageWord[]; startMs: number; endMs: number};

const groupPages = (words: CaptionWord[]): Page[] => {
  const pages: Page[] = [];
  let current: PageWord[] = [];
  let chars = 0;
  const flush = () => {
    if (current.length > 0) {
      pages.push({
        words: current,
        startMs: current[0].startMs,
        endMs: current[current.length - 1].endMs,
      });
      current = [];
      chars = 0;
    }
  };
  words.forEach((w, i) => {
    if (current.length > 0) {
      const prev = current[current.length - 1];
      const sentenceEnd = /[.!?]$/.test(prev.word);
      if (
        sentenceEnd ||
        w.startMs - prev.endMs > PAGE_BREAK_GAP_MS ||
        current.length >= PAGE_MAX_WORDS ||
        chars + 1 + w.word.length > PAGE_MAX_CHARS
      ) {
        flush();
      }
    }
    current.push({...w, index: i});
    chars = chars === 0 ? w.word.length : chars + 1 + w.word.length;
  });
  flush();
  return pages;
};

export const CaptionTrack: React.FC<{
  theme: ResolvedTheme;
  captions: CaptionWord[];
  /** Global caption-word indices to keep emphasized (accent) permanently. */
  emphasis?: number[];
}> = ({theme, captions, emphasis}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const nowMs = (frame / FPS) * 1000;

  const pages = useMemo(() => groupPages(captions), [captions]);
  const emphasized = useMemo(() => new Set(emphasis ?? []), [emphasis]);
  if (pages.length === 0) {
    return null;
  }
  // Keep the page visible briefly past its last word.
  const page = pages.find((p) => nowMs >= p.startMs && nowMs <= p.endMs + 350);
  if (!page) {
    return null;
  }

  // Page entrance: a quick settle, anchored to the page's own start.
  const pageStartFrame = Math.round((page.startMs / 1000) * FPS);
  const pop = spring({
    frame: frame - pageStartFrame,
    fps,
    config: {damping: 30, stiffness: 320, mass: 0.6},
    durationInFrames: 8,
  });

  return (
    <AbsoluteFill style={{justifyContent: 'flex-end', alignItems: 'center', pointerEvents: 'none'}}>
      <div
        style={{
          marginBottom: 64,
          // The caption panel is glass over whatever the stage is, so both its
          // fill and its hairline come from the mode's own tokens. Hardcoding
          // a near-black pill put a dark slab across the bottom of every
          // paper-mode frame.
          backgroundColor: theme.mode === 'light' ? `${theme.surface}e6` : `${theme.ink}b8`,
          border: `1px solid ${theme.mode === 'light' ? `${theme.ink}1f` : `${theme.text}14`}`,
          backdropFilter: 'blur(12px)',
          borderRadius: 18,
          padding: '20px 40px',
          maxWidth: 1500,
          opacity: pop,
          transform: `scale(${0.94 + 0.06 * pop}) translateY(${(1 - pop) * 10}px)`,
          display: 'flex',
          // The gap must be em-of-the-caption-font (the container would
          // otherwise inherit ~16px and the words visually touch), plus
          // headroom for the active word's 1.06x pop.
          fontSize: 42,
          gap: '0.42em',
          alignItems: 'baseline',
          flexWrap: 'wrap',
          justifyContent: 'center',
        }}
      >
        {page.words.map((w) => {
          const active = nowMs >= w.startMs && nowMs < w.endMs;
          const spoken = nowMs >= w.endMs;
          const isKey = emphasized.has(w.index);
          const wordStartFrame = Math.round((w.startMs / 1000) * FPS);
          const wordSpring = spring({
            frame: frame - wordStartFrame,
            fps,
            config: {damping: 16, stiffness: 420, mass: 0.5},
            durationInFrames: 9,
          });
          const scale = active ? 1 + 0.06 * wordSpring : 1;
          const color = isKey
            ? theme.accentText
            : active
              ? theme.accentText
              : spoken
                ? theme.text
                : `${theme.textMuted}9e`;
          return (
            <span
              key={w.index}
              style={{
                fontFamily: theme.fontBody,
                fontSize: 42,
                lineHeight: 1.35,
                fontWeight: isKey || active ? 800 : 600,
                color,
                display: 'inline-block',
                transform: `scale(${scale})`,
                transformOrigin: 'center bottom',
                // Lift off the glass on the dark stage; on paper the same
                // shadow only smudges dark text.
                textShadow: theme.mode === 'light' ? 'none' : '0 2px 14px rgba(0,0,0,0.45)',
              }}
            >
              {w.word}
            </span>
          );
        })}
      </div>
    </AbsoluteFill>
  );
};
