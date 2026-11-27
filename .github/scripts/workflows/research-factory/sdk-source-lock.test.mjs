// Asserts the compiled research-factory lock stages the cloud-sdk-go source runner-side without
// gh-aw treating the workflow as a Go runtime request: no actions/setup-go and no Go proxy hosts in
// the firewall allowlists. Also pins the prompt, critic and rubric text that tells the agent where
// the staged source is, and the go.mod / go.sum lines the staging step resolves.
import assert from 'node:assert/strict';
import test, { describe } from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const REPO_ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..', '..', '..');
const read = (...parts) => fs.readFileSync(path.join(REPO_ROOT, ...parts), 'utf8');

const LOCK = read('.github', 'workflows', 'research-factory-issue.lock.yml');
const LINES = LOCK.split('\n');
const SOURCE = read('.github', 'workflows', 'research-factory-issue.md');
const RUBRIC = read('.github', 'scripts', 'workflows', 'research-factory', 'critic-rubric.md');
const GO_MOD = read('go.mod');
const GO_SUM = read('go.sum');

const STAGED_ROOT = '/tmp/gh-aw/gomodcache';
const SDK_DIR_PREFIX = STAGED_ROOT + '/github.com/elastic/cloud-sdk-go@';

const STAGE = 'Stage cloud-sdk-go source for the agent';
const CHECKOUT = 'Checkout repository (gh-aw default)';
const EXECUTE = 'Execute Claude Code CLI';
const UPLOAD = 'Upload agent artifacts';

const GO_HOSTS = ['proxy.golang.org', 'sum.golang.org', 'goproxy.io', 'golang.org', 'go.dev', 'pkg.go.dev', 'storage.googleapis.com'];
const AGENT_ALLOWLIST_SIZE = 37;
const DETECTION_ALLOWLIST = ['openrouter.ai', 'www.elastic.co'];

const JOB_LINE = /^ {2}[a-z_]+:$/;
// gh-aw emits custom steps with their keys sorted, so a step can begin `- continue-on-error:` and
// carry its name on a later line. Split on the list marker, then look for `name:` anywhere at the
// step's key indentation.
const STEP_START = /^ {6}- /;
const STEP_NAME = /^ {6}(?:- | {2})name: (.*)$/;

/** Job ids under `jobs:`. */
function jobIds() {
  const jobsStart = LINES.indexOf('jobs:');
  assert.notEqual(jobsStart, -1, 'the lock has no jobs: key');
  return LINES.slice(jobsStart + 1).filter((l) => JOB_LINE.test(l)).map((l) => l.trim().slice(0, -1));
}

/** Lines of one job: from `  <jobId>:` (under `jobs:`) to the next job line. */
function jobLines(jobId) {
  const jobsStart = LINES.indexOf('jobs:');
  assert.notEqual(jobsStart, -1, 'the lock has no jobs: key');
  const start = LINES.indexOf(`  ${jobId}:`, jobsStart);
  assert.notEqual(start, -1, `the lock has no ${jobId} job`);
  let end = start + 1;
  while (end < LINES.length && !JOB_LINE.test(LINES[end])) {
    end += 1;
  }
  return LINES.slice(start, end);
}

/** The steps of a job, each as `{ name, lines }`. */
function steps(jobId) {
  const found = [];
  for (const line of jobLines(jobId)) {
    if (STEP_START.test(line)) {
      found.push({ name: undefined, lines: [line] });
    } else if (found.length > 0) {
      found[found.length - 1].lines.push(line);
    }
  }
  for (const s of found) {
    const nameLine = s.lines.map((l) => STEP_NAME.exec(l)).find(Boolean);
    s.name = nameLine ? nameLine[1].trim() : undefined;
  }
  return found;
}

function step(jobId, name) {
  const matches = steps(jobId).filter((s) => s.name === name);
  assert.equal(matches.length, 1, `expected exactly one "${name}" step in ${jobId}, found ${matches.length}`);
  return matches[0];
}

/** The lines of a step's `run: |` block scalar, without their indentation. */
function blockScript(s) {
  const at = s.lines.findIndex((l) => /^ {8}run: \|$/.test(l));
  assert.notEqual(at, -1, `step "${s.name}" has no block-scalar run: |`);
  const body = [];
  for (const line of s.lines.slice(at + 1)) {
    if (line !== '' && !line.startsWith(' '.repeat(10))) break;
    body.push(line.slice(10));
  }
  return body;
}

