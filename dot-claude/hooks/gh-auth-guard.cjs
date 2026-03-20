#!/usr/bin/env node
const { execSync } = require("child_process");

function run(cmd) {
  try {
    return execSync(cmd, { encoding: "utf8", stdio: ["pipe", "pipe", "pipe"] }).trim();
  } catch {
    return null;
  }
}

function checkGhAuth() {
  const gitUser = run("git config user.name");
  if (!gitUser) return;

  const ghStatus = run("gh auth status 2>&1");
  if (!ghStatus) return;

  // Parse active account from gh auth status
  const lines = ghStatus.split("\n");
  let currentAccount = null;
  let activeAccount = null;
  for (const line of lines) {
    const match = line.match(/Logged in to github\.com account (\S+)/);
    if (match) currentAccount = match[1];
    if (line.includes("Active account: true") && currentAccount) {
      activeAccount = currentAccount;
    }
  }

  if (!activeAccount) return;
  if (gitUser.toLowerCase() === activeAccount.toLowerCase()) return;

  // Auto-switch
  const result = run(`gh auth switch --user ${gitUser} 2>&1`);
  if (result !== null) {
    console.log(`gh auth: switched "${activeAccount}" → "${gitUser}" (matches git config)`);
  } else {
    console.log(
      `⚠ gh account "${activeAccount}" does not match git user "${gitUser}".\n` +
      `Auto-switch failed. Run manually: gh auth switch --user ${gitUser}`
    );
  }
}

// --- main ---
let input = "";
process.stdin.setEncoding("utf8");
process.stdin.on("data", (chunk) => { input += chunk; });
process.stdin.on("end", () => {
  let data;
  try { data = JSON.parse(input); } catch { process.exit(0); }

  if (data.tool_name !== "Bash") process.exit(0);

  const command = data.tool_input?.command || "";

  const needsAuth =
    (/\bgh\s/.test(command) && !/\bgh\s+auth\s/.test(command)) ||
    /\bgit\s+(push|pull|fetch|clone)\b/.test(command);

  if (needsAuth) {
    checkGhAuth();
  }

  process.exit(0);
});
