#!/usr/bin/env bun
// @bun

// src/config.ts
import { readFile } from "fs/promises";
import { join, dirname } from "path";
import { fileURLToPath } from "url";
function parseYAML(content) {
  const lines = content.split(`
`);
  const result = {};
  const stack = [{ obj: result, indent: -2 }];
  for (let line of lines) {
    if (line.trim().startsWith("#") || line.trim() === "")
      continue;
    const indent = line.search(/\S/);
    if (indent === -1)
      continue;
    const content2 = line.trim();
    if (content2.includes(":")) {
      const [key, ...valueParts] = content2.split(":");
      const value = valueParts.join(":").trim();
      while (stack.length > 1 && indent <= stack[stack.length - 1].indent) {
        stack.pop();
      }
      const parent = stack[stack.length - 1].obj;
      if (value === "") {
        const newObj = {};
        parent[key.trim()] = newObj;
        stack.push({ obj: newObj, indent });
      } else if (value.startsWith('"') || value.startsWith("'")) {
        parent[key.trim()] = value.slice(1, -1);
      } else {
        parent[key.trim()] = value;
      }
    } else if (content2.startsWith("- ")) {
      const parent = stack[stack.length - 1].obj;
      const lastKey = Object.keys(parent).pop();
      if (lastKey && !Array.isArray(parent[lastKey])) {
        parent[lastKey] = [];
      }
      if (lastKey) {
        const value = content2.slice(2).trim();
        parent[lastKey].push(value.startsWith('"') || value.startsWith("'") ? value.slice(1, -1) : value);
      }
    }
  }
  return result;
}
var __filename2 = fileURLToPath(import.meta.url);
var __dirname2 = dirname(__filename2);
var loadedConfig = null;
async function loadConfig() {
  if (loadedConfig) {
    return loadedConfig;
  }
  const defaultConfigPath = join(__dirname2, "..", "providers.yaml");
  const defaultData = await readFile(defaultConfigPath, "utf-8");
  const config = parseYAML(defaultData);
  const userConfig = await loadUserConfig();
  if (userConfig) {
    mergeConfigs(config, userConfig);
  }
  loadedConfig = config;
  return config;
}
async function loadUserConfig() {
  const configPath = getUserConfigPath();
  if (!configPath) {
    return null;
  }
  try {
    const data = await readFile(configPath, "utf-8");
    const config = parseYAML(data);
    return config;
  } catch (err) {
    return null;
  }
}
function getUserConfigPath() {
  const xdgConfigHome = process.env.XDG_CONFIG_HOME;
  if (xdgConfigHome) {
    return join(xdgConfigHome, "agents", "providers.yaml");
  }
  const homeDir = process.env.HOME;
  if (!homeDir) {
    return null;
  }
  return join(homeDir, ".config", "agents", "providers.yaml");
}
function mergeConfigs(base, override) {
  for (const [name, userProvider] of Object.entries(override.providers)) {
    const baseProvider = base.providers[name];
    if (baseProvider) {
      base.providers[name] = mergeProvider(baseProvider, userProvider);
    } else {
      base.providers[name] = userProvider;
    }
  }
}
function mergeProvider(base, override) {
  const result = { ...base };
  if (override.display_name) {
    result.display_name = override.display_name;
  }
  if (override.guideline.file) {
    result.guideline.file = override.guideline.file;
  }
  if (override.guideline.dir) {
    result.guideline.dir = override.guideline.dir;
  }
  if (override.guideline.source) {
    result.guideline.source = override.guideline.source;
  }
  if (override.guideline.global_paths.length > 0) {
    result.guideline.global_paths = override.guideline.global_paths;
  }
  if (override.commands.dir) {
    result.commands.dir = override.commands.dir;
  }
  if (override.commands.extension) {
    result.commands.extension = override.commands.extension;
  }
  if (override.commands.source_dir) {
    result.commands.source_dir = override.commands.source_dir;
  }
  if (override.commands.global_paths.length > 0) {
    result.commands.global_paths = override.commands.global_paths;
  }
  return result;
}
function expandPath(path) {
  if (!path.startsWith("~")) {
    return path;
  }
  const homeDir = process.env.HOME;
  if (!homeDir) {
    return path;
  }
  if (path === "~") {
    return homeDir;
  }
  return join(homeDir, path.slice(2));
}
function expandProviderPaths(provider) {
  const result = { ...provider };
  result.guideline.global_paths = provider.guideline.global_paths.map(expandPath);
  result.commands.global_paths = provider.commands.global_paths.map(expandPath);
  return result;
}