/** One glob path segment as an anchored regex: `*` and `?` never cross `/`. */
function segmentRegex(segment) {
  const body = segment.replace(/[.+^${}()|\\]/g, '\\$&').replaceAll('*', '[^/]*').replaceAll('?', '[^/]');
  return new RegExp(`^${body}$`);
}

/**
 * Whether an upload path (a directory, file or glob) would ship anything under `root`. Compared
 * segment by segment: a `**` segment, or running out of path segments (a directory upload), covers
 * everything below; a path that continues past `root`'s depth, inside it, ships part of it.
 */
function coversPath(uploadPath, root) {
  const up = uploadPath.replace(/\/+$/, '').split('/');
  const rt = root.split('/');
  for (let i = 0; i < up.length; i += 1) {
    if (up[i] === '**') return true;
    if (i >= rt.length) return true;
    if (!segmentRegex(up[i]).test(rt[i])) return false;
  }
  return true;
}

/** Every `allowDomains` array in the lock, in file order (the detection one is JSON-escaped). */
function allowLists() {
  return [...LOCK.matchAll(/\\?"allowDomains\\?":\[(.*?)\]/g)].map((m) => JSON.parse(`[${m[1].replaceAll('\\"', '"')}]`));
}

/** The prompt body: everything after the frontmatter's closing `---` line. */
function promptBody() {
  const lines = SOURCE.split('\n');
  assert.equal(lines[0], '---', 'research-factory-issue.md does not start with frontmatter');
  const close = lines.indexOf('---', 1);
  assert.notEqual(close, -1, 'research-factory-issue.md frontmatter is not closed');
  return { frontmatter: lines.slice(1, close), body: lines.slice(close + 1).join('\n') };
}

