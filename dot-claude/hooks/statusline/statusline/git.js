// Git segment — branch, dirty/staged, insertions/deletions, ahead/behind

const { execFileSync } = require('child_process');
const { osc8Link } = require('./shared');

function git(cwd, ...args) {
  try {
    return execFileSync('git', args, {
      encoding: 'utf8', cwd, timeout: 2000,
      stdio: ['pipe', 'pipe', 'pipe']
    }).trim();
  } catch { return null; }
}

function gitSegment(cwd) {
  if (!cwd) return '';
  const root = git(cwd, 'rev-parse', '--show-toplevel');
  if (!root) return '';

  const branch = git(cwd, 'rev-parse', '--abbrev-ref', 'HEAD');
  if (!branch) return '';

  let indicators = '';

  // Dirty (unstaged changes or untracked files)
  if (git(cwd, 'diff', '--quiet') === null) {
    indicators += ' *';
  } else {
    const untracked = git(cwd, 'ls-files', '--others', '--exclude-standard');
    if (untracked) indicators += ' *';
  }

  // Staged changes
  if (git(cwd, 'diff', '--cached', '--quiet') === null) {
    indicators += ' \u2713';
  }

  // Insertions/deletions
  const numstat = git(cwd, 'diff', '--numstat');
  if (numstat) {
    let ins = 0, del = 0;
    for (const line of numstat.split('\n')) {
      const fields = line.split(/\s+/);
      if (fields.length >= 2 && fields[0] !== '-') {
        ins += parseInt(fields[0]) || 0;
        del += parseInt(fields[1]) || 0;
      }
    }
    if (ins > 0) indicators += ` \x1b[32m+${ins}\x1b[33m`;
    if (del > 0) indicators += ` \x1b[31m-${del}\x1b[33m`;
  }

  // Ahead/behind remote
  const ab = git(cwd, 'rev-list', '--left-right', '--count', 'HEAD...@{upstream}');
  if (ab) {
    const [ahead, behind] = ab.split(/\s+/).map(Number);
    if (ahead > 0) indicators += ` \x1b[36m\u2191${ahead}\x1b[33m`;
    if (behind > 0) indicators += ` \x1b[36m\u2193${behind}\x1b[33m`;
  }

  // Branch text with optional GitHub link
  let branchText = `${branch}${indicators}`;
  const remote = git(cwd, 'remote', 'get-url', 'origin');
  if (remote) {
    let repoURL = '';
    if (remote.startsWith('git@github.com:')) {
      repoURL = 'https://github.com/' + remote.slice(15).replace(/\.git$/, '');
    } else if (remote.startsWith('https://github.com/')) {
      repoURL = remote.replace(/\.git$/, '');
    }
    if (repoURL) branchText = osc8Link(repoURL, branchText);
  }

  return `\x1b[33m\ue725 ${branchText}\x1b[0m`;
}

module.exports = { git, gitSegment };