// src/agents.ts
var supportedAgents = null;
async function initSupportedAgents() {
  if (supportedAgents) {
    return supportedAgents;
  }
  const config = await loadConfig();
  supportedAgents = {};
  for (const [name, provider] of Object.entries(config.providers)) {
    supportedAgents[name] = {
      name: provider.name,
      file: provider.guideline.file,
      dir: provider.guideline.dir
    };
  }
  return supportedAgents;
}
async function getSupportedAgents() {
  return await initSupportedAgents();
}
async function getAgentNames() {
  const agents = await initSupportedAgents();
  return Object.keys(agents);
}

// src/discovery.ts
import { readdir, stat, lstat } from "fs/promises";
import { join as join2, dirname as dirname2, basename } from "path";
var ignoreDir = new Set(["node_modules", ".git", "dist", "build", ".cursor"]);
async function discoverAgents() {
  const agents = [];
  const cwd = process.cwd();
  await walkDirectory(cwd, async (path, stats) => {
    if (stats.isFile() && basename(path) === "AGENTS.md") {
      agents.push(path);
    }
  });
  return agents;
}
async function discoverAll() {
  const files = [];
  const cwd = process.cwd();
  const supportedAgents2 = await getSupportedAgents();
  await walkDirectory(cwd, async (path, stats) => {
    if (stats.isFile()) {
      const dir = dirname2(path);
      const filename = basename(path);
      let agent = "";
      if (filename === "AGENTS.md") {
        agent = "AGENTS";
      } else {
        for (const [agentName, cfg] of Object.entries(supportedAgents2)) {
          if (filename === cfg.file) {
            if (cfg.dir === "" || path.includes(cfg.dir)) {
              agent = agentName.toUpperCase();
              break;
            }
          }
        }
      }
      if (agent) {
        const isSymlink = await isSymlinkFile(path);
        files.push({
          path,
          dir,
          agent,
          file: filename,
          isSymlink,
          size: stats.size
        });
      }
    }
  }, true);
  return files;
}
async function discoverGlobalOnly() {
  const files = [];
  const globalLocations = globalGuidelinePaths();
  for (const location of globalLocations) {
    if (await fileExists(location)) {
      try {
        const stats = await stat(location);
        const filename = basename(location);
        const dir = dirname2(location);
        const agent = inferAgentFromFilename(filename);
        if (agent) {
          const isSymlink = await isSymlinkFile(location);
          files.push({
            path: location,
            dir,
            agent,
            file: filename,
            isSymlink,
            size: stats.size
          });
        }
      } catch {}
    }
  }
  return files;
}
function globalGuidelinePaths() {
  const homeDir = process.env.HOME;
  if (!homeDir) {
    return [];
  }
  return [
    join2(homeDir, ".claude", "CLAUDE.md"),
    join2(homeDir, ".codex", "AGENTS.md"),
    join2(homeDir, ".gemini", "GEMINI.md"),
    join2(homeDir, ".config", "opencode", "AGENTS.md"),
    join2(homeDir, ".config", "amp", "AGENTS.md"),
    join2(homeDir, ".config", "AGENTS.md"),
    join2(homeDir, "AGENTS.md")
  ];
}
async function inferAgentFromFilename(filename) {
  if (filename === "AGENTS.md") {
    return "AGENTS";
  }
  const supportedAgents2 = await getSupportedAgents();
  for (const [agentName, cfg] of Object.entries(supportedAgents2)) {
    if (filename === cfg.file) {
      return agentName.toUpperCase();
    }
  }
  return "";
}
async function fileExists(path) {
  try {
    const stats = await stat(path);
    return stats.isFile();
  } catch {
    return false;
  }
}
async function dirExists(path) {
  try {
    const stats = await stat(path);
    return stats.isDirectory();
  } catch {
    return false;
  }
}
async function isSymlinkFile(path) {
  try {
    const stats = await lstat(path);
    return stats.isSymbolicLink();
  } catch {
    return false;
  }
}
async function walkDirectory(dir, callback, allowCursor = false) {
  try {
    const entries = await readdir(dir, { withFileTypes: true });
    for (const entry of entries) {
      const fullPath = join2(dir, entry.name);
      if (entry.isDirectory()) {
        if (ignoreDir.has(entry.name) && !(allowCursor && entry.name === ".cursor")) {
          continue;
        }
        await walkDirectory(fullPath, callback, allowCursor);
      } else {
        const stats = await stat(fullPath);
        await callback(fullPath, stats);
      }
    }
  } catch {}
}

