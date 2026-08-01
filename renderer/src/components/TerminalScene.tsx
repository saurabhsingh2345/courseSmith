import {OffthreadVideo, Sequence, staticFile, useVideoConfig} from 'remotion';
import {assetPath} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {Stage, STAGE_H} from './Stage';
import {CaptureCredit, Credit} from './CaptureCredit';

// TerminalScene plays a VHS demo recording inside a styled terminal window
// frame (title bar, traffic lights, drop shadow).
//
// The recording's aspect ratio is whatever VHS produced, so the window height
// is capped against the stage box rather than trusted — a tall demo would
// otherwise run down into the caption band.
//
// == Fitting a long recording into a short slot ==
//
// A scene's length comes from the narration, and a capture's length comes from
// how long a real tool really took. Those two numbers have no reason to match:
// the first real agent capture was 53s of recording in a 21s slot, so the video
// cut away ten seconds in and the viewer never saw the result — the whole point
// of the shot.
//
// Playing it faster everywhere would make the typing look silly. What the clip
// actually contains is a few moments separated by dead air, and the footage
// sidecar knows where those moments are: the marks. So the scene compresses the
// gaps and leaves the moments at real time, which is what the marks were built
// for and the first thing to actually use them.
//
// The plan itself is made in Go and written into the scene graph, the same way
// the motion tokens are: `lesson-video.json` records what was decided, a Go
// test checks the arithmetic, and this component only plays what it is told.
const TITLE_BAR_H = 62;

type Segment = {fromMs: number; toMs: number; rate: number};

export const TerminalScene: React.FC<{
  theme: ResolvedTheme;
  assetBase?: string;
  durationInFrames?: number;
  props: Record<string, unknown>;
}> = ({theme, assetBase, durationInFrames, props}) => {
  const src = String(props.src ?? '');
  const title = String(props.title ?? 'Terminal');
  const videoMaxH = STAGE_H - TITLE_BAR_H;
  const {fps} = useVideoConfig();

  // Absent for a clip that fits its slot, which is every ordinary demo.
  const segments = (props.segments as Segment[] | undefined) ?? [];
  const paced = segments.length > 0;

  const videoStyle: React.CSSProperties = {
    width: '100%',
    display: 'block',
    maxHeight: videoMaxH,
    objectFit: 'contain',
  };

  return (
    <Stage>
      <div
        style={{
          position: 'relative',
          borderRadius: 18,
          overflow: 'hidden',
          boxShadow: '0 44px 110px rgba(0, 0, 0, 0.55)',
          border: `1px solid ${theme.surfaceBorder}`,
          width: 1440,
          maxWidth: '100%',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            padding: '16px 22px',
            backgroundColor: '#1f2430',
          }}
        >
          {['#ff5f57', '#febc2e', '#28c840'].map((c) => (
            <div key={c} style={{width: 18, height: 18, borderRadius: 9, backgroundColor: c}} />
          ))}
          <div
            style={{
              flex: 1,
              textAlign: 'center',
              color: '#9aa4b2',
              fontSize: 24,
              fontFamily: theme.fontMono,
              whiteSpace: 'nowrap',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              paddingRight: 64, // balance the traffic lights
            }}
          >
            {title}
          </div>
        </div>
        {src && paced ? (
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
                    style={videoStyle}
                    // Demos are silent; the lesson voiceover carries the audio.
                    muted
                  />
                </Sequence>
              );
            });
          })()
        ) : src ? (
          <OffthreadVideo
            src={staticFile(assetPath(assetBase, src))}
            style={videoStyle}
            // Demos are silent; the lesson voiceover carries the audio.
            muted
          />
        ) : (
          <div
            style={{
              height: Math.min(700, videoMaxH),
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              backgroundColor: '#11151c',
              color: theme.accent,
              fontSize: 32,
            }}
          >
            demo recording unavailable
          </div>
        )}
        <CaptureCredit theme={theme} credit={props.provenance as Credit | undefined} />
      </div>
    </Stage>
  );
};
