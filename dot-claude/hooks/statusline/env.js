// Environment segment — project root + directory with editor deep links

const os = require('os');
const path = require('path');
const { RST, SEP, osc8Link, editorFileURL } = require('./shared');

function envSegment(cwd, root) {
  if (!cwd) return '';
  const parts = [];

  // Project root (only shown if cwd is inside a subdirectory)
  if (root && root !== cwd) {
    const name = path.basename(root);
    parts.push('\x1b[38;2;97;175;239m\uf07c ' + osc8Link(editorFileURL(root), name) + RST);
  }

  // Directory with ~ shortening and editor deep link
  const home = os.homedir();
  let short = cwd;
  if (home && cwd.startsWith(home)) {
    short = '~' + (cwd.slice(home.length) || '/');
  }
  const p = short.split('/');
  if (p.length > 3) {
    short = p[0] + '/' + p[1] + '/.../' + p[p.length - 1];
  }
  parts.push('\x1b[38;2;97;175;239m\ue5ff ' + osc8Link(editorFileURL(cwd), short) + RST);

  return parts.join(SEP);
}

module.exports = { envSegment };