// src/commands.ts
import { readdir as readdir2, stat as stat2, lstat as lstat2, symlink, unlink, readlink, mkdir } from "fs/promises";
import { join as join3, dirname as dirname3, basename as basename2, relative, extname } from "path";
var ignoreDir2 = new Set(["node_modules", ".git", "dist", "build", ".cursor"]);
async function discoverCommands() {
  const commandDirs = [];
  const cwd = process.cwd();
  await walkDirectory2(cwd, async (path, stats) => {
    if (stats.isDirectory() && basename2(path) === "COMMANDS") {
      commandDirs.push(path);
    }
  });
  return commandDirs;
}
async function discoverAllCommands() {
  const files = [];
  const cwd = process.cwd();
  const config = await loadConfig();
  await walkDirectory2(cwd, async (path, stats) => {
    if (stats.isFile()) {
      const dir = dirname3(path);
      const filename = basename2(path);
      for (const provider of Object.values(config.providers)) {
        if (!path.includes(provider.commands.dir)) {
          continue;
        }
        if (provider.commands.extension && !filename.endsWith(provider.commands.extension)) {
          continue;
        }
        let cmdName = filename;
        if (provider.commands.extension) {
          cmdName = filename.slice(0, -provider.commands.extension.length);
        }
        const isSymlink = await isSymlinkFile(path);
        files.push({
          path,
          dir,
          provider: provider.name.toUpperCase(),
          name: cmdName,
          file: filename,
          isSymlink,
          size: stats.size
        });
        break;
      }
    }
  });
  return files;
}
async function discoverGlobalCommands() {
  const files = [];
  const config = await loadConfig();
  for (const provider of Object.values(config.providers)) {
    const expandedProvider = expandProviderPaths(provider);
    for (const globalPath of expandedProvider.commands.global_paths) {
      if (!await dirExists(globalPath)) {
        continue;
      }
      await walkDirectory2(globalPath, async (path, stats) => {
        if (stats.isFile()) {
          const filename = basename2(path);
          if (provider.commands.extension && !filename.endsWith(provider.commands.extension)) {
            return;
          }
          let cmdName = filename;
          if (provider.commands.extension) {
            cmdName = filename.slice(0, -provider.commands.extension.length);
          }
          const isSymlink = await isSymlinkFile(path);
          files.push({
            path,
            dir: dirname3(path),
            provider: provider.name.toUpperCase(),
            name: cmdName,
            file: filename,
            isSymlink,
            size: stats.size
          });
        }
      });
    }
  }
  return files;
}
async function syncCommandFiles(commandDirs, selectedProviders, dryRun, verbose) {
  let created = 0;
  let skipped = 0;
  const operations = [];
  const config = await loadConfig();
  for (const commandDir of commandDirs) {
    const parentDir = dirname3(commandDir);
    let entries;
    try {
      entries = await readdir2(commandDir, { withFileTypes: true });
    } catch (err) {
      if (verbose) {
        operations.push(`error reading ${commandDir}: ${err}`);
      }
      continue;
    }
    for (const providerName of selectedProviders) {
      const provider = config.providers[providerName];
      if (!provider) {
        continue;
      }
      const targetDir = join3(parentDir, provider.commands.dir);
      if (!dryRun) {
        await mkdir(targetDir, { recursive: true });
      }
      for (const entry of entries) {
        if (entry.isDirectory()) {
          continue;
        }
        const sourceFile = entry.name;
        if (!sourceFile.endsWith(".md")) {
          continue;
        }
        const sourcePath = join3(commandDir, sourceFile);
        let targetFile = sourceFile;
        if (provider.commands.extension && provider.commands.extension !== extname(sourceFile)) {
          const baseName = sourceFile.slice(0, -extname(sourceFile).length);
          targetFile = baseName + provider.commands.extension;
        }
        const targetPath = join3(targetDir, targetFile);
        const relPath = relative(targetDir, sourcePath);
        const result = await createCommandSymlink(targetPath, relPath, dryRun, verbose);
        created += result.created;
        skipped += result.skipped;
        if (result.operation) {
          operations.push(result.operation);
        }
      }
    }
  }
  console.log(`COMMANDS directories found: ${commandDirs.length}`);
  console.log(`Command symlinks created: ${created}`);
  console.log(`Command symlinks skipped: ${skipped}`);
  if (verbose && operations.length > 0) {
    console.log(`
Operations:`);
    for (const op of operations) {
      console.log(op);
    }
  }
}
async function createCommandSymlink(targetPath, sourcePath, dryRun, verbose) {
  try {
    const stats = await lstat2(targetPath);
    if (stats.isSymbolicLink()) {
      const link = await readlink(targetPath);
      if (link === sourcePath) {
        return {
          created: 0,
          skipped: 1,
          operation: verbose ? `skipped: ${targetPath} (already correct)` : ""
        };
      }
    }
    if (!dryRun) {
      await unlink(targetPath);
    }
  } catch {}
  if (dryRun) {
    return {
      created: 1,
      skipped: 0,
      operation: verbose ? `would create: ${targetPath} -> ${sourcePath}` : ""
    };
  }
  try {
    await symlink(sourcePath, targetPath);
    return { created: 1, skipped: 0, operation: verbose ? `created: ${targetPath} -> ${sourcePath}` : "" };
  } catch (err) {
    return { created: 0, skipped: 1, operation: verbose ? `error: ${targetPath} (${err})` : "" };
  }
}
async function deleteCommandFiles(selectedProviders, dryRun, verbose) {
  let deleted = 0;
  let notFound = 0;
  const operations = [];
  const allFiles = await discoverAllCommands();
  for (const file of allFiles) {
    const shouldDelete = selectedProviders.some((providerName) => file.provider.toLowerCase() === providerName.toLowerCase());
    if (!shouldDelete) {
      continue;
    }
    if (dryRun) {
      deleted++;
      if (verbose) {
        operations.push(`would delete: ${file.path}`);
      }
    } else {
      try {
        await unlink(file.path);
        deleted++;
        if (verbose) {
          operations.push(`deleted: ${file.path}`);
        }
      } catch (err) {
        notFound++;
        if (verbose) {
          operations.push(`error: ${file.path} (${err})`);
        }
      }
    }
  }
  console.log(`Files deleted: ${deleted}`);
  if (notFound > 0) {
    console.log(`Errors: ${notFound}`);
  }
  if (verbose && operations.length > 0) {
    console.log("Operations:");
    for (const op of operations) {
      console.log(op);
    }
  }
}
async function walkDirectory2(dir, callback) {
  try {
    const entries = await readdir2(dir, { withFileTypes: true });
    for (const entry of entries) {
      const fullPath = join3(dir, entry.name);
      if (entry.isDirectory()) {
        if (ignoreDir2.has(entry.name)) {
          continue;
        }
        await walkDirectory2(fullPath, callback);
      } else {
        const stats = await stat2(fullPath);
        await callback(fullPath, stats);
      }
    }
  } catch {}
}

