#!/usr/bin/env node
// Replace any same-named Forgejo release asset, then upload the final binary.

import { readFileSync } from "node:fs";

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
  // 1. Validate the workflow arguments.
  // The workflow passes the release ID file, built package path, and the name
  // the package should have on the Forgejo release page.
  if (process.argv.length !== 5) {
    throw new Error("Usage: upload-release-asset.mjs ID_FILE ASSET_PATH ASSET_NAME");
  }
  const [, , idFile, assetPath, assetName] = process.argv;

  // 2. Read Forgejo CI environment and build the release-assets API URL.
  const { FORGEJO_TOKEN, FORGEJO_SERVER_URL, FORGEJO_REPOSITORY } = process.env;
  const [owner, repo] = FORGEJO_REPOSITORY.split("/");

  // The previous workflow step wrote this ID after creating or finding the
  // release. This step only needs that ID to attach the package file.
  const releaseId = readFileSync(idFile, "utf8").trim();
  const apiBase = `${FORGEJO_SERVER_URL.replace(/\/$/, "")}/api/v1`;
  const assetsUrl = `${apiBase}/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/releases/${encodeURIComponent(releaseId)}/assets`;

  // 3. List existing release assets.
  // List the current attachments so a rerun can replace the old package
  // rather than creating another asset with the same name.
  const assetsResponse = await forgejoRequest("GET", assetsUrl, FORGEJO_TOKEN, {
    action: "Failed to list release assets",
    ok: [200],
  });
  const assets = await assetsResponse.json();
  const existingAsset = assets.find((asset) => asset.name === assetName);

  // 4. Delete the same-named asset if one already exists.
  if (existingAsset) {
    // This is deliberately the simple path: delete the old file first, then
    // upload the replacement. A failed upload can leave the release empty.
    await forgejoRequest("DELETE", `${assetsUrl}/${encodeURIComponent(existingAsset.id)}`, FORGEJO_TOKEN, {
      action: "Failed to delete existing asset",
      ok: [204],
    });
  }

  // 5. Upload the package under its final release filename.
  // Send the built .ndp file as the release attachment under its final name.
  await forgejoRequest("POST", `${assetsUrl}?name=${encodeURIComponent(assetName)}`, FORGEJO_TOKEN, {
    action: "Failed to upload release asset",
    ok: [201],
    headers: { "Content-Type": "application/octet-stream" },
    body: readFileSync(assetPath),
  });
}

main().catch((error) => {
  // Keep CI logs useful without printing Forgejo responses or token data.
  console.error(error.message);
  process.exit(1);
});
