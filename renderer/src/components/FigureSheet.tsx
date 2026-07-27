import {useCurrentFrame} from 'remotion';
import {AbsoluteFill} from 'remotion';
import {FPS} from '../types';
import {FIGURES, FIGURE_BOX} from './artwork';

// A contact sheet of every figure in the vocabulary, on one frame.
//
// This exists because the figures are the one part of the illustration
// template that cannot be checked by reading it: a path that is geometrically
// fine can still read as a letter, or vanish because it was painted in the
// shading colour on a dark stage. Both of those actually happened. Rendering
// the whole vocabulary at once catches them in a single still instead of
// eleven, and it keeps catching them when a palette token changes.
//
// It is a development composition, not a scene: nothing in the pipeline emits
// it and it carries no baseline.

const CELL = 300;
const COLS = 4;

export const FigureSheet: React.FC = () => {
  const frame = useCurrentFrame();
  const names = Object.keys(FIGURES);
  const palette = {
    accent: '#ffd43b',
    primary: '#4f8fd0',
    ink: '#0a1220',
    soft: '#dbe4f2',
    line: '#a2aec4',
  };

  return (
    <AbsoluteFill
      style={{
        background: '#0d1524',
        display: 'flex',
        flexWrap: 'wrap',
        alignContent: 'center',
        justifyContent: 'center',
        gap: 24,
        padding: 40,
      }}
    >
      {names.map((name) => {
        const Figure = FIGURES[name];
        return (
          <div key={name} style={{width: CELL, textAlign: 'center'}}>
            <svg width={CELL} height={CELL} viewBox={`0 0 ${FIGURE_BOX} ${FIGURE_BOX}`}>
              <Figure build={1} t={frame / FPS} palette={palette} />
            </svg>
            <div style={{color: '#a2aec4', fontFamily: 'monospace', fontSize: 22, marginTop: -18}}>
              {name}
            </div>
          </div>
        );
      })}
    </AbsoluteFill>
  );
};