// src/symlink.ts
import { readFile as readFile2, unlink as unlink2, symlink as symlink2, mkdir as mkdir2, readlink as readlink2, lstat as lstat3 } from "fs/promises";
import { join as join4, relative as relative2, dirname as dirname4 } from "path";
import { createInterface } from "readline";
async function syncSymlinks(agentsPaths, selectedAgents, dryRun, verbose) {
  let created = 0;
  let skipped = 0;
  const operations = [];
  const supportedAgents2 = await getSupportedAgents();
  for (const agentPath of agentsPaths) {
    const dir = dirname4(agentPath);
    const filename = agentPath.split("/").pop();
    for (const agentName of selectedAgents) {
      const cfg = supportedAgents2[agentName];
      let result;
      if (cfg.dir === "") {
        result = await createSymlink(dir, filename, cfg.file, dryRun, verbose);
      } else {
        result = await createSymlinkInDir(dir, filename, cfg.dir, cfg.file, dryRun, verbose);
      }
      created += result.created;
      skipped += result.skipped;
      if (result.operation) {
        operations.push(result.operation);
      }
    }
  }
  console.log(`AGENTS.md files found: ${agentsPaths.length}`);
  console.log(`Symlinks created: ${created}`);
  console.log(`Symlinks skipped: ${skipped}`);
  if (verbose && operations.length > 0) {
    console.log(`
Operations:`);
    for (const op of operations) {
      console.log(op);
    }
  }
}
async function createSymlink(dir, source, target, dryRun, verbose) {
  const targetPath = join4(dir, target);
  try {
    const stats = await lstat3(targetPath);
    if (await shouldSkipOrOverwrite(targetPath, source, stats, join4(dir, source), dryRun, verbose)) {
      return { created: 0, skipped: 1, operation: verbose ? `skipped: ${targetPath} (already correct)` : "" };
    }
    if (!dryRun) {
      await unlink2(targetPath);
    }
  } catch {}
  if (dryRun) {
    return {
      created: 1,
      skipped: 0,
      operation: verbose ? `would create: ${targetPath} -> ${source}` : ""
    };
  }
  try {
    await symlink2(source, targetPath);
    return { created: 1, skipped: 0, operation: verbose ? `created: ${targetPath} -> ${source}` : "" };
  } catch (err) {
    return { created: 0, skipped: 1, operation: verbose ? `error: ${targetPath} (${err})` : "" };
  }
}
async function createSymlinkInDir(dir, source, subdir, target, dryRun, verbose) {
  const subdirPath = join4(dir, subdir);
  if (!dryRun) {
    await mkdir2(subdirPath, { recursive: true });
  }
  const targetPath = join4(subdirPath, target);
  const symTarget = relative2(subdirPath, join4(dir, source));
  try {
    const stats = await lstat3(targetPath);
    if (await shouldSkipOrOverwrite(targetPath, symTarget, stats, join4(dir, source), dryRun, verbose)) {
      return { created: 0, skipped: 1, operation: verbose ? `skipped: ${targetPath} (already correct)` : "" };
    }
    if (!dryRun) {
      await unlink2(targetPath);
    }
  } catch {}
  if (dryRun) {
    return {
      created: 1,
      skipped: 0,
      operation: verbose ? `would create: ${targetPath} -> ${symTarget}` : ""
    };
  }
  try {
    await symlink2(symTarget, targetPath);
    return { created: 1, skipped: 0, operation: verbose ? `created: ${targetPath} -> ${symTarget}` : "" };
  } catch (err) {
    return { created: 0, skipped: 1, operation: verbose ? `error: ${targetPath} (${err})` : "" };
  }
}
async function shouldSkipOrOverwrite(targetPath, expectedTarget, stats, sourcePath, dryRun, verbose) {
  if (stats.isSymbolicLink()) {
    try {
      const link = await readlink2(targetPath);
      if (link === expectedTarget) {
        return true;
      }
    } catch {}
    if (!dryRun && !await askForConfirmation(targetPath, expectedTarget)) {
      return true;
    }
    return false;
  }
  try {
    const sourceContent = await readFile2(sourcePath, "utf-8");
    const targetContent = await readFile2(targetPath, "utf-8");
    if (sourceContent === targetContent) {
      return true;
    }
  } catch {}
  if (!dryRun && !await askForConfirmation(targetPath, "different version")) {
    return true;
  }
  return false;
}
async function askForConfirmation(targetPath, reason) {
  return new Promise((resolve) => {
    const rl = createInterface({
      input: process.stdin,
      output: process.stdout
    });
    console.log(`
File already exists: ${targetPath} (${reason})`);
    rl.question("Overwrite? (y/n): ", (answer) => {
      rl.close();
      const response = answer.trim().toLowerCase();
      resolve(response === "y" || response === "yes");
    });
  });
}
async function deleteGuidelineFiles(selectedAgents, dryRun, verbose) {
  let deleted = 0;
  let notFound = 0;
  const operations = [];
  const allFiles = await discoverAll();
  for (const file of allFiles) {
    const shouldDelete = selectedAgents.some((agentName) => file.agent.toLowerCase() === agentName.toLowerCase());
    if (!shouldDelete) {
      continue;
    }
    if (dryRun) {
      deleted++;
      if (verbose) {
        operations.push(`would delete: ${file.path}`);
      }
    } else {
      try {
        await unlink2(file.path);
        deleted++;
        if (verbose) {
          operations.push(`deleted: ${file.path}`);
        }
      } catch (err) {
        notFound++;
        if (verbose) {
          operations.push(`error: ${file.path} (${err})`);
        }
      }
    }
  }
  console.log(`Files deleted: ${deleted}`);
  if (notFound > 0) {
    console.log(`Errors: ${notFound}`);
  }
  if (verbose && operations.length > 0) {
    console.log("Operations:");
    for (const op of operations) {
      console.log(op);
    }
  }
}

