#!/usr/bin/env node
// Create or reuse the Forgejo release for the current tag and write its ID.

import { execFileSync } from "node:child_process";
import { writeFileSync } from "node:fs";

function runGit(args) {
  return execFileSync("git", args, { encoding: "utf8", stdio: ["ignore", "pipe", "ignore"] }).trim();
}

function buildReleaseNotes(tagName) {
  const currentCommit = runGit(["rev-list", "-n", "1", tagName]);
  let previousTag = "";

  try {
    // Start at the parent of the tagged commit so `git describe` finds the
    // previous reachable v* tag, not the tag that triggered this workflow.
    previousTag = runGit(["describe", "--tags", "--match", "v*", "--abbrev=0", `${currentCommit}^`]);
  } catch {
    // No previous reachable v* tag.
  }

  const commitRange = previousTag ? `${previousTag}..${currentCommit}` : currentCommit;
  const heading = previousTag ? `## Changes since ${previousTag}` : "## Changes";
  const changelog = runGit(["log", "--no-merges", "--format=- %s (%h)", commitRange]) || "- No commits found.";

  return `${heading}\n\n${changelog}`;
}

async function forgejoRequest(method, url, token, options = {}) {
  const response = await fetch(url, {
    method,
    headers: {
      Authorization: `token ${token}`,
      Accept: "application/json",
      ...options.headers,
    },
    body: options.body,
  });

  if (!options.ok.includes(response.status)) {
    throw new Error(`${options.action}, HTTP status ${response.status}`);
  }

  return response;
}

async function main() {
  // The workflow passes the output path as the only argument and provides the
  // normal Forgejo CI environment variables used below.
  if (process.argv.length !== 3) {
    throw new Error("Usage: generate-release.mjs ID_FILE");
  }
  const idFile = process.argv[2];

  const { FORGEJO_TOKEN, FORGEJO_SERVER_URL, FORGEJO_REPOSITORY, FORGEJO_REF_NAME } = process.env;
  const [owner, repo] = FORGEJO_REPOSITORY.split("/");
  const apiBase = `${FORGEJO_SERVER_URL.replace(/\/$/, "")}/api/v1`;
  const releasesUrl = `${apiBase}/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/releases`;
  const releaseByTagUrl = `${releasesUrl}/tags/${encodeURIComponent(FORGEJO_REF_NAME)}`;

  // A 404 is an expected result here: it means this tag does not have a
  // release yet, so the script should create one instead of failing.
  let releaseResponse = await forgejoRequest("GET", releaseByTagUrl, FORGEJO_TOKEN, {
    action: "Failed to get release",
    ok: [200, 404],
  });

  if (releaseResponse.status === 404) {
    const releaseNotes = buildReleaseNotes(FORGEJO_REF_NAME);

    releaseResponse = await forgejoRequest("POST", releasesUrl, FORGEJO_TOKEN, {
      action: "Failed to create release",
      ok: [201],
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        tag_name: FORGEJO_REF_NAME,
        name: FORGEJO_REF_NAME,
        body: releaseNotes,
        draft: false,
        prerelease: false,
      }),
    });
  }

  const release = await releaseResponse.json();

  // The upload step reads this file so it knows which release to attach the
  // package asset to.
  writeFileSync(idFile, `${release.id}\n`);
}

main().catch((error) => {
  // Print only concise error text. Do not print API responses or token data.
  console.error(error.message);
  process.exit(1);
});
