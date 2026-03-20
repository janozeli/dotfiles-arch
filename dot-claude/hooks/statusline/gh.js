// GitHub account segment — shows active gh auth account

const { execSync } = require('child_process');
const { RST } = require('./shared');

function ghSegment() {
  try {
    const out = execSync('gh auth status 2>&1', { encoding: 'utf8', timeout: 3000 });
    let current = null;
    for (const line of out.split('\n')) {
      const match = line.match(/Logged in to github\.com account (\S+)/);
      if (match) current = match[1];
      if (line.includes('Active account: true') && current) {
        return `\x1b[38;2;110;84;148m\uf09b ${current}${RST}`;
      }
    }
  } catch {}
  return '';
}

module.exports = { ghSegment };
