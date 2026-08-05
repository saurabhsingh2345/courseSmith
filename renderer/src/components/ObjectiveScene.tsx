import {ResolvedTheme} from '../theme/theme';
import {CourseListScene, CourseRow} from './CourseListScene';

// The lesson's contract: what you will be able to DO, and the evidence.
//
// Each outcome carries its proof on the second line, and that pairing is the
// whole template — an outcome nobody can check is the thing this exists to
// refuse. The proof is tagged PROOF rather than left as prose so the eye can
// tell the promise from the way you'd settle it.

type Outcome = {action: string; evidence: string};

export const ObjectiveScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const outcomes = (Array.isArray(props.outcomes) ? props.outcomes : []) as Outcome[];
  const audience = String(props.audience ?? '');
  return (
    <CourseListScene
      theme={theme}
      sceneStartMs={sceneStartMs}
      props={props}
      count={outcomes.length}
      lead={audience ? `For you, if ${audience}` : undefined}
      row={(i, weight, focused) => (
        <CourseRow
          theme={theme}
          weight={weight}
          focused={focused}
          accent={theme.accentQuantity}
          primary={outcomes[i]?.action ?? ''}
          secondary={outcomes[i]?.evidence}
          tag="proof"
        />
      )}
    />
  );
};
