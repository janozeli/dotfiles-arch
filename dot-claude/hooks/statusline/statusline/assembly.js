// Output assembly — combines GSD line with custom extension lines
// Renders multi-line tree layout or plain single line.

const { DIM, RST } = require('./shared');
const { git, gitSegment } = require('./git');
const { envSegment } = require('./env');
const { loadUsageCache, usageSegment, saveUsage } = require('./usage');

// Line icons (Nerd Font)
const ICON_GSD   = '\uf4a5';
const ICON_ENV   = '\ue5ff';
const ICON_USAGE = '\uf080';
const ICON_GIT   = '\ue702';

/**
 * Render the full statusline output.
 * @param {string} gsdLine — pre-built GSD content (line 1)
 * @param {string} dir     — current working directory
 */
function render(gsdLine, dir) {
  const lines = [{ icon: ICON_GSD, content: gsdLine }];

  const root = git(dir, 'rev-parse', '--show-toplevel');

  const envSeg = envSegment(dir, root);
  if (envSeg) lines.push({ icon: ICON_ENV, content: envSeg });

  const { cache, needFetch } = loadUsageCache();
  const usageSeg = usageSegment(cache);
  if (usageSeg) lines.push({ icon: ICON_USAGE, content: usageSeg });

  const gitSeg = root ? gitSegment(dir) : '';
  if (gitSeg) lines.push({ icon: ICON_GIT, content: gitSeg });

  // Render: tree layout if multiple lines, plain if single
  if (lines.length === 1) {
    process.stdout.write(lines[0].content);
  } else {
    process.stdout.write(lines.map((l, i) => {
      const conn = i === lines.length - 1 ? '\u2570\u2500' : '\u251c\u2500';
      return `${DIM}${conn}${RST} ${l.icon}${DIM} \u2502 ${RST}${l.content}`;
    }).join('\n'));
  }

  // Deferred: refresh usage cache if stale (runs after output)
  if (needFetch) process.on('beforeExit', () => saveUsage());
}

module.exports = { render };
