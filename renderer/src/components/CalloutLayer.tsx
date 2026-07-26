import {ResolvedTheme} from '../theme/theme';
import {AbsoluteFill, spring, useCurrentFrame, useVideoConfig, interpolate} from 'remotion';
import {Callout, Theme, msToFrame} from '../types';

// CalloutLayer draws arrow/circle/label overlays at percentage coordinates,
// animated in and out at word-level timestamps.
export const CalloutLayer: React.FC<{
  theme: ResolvedTheme;
  callouts: Callout[];
  sceneStartMs: number;
}> = ({theme, callouts, sceneStartMs}) => {
  const frame = useCurrentFrame();
  const {fps, width, height} = useVideoConfig();

  if (callouts.length === 0) {
    return null;
  }
  return (
    <AbsoluteFill style={{pointerEvents: 'none'}}>
      {callouts.map((c, i) => {
        const startFrame = msToFrame(Math.max(0, c.atMs - sceneStartMs));
        const endFrame = startFrame + msToFrame(c.durMs);
        if (frame < startFrame || frame > endFrame + 10) {
          return null;
        }
        const enter = spring({
          frame: frame - startFrame,
          fps,
          config: {damping: 14, stiffness: 160},
        });
        const exit = interpolate(frame, [endFrame - 8, endFrame + 8], [1, 0], {
          extrapolateLeft: 'clamp',
          extrapolateRight: 'clamp',
        });
        const opacity = Math.min(enter, exit);
        const cx = c.x * width;
        const cy = c.y * height;

        return (
          <div key={i} style={{position: 'absolute', left: 0, top: 0, opacity}}>
            {c.shape === 'circle' ? (
              <svg
                style={{position: 'absolute', left: cx - 90, top: cy - 90}}
                width={180}
                height={180}
              >
                <circle
                  cx={90}
                  cy={90}
                  r={70 * enter}
                  fill="none"
                  stroke={theme.accent}
                  strokeWidth={8}
                />
              </svg>
            ) : (
              <svg
                style={{position: 'absolute', left: cx - 160, top: cy - 160}}
                width={180}
                height={180}
              >
                {/* Arrow pointing down-right at the target. */}
                <g
                  stroke={theme.accent}
                  strokeWidth={10}
                  fill="none"
                  strokeLinecap="round"
                  transform={`scale(${enter})`}
                  style={{transformOrigin: '160px 160px'}}
                >
                  <line x1={30} y1={30} x2={150} y2={150} />
                  <polyline points="150,90 150,150 90,150" fill="none" />
                </g>
              </svg>
            )}
            <div
              style={{
                position: 'absolute',
                left: c.shape === 'circle' ? cx + 100 : cx - 170,
                top: c.shape === 'circle' ? cy - 24 : cy - 220,
                backgroundColor: theme.accent,
                color: '#101828',
                fontSize: 30,
                fontWeight: 700,
                padding: '10px 22px',
                borderRadius: 12,
                whiteSpace: 'nowrap',
                boxShadow: '0 10px 30px rgba(15, 23, 42, 0.25)',
                transform: `scale(${enter})`,
              }}
            >
              {c.label}
            </div>
          </div>
        );
      })}
    </AbsoluteFill>
  );
};
