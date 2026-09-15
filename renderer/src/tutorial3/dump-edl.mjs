// Transpile edl.ts and print the beat list as JSON, so the narration builder
// and the renderer read the same single source of truth for the cut.
import {build} from 'esbuild';
import {writeFileSync, mkdtempSync} from 'fs';
import {tmpdir} from 'os';
import {join, dirname} from 'path';
import {fileURLToPath, pathToFileURL} from 'url';

const here = dirname(fileURLToPath(import.meta.url));
const out = join(mkdtempSync(join(tmpdir(), 'edl-')), 'edl.mjs');

await build({
  entryPoints: [join(here, 'edl.ts')],
  outfile: out,
  bundle: true,
  format: 'esm',
  platform: 'node',
  logLevel: 'error',
});

const {BEATS} = await import(pathToFileURL(out).href);
const target = process.argv[2] ?? join(here, 'edl.json');
writeFileSync(target, JSON.stringify(BEATS, null, 2));
console.error(`wrote ${BEATS.length} beats to ${target}`);
