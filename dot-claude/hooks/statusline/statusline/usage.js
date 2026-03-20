// Usage segment — Anthropic API (session, weekly, extra credits)

const fs = require('fs');
const os = require('os');
const path = require('path');
const { execFileSync } = require('child_process');
const { RST, SEP, osc8Link, thresholdColor, formatResetCompact } = require('./shared');

const CACHE_FILE   = path.join(os.tmpdir(), 'claude-usage-cache.json');
const LOCK_FILE    = path.join(os.tmpdir(), 'claude-usage-lock.json');
const CACHE_TTL    = 60;  // seconds
const MIN_INTERVAL = 30;  // minimum seconds between API calls

function loadUsageCache() {
  let cache = {};
  let needFetch = true;

  try {
    cache = JSON.parse(fs.readFileSync(CACHE_FILE, 'utf8'));
    if (Date.now() / 1000 - (cache.cached_at || 0) < CACHE_TTL) needFetch = false;
  } catch { /* no cache yet */ }

  if (needFetch) {
    try {
      const lock = JSON.parse(fs.readFileSync(LOCK_FILE, 'utf8'));
      const now = Date.now() / 1000;
      if ((lock.backoff_until || 0) > now) needFetch = false;
      else if (now - (lock.last_fetch_at || 0) < MIN_INTERVAL) needFetch = false;
    } catch { /* no lock yet */ }
  }

  return { cache, needFetch };
}

function fetchUsage() {
  const credsFile = path.join(os.homedir(), '.claude', '.credentials.json');
  let token;
  try {
    const creds = JSON.parse(fs.readFileSync(credsFile, 'utf8'));
    token = creds?.claudeAiOauth?.accessToken;
  } catch { return null; }
  if (!token) return null;

  try {
    const raw = execFileSync('curl', [
      '-s', '-w', '\n%{http_code}',
      '-H', `Authorization: Bearer ${token}`,
      '-H', 'Content-Type: application/json',
      '-H', 'anthropic-beta: oauth-2025-04-20',
      'https://api.anthropic.com/api/oauth/usage'
    ], { encoding: 'utf8', timeout: 5000, stdio: ['pipe', 'pipe', 'pipe'] });

    const lines = raw.trimEnd().split('\n');
    const code = parseInt(lines.pop());
    if (code === 429) return { __rate_limited: true };
    if (code !== 200) return null;
    return JSON.parse(lines.join('\n'));
  } catch { return null; }
}

function saveUsage() {
  const lock = { last_fetch_at: Math.floor(Date.now() / 1000) };
  try { fs.writeFileSync(LOCK_FILE, JSON.stringify(lock)); } catch { /* best-effort */ }

  const usage = fetchUsage();
  if (!usage) return;

  if (usage.__rate_limited) {
    lock.backoff_until = Math.floor(Date.now() / 1000) + 300;
    try { fs.writeFileSync(LOCK_FILE, JSON.stringify(lock)); } catch { /* best-effort */ }
    return;
  }

  try {
    fs.writeFileSync(CACHE_FILE, JSON.stringify({
      five_hour:   usage.five_hour   || {},
      seven_day:   usage.seven_day   || {},
      extra_usage: usage.extra_usage || {},
      cached_at:   Math.floor(Date.now() / 1000)
    }));
  } catch { /* best-effort */ }
}

function usageSegment(cache) {
  const parts = [];
  const usageURL = 'https://claude.ai/settings/usage';

  if (cache.five_hour?.utilization) {
    const util = Math.round(cache.five_hour.utilization);
    const color = thresholdColor(util);
    const reset = formatResetCompact(cache.five_hour.resets_at);
    parts.push(color + osc8Link(usageURL, `Session: ${util}%${reset}`) + RST);
  }

  if (cache.seven_day?.utilization) {
    const util = Math.round(cache.seven_day.utilization);
    const color = thresholdColor(util);
    const reset = formatResetCompact(cache.seven_day.resets_at);
    parts.push(color + osc8Link(usageURL, `Weekly: ${util}%${reset}`) + RST);
  }

  if (cache.extra_usage?.is_enabled) {
    const used  = cache.extra_usage.used_credits  || 0;
    const limit = cache.extra_usage.monthly_limit || 0;
    const util  = cache.extra_usage.utilization   || 0;
    const color = thresholdColor(util);
    parts.push(color + osc8Link(usageURL, `Extra: $${used.toFixed(2)}/$${limit.toFixed(2)}`) + RST);
  }

  return parts.join(SEP);
}

module.exports = { loadUsageCache, usageSegment, saveUsage };