// src/output.ts
import { relative as relative3 } from "path";
function estimateTokens(size) {
  return Math.floor(size / 4);
}
function formatList(files, verbose) {
  if (files.length === 0) {
    console.log("No guideline files found.");
    return;
  }
  console.log("Directory".padEnd(30) + "File".padEnd(15) + "Tokens");
  console.log("-".repeat(55));
  const cwd = process.cwd();
  const homeDir = process.env.HOME || "";
  let showRelativeToHome = false;
  if (files.length > 0 && homeDir) {
    const standardGlobalPatterns = [
      `${homeDir}/.claude`,
      `${homeDir}/.codex`,
      `${homeDir}/.gemini`,
      `${homeDir}/.config`
    ];
    const allFilesAreGlobal = files.every((f) => standardGlobalPatterns.some((pattern) => f.dir.startsWith(pattern)));
    showRelativeToHome = allFilesAreGlobal;
  }
  for (const f of files) {
    let displayDir;
    if (showRelativeToHome && homeDir && f.dir.startsWith(homeDir)) {
      const relDir = f.dir.slice(homeDir.length);
      displayDir = relDir === "" || relDir === "/" ? "~/" : `~${relDir}`;
    } else {
      const relDir = relative3(cwd, f.dir);
      if (relDir === "") {
        displayDir = "./";
      } else if (!relDir.startsWith(".")) {
        displayDir = "./" + relDir;
      } else {
        displayDir = relDir + "/";
      }
    }
    let filename = f.file;
    let tokensStr;
    if (f.isSymlink) {
      filename = "*" + filename;
      tokensStr = "-";
    } else {
      const tokens = estimateTokens(f.size);
      tokensStr = tokens.toString();
    }
    console.log(displayDir.padEnd(30) + filename.padEnd(15) + tokensStr);
  }
  if (verbose) {
    console.log(`
Total: ${files.length} files found`);
  }
}
function formatCommandList(files, verbose) {
  if (files.length === 0) {
    console.log("No command files found.");
    return;
  }
  console.log("Directory".padEnd(30) + "Command".padEnd(20) + "Provider".padEnd(15) + "Tokens");
  console.log("-".repeat(75));
  const cwd = process.cwd();
  const homeDir = process.env.HOME || "";
  let showRelativeToHome = false;
  if (files.length > 0 && homeDir) {
    const standardGlobalPatterns = [
      `${homeDir}/.claude`,
      `${homeDir}/.cursor`,
      `${homeDir}/.config`
    ];
    const allFilesAreGlobal = files.every((f) => standardGlobalPatterns.some((pattern) => f.dir.startsWith(pattern)));
    showRelativeToHome = allFilesAreGlobal;
  }
  for (const f of files) {
    let displayDir;
    if (showRelativeToHome && homeDir && f.dir.startsWith(homeDir)) {
      const relDir = f.dir.slice(homeDir.length);
      displayDir = relDir === "" || relDir === "/" ? "~/" : `~${relDir}`;
    } else {
      const relDir = relative3(cwd, f.dir);
      if (relDir === "") {
        displayDir = "./";
      } else if (!relDir.startsWith(".")) {
        displayDir = "./" + relDir;
      } else {
        displayDir = relDir + "/";
      }
    }
    let cmdName = f.name;
    let tokensStr;
    if (f.isSymlink) {
      cmdName = "*" + cmdName;
      tokensStr = "-";
    } else {
      const tokens = estimateTokens(f.size);
      tokensStr = tokens.toString();
    }
    console.log(displayDir.padEnd(30) + cmdName.padEnd(20) + f.provider.padEnd(15) + tokensStr);
  }
  if (verbose) {
    console.log(`
Total: ${files.length} command files found`);
  }
}

