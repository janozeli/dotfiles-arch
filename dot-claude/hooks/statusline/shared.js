// Shared ANSI helpers and utilities for statusline extensions

const path = require('path');

const DIM = '\x1b[2m';
const RST = '\x1b[0m';
const SEP = ` ${DIM}\u2502${RST} `;

function osc8Link(url, text) {
  return `\x1b]8;;${url}\x1b\\${text}\x1b]8;;\x1b\\`;
}

function editorFileURL(filePath) {
  const editor = path.basename(process.env.EDITOR || '');
  const schemes = { zed: 'zed://file', code: 'vscode://file', cursor: 'cursor://file' };
  return (schemes[editor] || 'file://') + filePath;
}

function thresholdColor(pct) {
  if (pct >= 85) return '\x1b[31m';
  if (pct >= 60) return '\x1b[33m';
  return '\x1b[32m';
}

function formatResetCompact(resetsAt) {
  if (!resetsAt) return '';
  const diff = new Date(resetsAt) - Date.now();
  if (diff <= 0) return '';
  const h = Math.floor(diff / 3600000);
  const m = Math.floor((diff % 3600000) / 60000);
  const d = Math.floor(h / 24);
  if (d > 0) return ` \uf1da ${d}d${String(h % 24).padStart(2, '0')}h`;
  return ` \uf1da ${h}h${String(m).padStart(2, '0')}`;
}

module.exports = { DIM, RST, SEP, osc8Link, editorFileURL, thresholdColor, formatResetCompact };