describe('research-factory lock: cloud-sdk-go staging', () => {
  test('the agent job stages the SDK after checkout and before the agent runs, and no other job does', () => {
    const names = steps('agent').map((s) => s.name);
    step('agent', STAGE);
    const at = names.indexOf(STAGE);
    const checkout = names.indexOf(CHECKOUT);
    const execute = names.indexOf(EXECUTE);
    assert.notEqual(checkout, -1, `agent job has no "${CHECKOUT}" step`);
    assert.notEqual(execute, -1, `agent job has no "${EXECUTE}" step`);
    assert.ok(checkout < at && at < execute, `step order broken: ${names.join(' | ')}`);
    for (const job of jobIds().filter((j) => j !== 'agent')) {
      assert.ok(!steps(job).some((s) => s.name === STAGE), `job ${job} also has "${STAGE}"`);
    }
  });

  test('a staging failure fails the run, within a bounded time, and the step compiles to a block scalar', () => {
    const s = step('agent', STAGE);
    // Every source or no run: a missing SDK must be a red run, not a warning in a green one.
    assert.ok(
      !s.lines.some((l) => /^ {6}(?:- | {2})continue-on-error:/.test(l)),
      'staging step has continue-on-error; a failed staging must fail the run',
    );
    assert.ok(s.lines.some((l) => l === `          SDK_GOMODCACHE: ${STAGED_ROOT}`), `staging step lost SDK_GOMODCACHE: ${STAGED_ROOT}`);
    // Backstop above the script's worst case (4 x 60 s attempts + 155 s of back-off).
    assert.ok(s.lines.some((l) => /^ {6}(?:- | {2})timeout-minutes: 8$/.test(l)), 'staging step lost timeout-minutes: 8');
    assert.ok(
      !s.lines.some((l) => /^ {8}run: "/.test(l)),
      'staging step compiled to a one-line run: "..." (a # comment inside the script?); gh-aw then detects `go` and injects actions/setup-go',
    );
    blockScript(s);
  });

  test('the staging step exports GOMODCACHE and GOPROXY=off before it downloads anything', () => {
    const script = blockScript(step('agent', STAGE));
    const open = script.indexOf('{');
    const close = script.indexOf('} >> "$GITHUB_ENV"');
    const download = script.findIndex((l) => l.includes('go mod download'));
    assert.ok(open !== -1 && close !== -1 && download !== -1, `staging script changed shape:\n${script.join('\n')}`);
    assert.deepEqual(
      script.slice(open + 1, close).map((l) => l.trim()).sort(),
      ['echo "GOMODCACHE=${SDK_GOMODCACHE}"', 'echo "GOPROXY=off"'],
      'the { ... } >> "$GITHUB_ENV" group must write exactly GOMODCACHE and GOPROXY=off',
    );
    assert.ok(open < close && close < download, 'GOMODCACHE / GOPROXY=off must reach GITHUB_ENV before go mod download');
    // Without the in-step export, go mod download fills ~/go/pkg/mod, the step still prints
    // "staged", and the sandbox's GOMODCACHE points at an empty directory.
    const exported = script.indexOf('export GOMODCACHE="${SDK_GOMODCACHE}"');
    assert.ok(exported !== -1 && exported < download, 'the staging script must export GOMODCACHE before go mod download');
    // The in-script deadline turns a stalled download into an ordinary failure, so the cleanup
    // branch still runs; the step's timeout-minutes would kill the script before it.
    assert.match(script[download], /^\s*timeout \d+ go mod download /, 'go mod download must run under an in-script timeout');
    // go list sits outside the per-attempt timeout; with the proxy off it cannot block on the network.
    assert.ok(script.some((l) => /GOPROXY=off go list -m /.test(l)), 'go list must run with GOPROXY=off');
  });

  test('the staging step retries, then fails with an error annotation', () => {
    const script = blockScript(step('agent', STAGE));
    const loop = script.map((l) => /^for delay in ([\d ]+); do$/.exec(l)).find(Boolean);
    assert.ok(loop, `staging script has no retry loop:\n${script.join('\n')}`);
    assert.equal(loop[1].trim().split(/\s+/).length, 4, 'staging script must make four attempts');
    const tail = script.filter((l) => l.trim() !== '').slice(-2).map((l) => l.trim());
    assert.match(tail[0], /^echo "::error::could not stage /, 'staging script must end with an ::error:: annotation');
    assert.equal(tail[1], 'exit 1', 'staging script must exit 1 after the last failed attempt');
    // Success must leave the loop at once: exit 0 inside the `if stage; then` block.
    const ifStage = script.findIndex((l) => l.trim() === 'if stage; then');
    const fiAfter = script.findIndex((l, i) => i > ifStage && l.trim() === 'fi');
    const exit0 = script.findIndex((l) => l.trim() === 'exit 0');
    assert.ok(ifStage !== -1 && ifStage < exit0 && exit0 < fiAfter, 'staging script must exit 0 inside the if stage; then block');
    // Go makes module directories read-only: without the chmod, rm -rf fails and bash -e ends the
    // step on the first failed attempt, skipping the retries and the ::error:: hint.
    const loopAt = script.findIndex((l) => /^for delay in /.test(l));
    const chmod = script.findIndex((l, i) => i > loopAt && /^\s*chmod -R u\+w "\$\{SDK_GOMODCACHE\}"/.test(l));
    const rm = script.findIndex((l, i) => i > loopAt && /^\s*rm -rf "\$\{SDK_GOMODCACHE\}"$/.test(l));
    assert.ok(chmod !== -1 && rm !== -1 && fiAfter < chmod && chmod < rm, 'a failed attempt must chmod -R u+w, then rm -rf, the staged cache');
  });

  test('the staging step finishes inside its timeout-minutes backstop', () => {
    const s = step('agent', STAGE);
    const script = blockScript(s);
    const minutes = Number(s.lines.map((l) => /^ {6}(?:- | {2})timeout-minutes: (\d+)$/.exec(l)).find(Boolean)?.[1]);
    const perAttempt = Number(script.map((l) => /^\s*timeout (\d+) go mod download /.exec(l)).find(Boolean)?.[1]);
    const delays = script.map((l) => /^for delay in ([\d ]+); do$/.exec(l)).find(Boolean)?.[1].trim().split(/\s+/).map(Number);
    assert.ok(minutes > 0 && perAttempt > 0 && delays?.length > 0, 'could not read timeout-minutes, the per-attempt timeout or the delays');
    // One attempt per delay; 30 s of slack for go list, cleanup and process start-up.
    const worst = delays.length * perAttempt + delays.reduce((a, b) => a + b, 0) + 30;
    assert.ok(worst < minutes * 60, `worst case ${worst}s does not fit timeout-minutes: ${minutes}; a kill would skip the cleanup and the ::error:: line`);
    // set -o pipefail aborts sh before anything reaches GITHUB_ENV.
    assert.ok(step('agent', STAGE).lines.some((l) => l === '        shell: bash'), 'the staging step must run with shell: bash');
  });

  test('gh-aw did not inject a Go toolchain setup', () => {
    assert.ok(!LOCK.includes('actions/setup-go'), 'the lock uses actions/setup-go; see .github/aw/README.md (runtimes: go / comment lines in run:)');
    for (const job of jobIds()) {
      assert.ok(!steps(job).some((s) => s.name === 'Setup Go'), `job ${job} has a "Setup Go" step`);
    }
  });

  test('the firewall allowlists gained no Go proxy host', () => {
    const lists = allowLists();
    assert.equal(lists.length, 2, `expected two allowDomains arrays (agent, detection), found ${lists.length}`);
    const isGoHost = (h) => {
      const host = h.replace(/^\*?\./, '');
      return GO_HOSTS.some((g) => host === g || host.endsWith(`.${g}`) || g.endsWith(`.${host}`));
    };
    for (const list of lists) {
      const go = list.filter(isGoHost);
      assert.deepEqual(go, [], `allowDomains contains Go hosts: ${go.join(', ')}`);
    }
    const [agent, detection] = lists;
    assert.equal(
      agent.length,
      AGENT_ALLOWLIST_SIZE,
      'the agent allowlist changed size; if this is a gh-aw bump that changed the `defaults` set, review the new hosts and update the number',
    );
    assert.deepEqual([...detection].sort(), DETECTION_ALLOWLIST);
  });

  test('the staged module is not uploaded by any agent-job artifact step', () => {
    const uploads = steps('agent').filter((s) => s.lines.some((l) => /uses: actions\/upload-artifact@/.test(l)));
    assert.ok(uploads.some((s) => s.name === UPLOAD), `the agent job has no "${UPLOAD}" step`);
    for (const s of uploads) {
      const paths = [];
      const single = s.lines.map((l) => /^ {10}path: (?!\|)(.+)$/.exec(l)).find(Boolean);
      if (single) paths.push(single[1].trim());
      const at = s.lines.findIndex((l) => /^ {10}path: \|$/.test(l));
      if (at !== -1) {
        for (const line of s.lines.slice(at + 1)) {
          if (!line.startsWith(' '.repeat(12))) break;
          paths.push(line.trim());
        }
      }
      assert.ok(paths.length > 0, `"${s.name}" lists no paths`);
      for (const p of paths.filter((x) => !x.startsWith('!'))) {
        assert.ok(!coversPath(p, STAGED_ROOT), `"${s.name}" path ${p} covers ${STAGED_ROOT}`);
      }
    }
  });

  test('the compiled critic prompt names the staged SDK directory', () => {
    const s = step('agent', EXECUTE);
    const agents = s.lines.find((l) => l.includes('research-critic'));
    assert.ok(agents, `"${EXECUTE}" has no research-critic --agents text`);
    assert.ok(agents.includes(SDK_DIR_PREFIX), `the compiled critic prompt does not name ${SDK_DIR_PREFIX}`);
    assert.ok(!agents.includes('is not in this checkout'), 'the compiled critic prompt still says cloud-sdk-go is not in this checkout');
  });
});

describe('research-factory sources: cloud-sdk-go staging', () => {
  test('the prompt body points at the staged SDK and renders no template action', () => {
    const { body } = promptBody();
    assert.ok(body.includes(SDK_DIR_PREFIX), `the prompt body does not name ${SDK_DIR_PREFIX}`);
    assert.ok(!body.includes('no Go toolchain'), 'the prompt body still says there is no Go toolchain');
    const stray = body.replaceAll('${{', '').includes('{{');
    assert.ok(!stray, 'the prompt body contains {{ outside ${{ ... }}; gh-aw renders the body as a template');
  });

  test('the frontmatter does not declare a runtimes key', () => {
    const { frontmatter } = promptBody();
    assert.ok(
      !frontmatter.some((l) => /^runtimes:/.test(l)),
      'runtimes: in research-factory-issue.md frontmatter; runtimes: go adds the Go proxy hosts to the firewall allowlist',
    );
  });

  test('the critic rubric points at the staged SDK', () => {
    assert.ok(RUBRIC.includes(SDK_DIR_PREFIX), `critic-rubric.md does not name ${SDK_DIR_PREFIX}`);
    assert.ok(!RUBRIC.includes('You cannot open it.'), 'critic-rubric.md still says the SDK cannot be opened');
  });

  test('go.mod requires cloud-sdk-go and go.sum carries its module hash', () => {
    const match = /^\s*github\.com\/elastic\/cloud-sdk-go (v\S+)$/m.exec(GO_MOD);
    assert.ok(match, 'go.mod no longer requires github.com/elastic/cloud-sdk-go; the staging step would fail every research run');
    const version = match[1];
    const escaped = version.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const sumLine = new RegExp(`^github\\.com/elastic/cloud-sdk-go ${escaped} h1:\\S+$`, 'm');
    assert.ok(sumLine.test(GO_SUM), `go.sum has no module hash for github.com/elastic/cloud-sdk-go ${version}`);
  });
});
