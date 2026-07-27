// The flat-vector object vocabulary for the illustration and story templates.
//
// These are drawn here rather than imported. The obvious alternative was to
// bundle a CC0 set (unDraw and friends), and it was rejected for a reason that
// only shows up once you try to animate one: a downloaded illustration is a
// single flat blob of paths, so a rocket's flame cannot flicker and a gear
// cannot turn. Owning the geometry means every figure has *parts*, and parts
// are what make this a video instead of a slideshow of clipart. It also means
// the figures speak the design system's palette instead of being recoloured
// towards it, and there is no third-party asset licence to keep track of.
//
// That trade is the opposite of the one cast.tsx makes for the character, and
// deliberately so. Nobody knows what a "correct" rocket looks like, so a built
// one is simply a rocket; everybody knows what a person looks like, and a
// person built from primitives is a stick figure forever. Objects are worth
// owning. People are worth importing.
//
// The figures themselves live in themed modules under artwork/ — the vocabulary
// is a hundred objects now, and one file holding all of them was a file nobody
// could find anything in. What stays here is the vocabulary itself: the map Go
// mirrors, and the lookup every scene goes through. The kit each figure is
// built from (the box, the two clocks, the staggered entrance) is artwork/kit.
//
// Anything imported from here re-exports from kit, so scenes keep importing
// artwork and never have to know how the drawer is organised.

export {FIGURE_BOX, type Figure, type FigurePalette, type FigureProps} from './artwork/kit';

import type {Figure} from './artwork/kit';
import * as tech from './artwork/tech';
import * as product from './artwork/product';
import * as science from './artwork/science';
import * as nature from './artwork/nature';
import * as work from './artwork/work';
import * as abstract from './artwork/abstract';
import * as learn from './artwork/learn';

/**
 * The closed figure vocabulary.
 *
 * Go owns the list (artFigureVocab in snippet_illustration.go) and normalizes
 * anything outside it to "spark"; a drift test keeps the two identical, because
 * a figure Go allows and this map does not have would render as an empty stage
 * and nobody would notice until they watched the clip.
 *
 * Ordered by theme rather than alphabetically, because the useful question when
 * adding one is "what is already near this idea", not "what starts with the
 * same letter".
 */
export const FIGURES: Record<string, Figure> = {
  // Software, infrastructure and the machines it runs on.
  server: tech.ServerFigure,
  database: tech.DatabaseFigure,
  terminal: tech.TerminalFigure,
  code: tech.CodeFigure,
  branch: tech.BranchFigure,
  bug: tech.BugFigure,
  package: tech.PackageFigure,
  api: tech.ApiFigure,
  pipeline: tech.PipelineFigure,
  cache: tech.CacheFigure,
  queue: tech.QueueFigure,
  container: tech.ContainerFigure,
  firewall: tech.FirewallFigure,
  cpu: tech.CpuFigure,
  memory: tech.MemoryFigure,
  disk: tech.DiskFigure,
  laptop: tech.LaptopFigure,
  phone: tech.PhoneFigure,
  monitor: tech.MonitorFigure,
  cloud: tech.CloudFigure,
  network: tech.NetworkFigure,
  lock: tech.LockFigure,

  // The web, and the things people do on it.
  browser: product.BrowserFigure,
  cursor: product.CursorFigure,
  search: product.SearchFigure,
  form: product.FormFigure,
  cart: product.CartFigure,
  tag: product.TagFigure,
  wallet: product.WalletFigure,
  receipt: product.ReceiptFigure,
  star: product.StarFigure,
  heart: product.HeartFigure,
  bell: product.BellFigure,
  envelope: product.EnvelopeFigure,
  chat: product.ChatFigure,
  megaphone: product.MegaphoneFigure,
  share: product.ShareFigure,
  video: product.VideoFigure,

  // Measuring, testing and looking closely.
  atom: science.AtomFigure,
  flask: science.FlaskFigure,
  microscope: science.MicroscopeFigure,
  telescope: science.TelescopeFigure,
  dna: science.DnaFigure,
  magnet: science.MagnetFigure,
  battery: science.BatteryFigure,
  prism: science.PrismFigure,
  balance: science.BalanceFigure,
  compass: science.CompassFigure,
  ruler: science.RulerFigure,
  calculator: science.CalculatorFigure,
  satellite: science.SatelliteFigure,
  gauge: science.GaugeFigure,
  stack: science.StackFigure,
  clock: science.ClockFigure,
  hourglass: science.HourglassFigure,

  // The world, weather and growing things.
  globe: nature.GlobeFigure,
  tree: nature.TreeFigure,
  leaf: nature.LeafFigure,
  mountain: nature.MountainFigure,
  sun: nature.SunFigure,
  moon: nature.MoonFigure,
  drop: nature.DropFigure,
  fire: nature.FireFigure,
  wind: nature.WindFigure,
  seed: nature.SeedFigure,
  wave: nature.WaveFigure,
  recycle: nature.RecycleFigure,

  // Desks, days and the things people aim at.
  book: work.BookFigure,
  notebook: work.NotebookFigure,
  pencil: work.PencilFigure,
  folder: work.FolderFigure,
  calendar: work.CalendarFigure,
  briefcase: work.BriefcaseFigure,
  coffee: work.CoffeeFigure,
  target: work.TargetFigure,
  trophy: work.TrophyFigure,
  flag: work.FlagFigure,
  key: work.KeyFigure,
  door: work.DoorFigure,
  handshake: work.HandshakeFigure,
  brain: work.BrainFigure,
  eye: work.EyeFigure,
  checklist: work.ChecklistFigure,
  lightbulb: work.LightbulbFigure,
  gears: work.GearsFigure,

  // Teaching, and the shapes a lesson is made of. These came last and are the
  // set this project most needed: the vocabulary was built for explaining
  // systems, while courseSmith mostly explains an idea to a beginner. There was
  // a figure for a load balancer and none for a question.
  question: learn.QuestionFigure,
  chalkboard: learn.ChalkboardFigure,
  insight: learn.InsightFigure,
  timer: learn.TimerFigure,
  certificate: learn.CertificateFigure,
  answer: learn.AnswerFigure,
  library: learn.LibraryFigure,
  highlighter: learn.HighlighterFigure,
  signpost: learn.SignpostFigure,
  foundation: learn.FoundationFigure,
  progress: learn.ProgressFigure,
  discussion: learn.DiscussionFigure,
  study: learn.StudyFigure,
  steps: learn.StepsFigure,
  graduate: learn.GraduateFigure,
  bookmark: learn.BookmarkFigure,

  // Shapes for ideas that are not objects.
  arrow: abstract.ArrowFigure,
  loop: abstract.LoopFigure,
  funnel: abstract.FunnelFigure,
  filter: abstract.FilterFigure,
  switch: abstract.SwitchFigure,
  slider: abstract.SliderFigure,
  puzzle: abstract.PuzzleFigure,
  ladder: abstract.LadderFigure,
  bridge: abstract.BridgeFigure,
  maze: abstract.MazeFigure,
  orbit: abstract.OrbitFigure,
  growth: abstract.GrowthFigure,
  rocket: abstract.RocketFigure,
  shield: abstract.ShieldFigure,
  chart: abstract.ChartFigure,
  spark: abstract.SparkFigure,
};

/** The figure for a vocabulary name, falling back to the neutral burst. */
export const figureFor = (name: string | undefined): Figure =>
  (name && FIGURES[name]) || FIGURES.spark;
