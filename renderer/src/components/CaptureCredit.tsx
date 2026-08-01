import {ResolvedTheme} from '../theme/theme';

// CaptureCredit is the small line a capture states about itself.
//
// It exists because of what pacing introduced. Compressing 53 seconds of real
// agent work into a 21-second slot and saying nothing is a quiet
// misrepresentation of how long the tool took — in a course whose whole moat is
// "the tool really did that". Every other claim here is defended by a
// recording; the length of the recording cannot be the one claim allowed to
// drift.
//
// So a condensed clip says so, in frame, with both numbers. Nothing here is
// generated: the tool name comes from the engine's registry, the version was
// read from the binary at capture time, and the durations were measured.
//
// It is deliberately quiet — a footnote, not a badge. The viewer should be able
// to check it and otherwise ignore it.

export type Credit = {
  tool?: string;
  version?: string;
  realMs?: number;
  shownMs?: number;
};

const secs = (ms: number): string => {
  const s = Math.round(ms / 1000);
  if (s < 60) return `${s}s`;
  return `${Math.floor(s / 60)}m ${s % 60}s`;
};

export const creditLine = (c: Credit): string => {
  const parts: string[] = [];
  if (c.tool) parts.push(c.version ? `${c.tool} ${c.version}` : c.tool);
  // Only claim a compression that happened; the Go side decides that by
  // leaving shownMs unset otherwise.
  if (c.realMs && c.shownMs) {
    parts.push(`${secs(c.realMs)} real, shown in ${secs(c.shownMs)}`);
  } else if (c.realMs) {
    parts.push(`${secs(c.realMs)} real time`);
  }
  return parts.join('  ·  ');
};

export const CaptureCredit: React.FC<{theme: ResolvedTheme; credit?: Credit}> = ({
  theme,
  credit,
}) => {
  if (!credit) return null;
  const line = creditLine(credit);
  if (!line) return null;
  return (
    <div
      style={{
        position: 'absolute',
        right: 26,
        bottom: 18,
        padding: '7px 16px',
        borderRadius: 999,
        backgroundColor: 'rgba(8, 11, 17, 0.72)',
        border: '1px solid rgba(255,255,255,0.10)',
        color: '#9aa4b2',
        fontSize: 19,
        fontFamily: theme.fontMono,
        letterSpacing: 0.2,
        whiteSpace: 'nowrap',
      }}
    >
      {line}
    </div>
  );
};