// src/main.ts
async function main() {
  const args = process.argv.slice(2);
  if (args.length === 0) {
    printHelp();
    process.exit(1);
  }
  const command = args[0];
  switch (command) {
    case "list":
      await cmdList(args.slice(1));
      break;
    case "list-commands":
    case "list-cmds":
      await cmdListCommands(args.slice(1));
      break;
    case "sync":
      await cmdSync(args.slice(1));
      break;
    case "sync-commands":
    case "sync-cmds":
      await cmdSyncCommands(args.slice(1));
      break;
    case "rm":
      await cmdRm(args.slice(1));
      break;
    case "rm-commands":
    case "rm-cmds":
      await cmdRmCommands(args.slice(1));
      break;
    case "help":
    case "-h":
    case "--help":
      printHelp();
      break;
    default:
      console.log(`Unknown command: ${command}`);
      process.exit(1);
  }
}
async function printHelp() {
  const agentNames = await getAgentNames();
  const agents = await getSupportedAgents();
  console.log(`Agent Guidelines Manager CLI

Usage: agents <command> [flags]

Commands:
  list                     Discover and display all guideline files with metadata
                           Flags:
                             --verbose    Show detailed output
                             -g           Show only user/system-wide agent guideline files
                             --global     Show only user/system-wide agent guideline files
                             --<agent>    Filter by specific agent files (e.g., --claude, --cursor)

  list-commands            Discover and display all slash command files
  list-cmds                Alias for list-commands
                           Flags:
                             --verbose    Show detailed output
                             -g           Show only user/system-wide command files
                             --global     Show only user/system-wide command files
                             --<agent>    Filter by specific agent (e.g., --claude, --cursor)

  sync                     Find all AGENTS.md files and create symlinks
                           Flags:`);
  for (const agent of agentNames) {
    const cfg = agents[agent];
    console.log(`                             --${cfg.name}       Create ${cfg.file} symlinks`);
  }
  console.log(`                             --dry-run    Show what would be created without making changes
                             --verbose    Show detailed output of all operations

  sync-commands            Find all COMMANDS directories and create command symlinks
  sync-cmds                Alias for sync-commands
                           Flags:`);
  for (const agent of agentNames) {
    const cfg = agents[agent];
    console.log(`                             --${cfg.name}       Create command symlinks for ${cfg.name}`);
  }
  console.log(`                             --dry-run    Show what would be created without making changes
                             --verbose    Show detailed output of all operations

  rm                       Delete guideline files for specified agents
                           Flags:`);
  for (const agent of agentNames) {
    const cfg = agents[agent];
    console.log(`                             --${cfg.name}       Delete ${cfg.file} files`);
  }
  console.log(`                             --dry-run    Show what would be deleted without making changes
                             --verbose    Show detailed output of all operations

  rm-commands              Delete command files for specified agents
  rm-cmds                  Alias for rm-commands
                           Flags:`);
  for (const agent of agentNames) {
    const cfg = agents[agent];
    console.log(`                             --${cfg.name}       Delete command files for ${cfg.name}`);
  }
  console.log(`                             --dry-run    Show what would be deleted without making changes
                             --verbose    Show detailed output of all operations

  help                     Show this help message

Examples:
  # Guideline files
  agents list
  agents list --verbose
  agents list --claude
  agents list --gemini --global
  agents sync --claude --cursor
  agents sync --claude --cursor --dry-run
  agents rm --claude
  agents rm --cursor --gemini --dry-run

  # Slash commands
  agents list-commands
  agents list-commands --verbose
  agents list-commands --claude --global
  agents sync-commands --claude --cursor
  agents sync-commands --claude --dry-run
  agents rm-commands --claude
`);
}
async function cmdList(args) {
  const flags = parseFlags(args, ["verbose", "g", "global"]);
  const verbose = flags.verbose;
  const global = flags.g || flags.global;
  const agentNames = await getAgentNames();
  const filterAgents = [];
  for (const agent of agentNames) {
    if (flags[agent]) {
      filterAgents.push(agent);
    }
  }
  let files = global ? await discoverGlobalOnly() : await discoverAll();
  if (filterAgents.length > 0) {
    files = files.filter((f) => filterAgents.some((agent) => agent.toUpperCase() === f.agent));
  }
  formatList(files, verbose);
}
async function cmdListCommands(args) {
  const flags = parseFlags(args, ["verbose", "g", "global"]);
  const verbose = flags.verbose;
  const global = flags.g || flags.global;
  const agentNames = await getAgentNames();
  const filterAgents = [];
  for (const agent of agentNames) {
    if (flags[agent]) {
      filterAgents.push(agent);
    }
  }
  let files = global ? await discoverGlobalCommands() : await discoverAllCommands();
  if (filterAgents.length > 0) {
    files = files.filter((f) => filterAgents.some((agent) => agent.toUpperCase() === f.provider));
  }
  formatCommandList(files, verbose);
}
async function cmdSync(args) {
  const flags = parseFlags(args, ["dry-run", "verbose"]);
  const dryRun = flags["dry-run"];
  const verbose = flags.verbose;
  const agentNames = await getAgentNames();
  const selectedAgents = [];
  for (const agent of agentNames) {
    if (flags[agent]) {
      selectedAgents.push(agent);
    }
  }
  if (selectedAgents.length === 0) {
    console.log(`Please specify at least one agent flag: --${agentNames.join(", --")}`);
    process.exit(1);
  }
  const agentsFiles = await discoverAgents();
  await syncSymlinks(agentsFiles, selectedAgents, dryRun, verbose);
}
async function cmdSyncCommands(args) {
  const flags = parseFlags(args, ["dry-run", "verbose"]);
  const dryRun = flags["dry-run"];
  const verbose = flags.verbose;
  const agentNames = await getAgentNames();
  const selectedAgents = [];
  for (const agent of agentNames) {
    if (flags[agent]) {
      selectedAgents.push(agent);
    }
  }
  if (selectedAgents.length === 0) {
    console.log(`Please specify at least one agent flag: --${agentNames.join(", --")}`);
    process.exit(1);
  }
  const commandDirs = await discoverCommands();
  await syncCommandFiles(commandDirs, selectedAgents, dryRun, verbose);
}
async function cmdRm(args) {
  const flags = parseFlags(args, ["dry-run", "verbose"]);
  const dryRun = flags["dry-run"];
  const verbose = flags.verbose;
  const agentNames = await getAgentNames();
  const selectedAgents = [];
  for (const agent of agentNames) {
    if (flags[agent]) {
      selectedAgents.push(agent);
    }
  }
  if (selectedAgents.length === 0) {
    console.log(`Please specify at least one agent flag: --${agentNames.join(", --")}`);
    process.exit(1);
  }
  await deleteGuidelineFiles(selectedAgents, dryRun, verbose);
}
async function cmdRmCommands(args) {
  const flags = parseFlags(args, ["dry-run", "verbose"]);
  const dryRun = flags["dry-run"];
  const verbose = flags.verbose;
  const agentNames = await getAgentNames();
  const selectedAgents = [];
  for (const agent of agentNames) {
    if (flags[agent]) {
      selectedAgents.push(agent);
    }
  }
  if (selectedAgents.length === 0) {
    console.log(`Please specify at least one agent flag: --${agentNames.join(", --")}`);
    process.exit(1);
  }
  await deleteCommandFiles(selectedAgents, dryRun, verbose);
}
function parseFlags(args, knownFlags) {
  const flags = {};
  for (const arg of args) {
    if (arg.startsWith("--")) {
      const flag = arg.slice(2);
      flags[flag] = true;
    } else if (arg.startsWith("-")) {
      const flag = arg.slice(1);
      flags[flag] = true;
    }
  }
  return flags;
}
main().catch((err) => {
  console.error("Error:", err);
  process.exit(1);
});
