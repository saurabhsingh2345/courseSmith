import {
  Img,
  OffthreadVideo,
  Sequence,
  staticFile,
  useCurrentFrame,
  useVideoConfig,
  interpolate,
  Easing,
} from 'remotion';
import {assetPath} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {Stage, STAGE_H} from './Stage';
import {CaptureCredit, Credit} from './CaptureCredit';

// FootageScene shows real captured frames of somebody else's product, inside
// browser chrome, pushing in on the part of the frame the shot was about.
//
// The chrome is not decoration. A bare screenshot on a stage reads as an image
// in a slide deck; the same pixels behind an address bar read as a thing that
// happened on a computer, which is the entire reason this footage was captured
// rather than drawn. The address shown is the clip's recorded origin — the one
// the driver wrote, never a caption somebody typed.
//
// Frames carry no duration of their own. The scene divides its screen time
// evenly, so a take with three shots holds each for a third. That keeps timing
// a property of the narration rather than of how many screenshots a take
// happened to grab.

const CHROME_H = 64;
const PUSH_SCALE = 1.18;

type Focus = {x: number; y: number; w: number; h: number};
type Frame = {path: string; mark?: string; focus?: Focus | null};
type Segment = {fromMs: number; toMs: number; rate: number};

const FILL: React.CSSProperties = {
  width: '100%',
  height: '100%',
  objectFit: 'cover',
  objectPosition: 'top center',
};

// Chrome is the browser window every web capture sits inside.
//
// It is not decoration. A bare screenshot on a stage reads as an image in a
// slide deck; the same pixels behind an address bar read as something that
// happened on a computer, which is the whole reason this footage was captured
// rather than drawn. The address is the clip's recorded origin — written by the
// driver, never a caption somebody typed.
const Chrome: React.FC<{
  theme: ResolvedTheme;
  title: string;
  credit?: Credit;
  children: React.ReactNode;
}> = ({theme, title, credit, children}) => (
  <Stage>
    <div
      style={{
        position: 'relative',
        borderRadius: 18,
        overflow: 'hidden',
        boxShadow: '0 44px 110px rgba(0, 0, 0, 0.55)',
        border: `1px solid ${theme.surfaceBorder}`,
        width: 1520,
        maxWidth: '100%',
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 12,
          padding: '16px 22px',
          backgroundColor: '#1f2430',
          height: CHROME_H,
          boxSizing: 'border-box',
        }}
      >
        {['#ff5f57', '#febc2e', '#28c840'].map((c) => (
          <div key={c} style={{width: 16, height: 16, borderRadius: 8, backgroundColor: c}} />
        ))}
        <div
          style={{
            flex: 1,
            marginLeft: 10,
            padding: '6px 18px',
            borderRadius: 999,
            backgroundColor: '#12161f',
            color: '#9aa4b2',
            fontSize: 20,
            fontFamily: theme.fontMono,
            whiteSpace: 'nowrap',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
          }}
        >
          {title}
        </div>
      </div>
      <div
        style={{
          position: 'relative',
          overflow: 'hidden',
          height: Math.min(820, STAGE_H - CHROME_H),
          backgroundColor: '#0b0d12',
        }}
      >
        {children}
      </div>
      <CaptureCredit theme={theme} credit={credit} />
    </div>
  </Stage>
);

// pushTransform moves the frame toward its focus box over the shot's life.
//
// The move is small and slow on purpose: this is a still, and anything faster
// than a drift reads as a camera move nobody asked for. With no focus box the
// frame holds — a push toward the middle of a screenshot is motion for its own
// sake, and the honest thing is to keep still.
const pushTransform = (focus: Focus | null | undefined, t: number): string => {
  if (!focus) return 'none';
  const cx = focus.x + focus.w / 2;
  const cy = focus.y + focus.h / 2;
  const scale = interpolate(t, [0, 1], [1, PUSH_SCALE], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
    easing: Easing.inOut(Easing.ease),
  });
  // Translate so the focus centre stays put as the frame grows around it.
  const dx = (0.5 - cx) * (scale - 1) * 100;
  const dy = (0.5 - cy) * (scale - 1) * 100;
  return `scale(${scale}) translate(${dx}%, ${dy}%)`;
};

export const FootageScene: React.FC<{
  theme: ResolvedTheme;
  assetBase?: string;
  sceneStartMs: number;
  durationInFrames: number;
  props: Record<string, unknown>;
}> = ({theme, assetBase, durationInFrames, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const frames = (props.frames as Frame[] | undefined) ?? [];
  const origin = String(props.origin ?? '');
  const title = String(props.title ?? origin);
  const src = String(props.src ?? '');
  const credit = props.provenance as Credit | undefined;
  // Present only when the clip runs longer than its slot; see PlanTerminalPacing.
  const segments = (props.segments as Segment[] | undefined) ?? [];

  // A recorded clip — an app assembling itself, a deploy going green. Paced the
  // same way a terminal capture is, because it runs long for the same reason.
  if (src) {
    return (
      <Chrome theme={theme} title={title} credit={credit}>
        {segments.length > 0 ? (
          (() => {
            let atFrame = 0;
            return segments.map((seg, i) => {
              const lenFrames = Math.max(
                1,
                Math.round((((seg.toMs - seg.fromMs) / seg.rate) / 1000) * fps),
              );
              const from = atFrame;
              atFrame += lenFrames;
              return (
                <Sequence key={i} from={from} durationInFrames={lenFrames} layout="none">
                  <OffthreadVideo
                    src={staticFile(assetPath(assetBase, src))}
                    startFrom={Math.round((seg.fromMs / 1000) * fps)}
                    endAt={Math.round((seg.toMs / 1000) * fps)}
                    playbackRate={seg.rate}
                    style={FILL}
                    muted
                  />
                </Sequence>
              );
            });
          })()
        ) : (
          <OffthreadVideo src={staticFile(assetPath(assetBase, src))} style={FILL} muted />
        )}
      </Chrome>
    );
  }

  if (frames.length === 0) {
    return (
      <Stage>
        <div style={{color: theme.accent, fontSize: 32}}>capture unavailable</div>
      </Stage>
    );
  }

  // Which frame is on screen, and how far through its own slice we are.
  const total = Math.max(durationInFrames, 1);
  const per = total / frames.length;
  const index = Math.min(frames.length - 1, Math.floor(frame / per));
  const current = frames[index];
  const t = Math.min(1, Math.max(0, (frame - index * per) / per));

  // Each frame fades in over a few frames so a cut between stills is a cut and
  // not a flicker.
  const fadeFrames = Math.min(6, per / 3);
  const opacity = interpolate(frame - index * per, [0, fadeFrames], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  return (
    <Chrome theme={theme} title={title} credit={credit}>
      <Img
        src={staticFile(assetPath(assetBase, current.path))}
        style={{
          ...FILL,
          opacity,
          transform: pushTransform(current.focus, t),
          transformOrigin: 'center center',
        }}
      />
    </Chrome>
  );
};
