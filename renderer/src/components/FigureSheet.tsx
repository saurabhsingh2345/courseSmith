import {useCurrentFrame} from 'remotion';
import {AbsoluteFill} from 'remotion';
import {FPS} from '../types';
import {FIGURES, FIGURE_BOX} from './artwork';
import {
  CASTS,
  Character,
  EXPRESSION_NAMES,
  castFor,
  castPaletteFor,
  POSE_NAMES,
  aspectFor,
  faceByName,
  poseByName,
  viewBoxFor,
} from './cast';

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

const CELL = 148;

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
        gap: 6,
        padding: 20,
      }}
    >
      {names.map((name) => {
        const Figure = FIGURES[name];
        return (
          <div key={name} style={{width: CELL, textAlign: 'center'}}>
            <svg width={CELL} height={CELL} viewBox={`0 0 ${FIGURE_BOX} ${FIGURE_BOX}`}>
              <Figure build={1} t={frame / FPS} palette={palette} />
            </svg>
            <div style={{color: '#a2aec4', fontFamily: 'monospace', fontSize: 15, marginTop: -22}}>
              {name}
            </div>
          </div>
        );
      })}
    </AbsoluteFill>
  );
};

// The same idea for the character: every pose on one frame, and every
// expression on the next.
//
// A character fails differently from a figure. The artwork is always "valid" —
// it is somebody else's drawing and it renders — so the failures are ones only
// an eye catches: a head that has drifted off the hand resting against it, a
// pose whose arms are clipped by the frame it was given, a face that says
// something other than the word it is filed under. None of those are things a
// test can assert, and all of them are obvious in one still.
const CAST_CELL_H = 230;

export const CastSheet: React.FC = () => {
  const frame = useCurrentFrame();
  const t = frame / FPS;
  // The dark stage's tokens, since that is where the character's colours are
  // hardest — see castPaletteFor.
  // The dark stage's tokens, since that is where the character's colours are
  // hardest — see castPaletteFor.
  const theme = {ink: '#16202e', primary: '#4f8fd0'};

  const cell = (
    label: string,
    poseName: string,
    expression: string,
    presenter = castFor('sheet'),
  ) => {
    const pose = poseByName(poseName);
    const palette = castPaletteFor(theme, presenter);
    return (
      <div key={label} style={{textAlign: 'center'}}>
        <svg
          width={CAST_CELL_H * aspectFor(pose)}
          height={CAST_CELL_H}
          viewBox={viewBoxFor(pose)}
        >
          <Character
            pose={pose}
            face={faceByName(expression)}
            character={presenter}
            palette={palette}
            t={t}
            seed={label}
          />
        </svg>
        <div style={{color: '#a2aec4', fontFamily: 'monospace', fontSize: 19}}>{label}</div>
      </div>
    );
  };

  return (
    <AbsoluteFill
      style={{
        background: '#0d1524',
        display: 'flex',
        flexWrap: 'wrap',
        alignContent: 'center',
        justifyContent: 'center',
        gap: 10,
        padding: 24,
      }}
    >
      {POSE_NAMES.map((name) => cell(name, name, 'neutral'))}
      {EXPRESSION_NAMES.map((name) => cell(name, 'idle', name))}
      {/* And the cast itself: one idle each, so it is obvious at a glance
          whether two presenters actually read as two people. */}
      {CASTS.map((c, i) => cell(`cast ${i}`, 'idle', 'happy', c))}
    </AbsoluteFill>
  );
};
