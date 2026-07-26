import {useEffect, useMemo, useState} from 'react';
import {
  AbsoluteFill,
  continueRender,
  delayRender,
  staticFile,
  useCurrentFrame,
} from 'remotion';
import {assetPath} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_H, STAGE_W} from './Stage';

// Frames between consecutive group reveals.
const GROUP_STAGGER = 12;
const GROUP_FADE = 14;

/** Ceiling on diagram upscale — see the sizing note in the component. */
const MAX_UPSCALE = 4;

// DiagramScene loads a generated SVG and reveals its top-level <g> groups
// sequentially (opacity + translateY), so diagrams build up as the narrator
// talks instead of appearing all at once.
export const DiagramScene: React.FC<{
  theme: ResolvedTheme;
  assetBase?: string;
  props: Record<string, unknown>;
}> = ({theme, assetBase, props}) => {
  const frame = useCurrentFrame();
  const src = String(props.src ?? '');
  const title = String(props.title ?? '');

  const [svgText, setSvgText] = useState<string | null>(null);
  const [handle] = useState(() => delayRender('load-diagram-svg'));

  useEffect(() => {
    let cancelled = false;
    fetch(staticFile(assetPath(assetBase, src)))
      .then((res) => (res.ok ? res.text() : Promise.reject(new Error(`HTTP ${res.status}`))))
      .then((text) => {
        if (!cancelled) {
          setSvgText(text);
          continueRender(handle);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setSvgText('');
          continueRender(handle);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [assetBase, src, handle]);

  // Animate top-level groups by injecting per-frame styles into the markup,
  // and read the viewBox so the diagram can be scaled to fill the stage.
  const animated = useMemo(() => {
    if (!svgText) {
      return null;
    }
    const doc = new DOMParser().parseFromString(svgText, 'image/svg+xml');
    const root = doc.documentElement;
    if (root.nodeName === 'parsererror') {
      return {markup: svgText, vbW: 0, vbH: 0};
    }
    const groups = Array.from(root.children).filter((el) => el.tagName === 'g');
    groups.forEach((g, i) => {
      const start = i * GROUP_STAGGER;
      const p = Math.max(0, Math.min(1, (frame - start) / GROUP_FADE));
      (g as SVGGElement).setAttribute(
        'style',
        `opacity:${p};transform:translateY(${(1 - p) * 24}px)`,
      );
    });
    // Mermaid/D2 cap their own display size (inline max-width style); strip
    // it and let the viewBox scale the drawing to whatever box we give it.
    root.removeAttribute('style');
    let vbW = 0;
    let vbH = 0;
    const vb = root.getAttribute('viewBox');
    if (vb) {
      const parts = vb.trim().split(/[\s,]+/).map(Number);
      if (parts.length === 4 && parts[2] > 0 && parts[3] > 0) {
        vbW = parts[2];
        vbH = parts[3];
      }
    }
    if (!vbW || !vbH) {
      const w = parseFloat(root.getAttribute('width') ?? '');
      const h = parseFloat(root.getAttribute('height') ?? '');
      if (w > 0 && h > 0) {
        root.setAttribute('viewBox', `0 0 ${w} ${h}`);
        vbW = w;
        vbH = h;
      }
    }
    root.setAttribute('width', '100%');
    root.setAttribute('height', '100%');
    root.setAttribute('preserveAspectRatio', 'xMidYMid meet');
    return {markup: new XMLSerializer().serializeToString(root), vbW, vbH};
  }, [svgText, frame]);

  if (svgText === null) {
    return null;
  }
  if (svgText === '') {
    return (
      <AbsoluteFill style={{alignItems: 'center', justifyContent: 'center'}}>
        <div style={{color: theme.text, fontFamily: theme.fontBody, fontSize: 40}}>
          Diagram unavailable: {src}
        </div>
      </AbsoluteFill>
    );
  }

  // Diagrams are compiled with the course's dark theme (transparent canvas),
  // so they sit directly on the scene background — no framing card — and are
  // scaled to fill the stage box.
  //
  // The upscale used to be capped at 2.2x, which is what left mermaid output
  // stranded: a typical three-node flowchart has viewBox 666x96.5, so 2.2x
  // produced a 1465x212 ribbon floating in a 1080 frame with ~430px of dead
  // space above and below it. The cap now scales with how wide the drawing is
  // relative to its height — a short, wide graph is *allowed* to grow until it
  // spans the stage, while a tall one still can't balloon past legibility.
  const headerH = title ? 118 : 0;
  const boxMaxW = STAGE_W;
  const boxMaxH = STAGE_H - headerH;
  let boxW = boxMaxW;
  let boxH = boxMaxH;
  if (animated && animated.vbW > 0 && animated.vbH > 0) {
    const fit = Math.min(boxMaxW / animated.vbW, boxMaxH / animated.vbH);
    // Cap on absolute glyph growth, not on the fit itself: mermaid draws at
    // 18px type, so 4x still lands around a 72px label — large, not silly.
    const scale = Math.min(fit, MAX_UPSCALE);
    boxW = animated.vbW * scale;
    boxH = animated.vbH * scale;
  }
  return (
    <Stage>
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={38} />
      <div style={{position: 'relative'}}>
        {/* Soft accent pool behind the drawing so a wide, short diagram is
            anchored on the stage instead of floating in flat darkness. */}
        <div
          style={{
            position: 'absolute',
            left: '50%',
            top: '50%',
            width: boxW + 360,
            height: boxH + 300,
            transform: 'translate(-50%, -50%)',
            borderRadius: '50%',
            background: `radial-gradient(ellipse, ${theme.primary}26 0%, transparent 65%)`,
          }}
        />
        <div
          style={{
            position: 'relative',
            width: boxW,
            height: boxH,
            filter: 'drop-shadow(0 26px 60px rgba(0, 0, 0, 0.45))',
          }}
          // The SVG comes from our own visuals stage, which validates it
          // (well-formed, no scripts, no external refs) before it reaches disk.
          dangerouslySetInnerHTML={{__html: animated?.markup ?? ''}}
        />
      </div>
    </Stage>
  );
};
