import {OffthreadVideo, staticFile} from 'remotion';
import {assetPath} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {Stage, STAGE_H} from './Stage';

// TerminalScene plays a VHS demo recording inside a styled terminal window
// frame (title bar, traffic lights, drop shadow).
//
// The recording's aspect ratio is whatever VHS produced, so the window height
// is capped against the stage box rather than trusted — a tall demo would
// otherwise run down into the caption band.
const TITLE_BAR_H = 62;

export const TerminalScene: React.FC<{
  theme: ResolvedTheme;
  assetBase?: string;
  props: Record<string, unknown>;
}> = ({theme, assetBase, props}) => {
  const src = String(props.src ?? '');
  const title = String(props.title ?? 'Terminal');
  const videoMaxH = STAGE_H - TITLE_BAR_H;

  return (
    <Stage>
      <div
        style={{
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
        {src ? (
          <OffthreadVideo
            src={staticFile(assetPath(assetBase, src))}
            style={{
              width: '100%',
              display: 'block',
              maxHeight: videoMaxH,
              objectFit: 'contain',
            }}
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
      </div>
    </Stage>
  );
};
