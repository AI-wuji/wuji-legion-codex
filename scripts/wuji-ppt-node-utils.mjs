#!/usr/bin/env node

import path from "node:path";
import { createRequire } from "node:module";
import { pathToFileURL } from "node:url";

let artifactToolUtilsPromise;

export async function loadArtifactToolUtils() {
  if (!artifactToolUtilsPromise) {
    const skillDir = process.env.WUJI_PRESENTATIONS_SKILL_DIR;
    if (!skillDir) {
      throw new Error("Missing WUJI_PRESENTATIONS_SKILL_DIR. Invoke this script via a wuji PPT PowerShell wrapper.");
    }
    const modulePath = path.join(skillDir, "scripts", "artifact_tool_utils.mjs");
    artifactToolUtilsPromise = import(pathToFileURL(modulePath).href).catch((error) => {
      throw new Error(`Failed to load artifact_tool_utils from ${modulePath}: ${error.message}`);
    });
  }
  return artifactToolUtilsPromise;
}

export function loadNodePackage(packageName) {
  const require = createRequire(import.meta.url);
  const nodePath = process.env.NODE_PATH || "";
  const roots = nodePath
    .split(path.delimiter)
    .map((value) => value.trim())
    .filter(Boolean);

  const candidates = [packageName];
  for (const root of roots) {
    candidates.push(path.join(root, packageName));
    candidates.push(path.join(root, ".pnpm", "node_modules", packageName));
  }

  const errors = [];
  for (const candidate of [...new Set(candidates)]) {
    try {
      const mod = require(candidate);
      return mod.default || mod;
    } catch (error) {
      errors.push(`${candidate}: ${error.message}`);
    }
  }

  throw new Error(`Could not load ${packageName}.\n${errors.join("\n")}`);
}

export function loadJSZip() {
  return loadNodePackage("jszip");
}

export function relativeFromWorkspace(workspaceDir, filePath) {
  return path.relative(workspaceDir, filePath).split(path.sep).join("/");
}
