### Key Configuration Options for AI Guideline Files

Research indicates that most of these AI agents support customizable guideline files, often inspired by standards like AGENTS.md, to provide project-specific instructions for coding behaviors, setup, and conventions. However, support varies, with some using tool-specific files and others adopting or falling back to AGENTS.md. Configurations typically allow for both project-level (repository-specific) and user/system-level (global) setups, enabling hierarchical overrides. Below are summaries for each agent based on available documentation.

- **Claude Code**: Uses CLAUDE.md primarily, with support for hierarchical placement; can complement with AGENTS.md for compatibility.
- **Codex**: Relies on AGENTS.md with a clear precedence hierarchy for global and project scopes.
- **Amp Code**: Prefers AGENTS.md but falls back to AGENT.md or CLAUDE.md; supports user-level configs.
- **Opencode**: Employs AGENTS.md at project or global levels, with additional options via opencode.json.
- **Gemini CLI**: Utilizes GEMINI.md or AGENT.md, with global and project-specific placements.
- **Qwen-Code CLI**: Limited direct support found; may use QWEN.md or fall back to AGENTS.md based on integrations, but official docs emphasize general setup over specific guideline files.
- **Warp Terminal**: Employs a Rules system with WARP.md for projects and global rules; can link to other formats like AGENTS.md.
- **Cursor**: Supports AGENTS.md as an alternative to its .cursor/rules directory; includes team-level options.

#### Claude Code Configuration
Place CLAUDE.md at the project root for shared guidelines or in subdirectories for specific contexts. For user-level, use ~/.claude/CLAUDE.md. Edit manually or via commands like /init. See Anthropic's best practices for details. Documentation: [Claude Code Best Practices](https://www.anthropic.com/engineering/claude-code-best-practices).

#### Codex Configuration
AGENTS.md can be at the project root or subdirectories; global in ~/.codex/. Use AGENTS.override.md for temporary changes. Configurable via config.toml for fallbacks. Documentation: [Custom Instructions with AGENTS.md](https://developers.openai.com/codex/guides/agents-md/).

#### Amp Code Configuration
AGENTS.md at project root or subdirs; user-level at ~/.config/amp/ or ~/.config/. Supports YAML front matter for conditional rules. Generate via Amp or edit manually. Documentation: [Amp Owner’s Manual](https://ampcode.com/manual).

#### Opencode Configuration
AGENTS.md at project root or ~/.config/opencode/ for global. Use opencode.json to reference additional files. Initialize with /init. Documentation: [Opencode Rules](https://opencode.ai/docs/rules/).

#### Gemini CLI Configuration
GEMINI.md or AGENT.md at project root/subdirs; global at ~/.gemini/. Supports hierarchical loading. Documentation: [Gemini Code Assist Agent Mode](https://developers.google.com/gemini-code-assist/docs/use-agentic-chat-pair-programmer).

#### Qwen-Code CLI Configuration
Evidence suggests use of QWEN.md for context, potentially aligning with AGENTS.md standards, but official repo focuses on general CLI setup without explicit guideline file details. Place at project root; no clear user-level mentioned. Documentation: [Qwen Code GitHub](https://github.com/QwenLM/qwen-code).

#### Warp Terminal Configuration
Uses Rules via WARP.md at project root/subdirs; global rules in settings. Initialize with /init; link to AGENTS.md if needed. Documentation: [Warp Rules](https://docs.warp.dev/knowledge-and-collaboration/rules).

#### Cursor Configuration
AGENTS.md at project root as alternative to .cursor/rules directory (with RULE.md files). Supports user and team levels via settings/dashboard. Documentation: [Cursor Rules](https://cursor.com/docs/context/rules).

---

### Comprehensive Guide to Configuring AI Guideline Files for Coding Agents

In the evolving landscape of AI-assisted coding, guideline files serve as essential "constitutions" for agents, providing persistent instructions on project setup, coding standards, testing workflows, and behavioral guardrails. These files, often rooted in the open AGENTS.md standard, allow developers to customize AI behaviors at multiple levels—project-specific for tailored repository guidance, and user/system-level for global preferences. This approach ensures consistency across sessions while enabling overrides for subcomponents in complex codebases like monorepos.

The AGENTS.md format, popularized as a "README for agents," has gained traction across tools, with over 60,000 open-source projects adopting it by late 2025. It typically resides at the repository root but supports hierarchical placement, where files in subdirectories or parent paths are loaded based on context. Content focuses on essentials like environment setup, build commands, style guides, and common pitfalls, kept concise to avoid bloating AI context windows.

While some agents adhere strictly to AGENTS.md, others use proprietary equivalents (e.g., CLAUDE.md) with fallbacks for compatibility. Configurations often involve manual editing, auto-generation via commands like /init, and integration with config files (e.g., JSON or TOML) for advanced features like file referencing or conditional application. Below, we detail each agent's approach, drawing from official documentation and best practices.

#### Claude Code: Hierarchical CLAUDE.md with AGENTS.md Compatibility
Claude Code, Anthropic's agentic coding tool, primarily uses CLAUDE.md to inject project-specific knowledge into every conversation. This file acts as a central repository for bash commands, code styles, testing instructions, and setup details, ensuring the agent adheres to team norms.

- **Project Level**: Place CLAUDE.md at the repository root for broad applicability—commit it to Git for sharing or use CLAUDE.local.md (gitignore'd) for personal tweaks. In monorepos, add files in parent directories (e.g., root/CLAUDE.md) or child subdirectories; Claude loads them on demand when working in those areas.
- **User/System Level**: For global rules across all sessions, create ~/.claude/CLAUDE.md.
- **How to Configure**: Run /init in Claude Code to auto-generate the file by scanning the project. Edit manually with emphasis like "IMPORTANT" for key rules, or use the # key to instruct Claude to update it. Avoid negative constraints; provide positive alternatives. For large codebases, keep under 25 KB, focusing on high-usage tools (≥30% of engineers). Complement with AGENTS.md for cross-tool compatibility.
- **Relevant Documentation**: [Claude Code Best Practices](https://www.anthropic.com/engineering/claude-code-best-practices); [Advanced Usage Guide](https://blog.sshh.io/p/how-i-use-every-claude-code-feature).

#### Codex: AGENTS.md with Precedence and Overrides
OpenAI's Codex employs AGENTS.md for layered guidance, discovering files via a strict hierarchy to merge global defaults with project specifics. This ensures proximity-based overrides, where subdirectory rules trump root-level ones.

- **Project Level**: Add AGENTS.md at the repository root for norms like PR rules or security guardrails. Use subdirectories for component-specific instructions, and AGENTS.override.md for temporary overrides.
- **User/System Level**: Global files in ~/.codex/AGENTS.md (or AGENTS.override.md) apply defaults like personal coding preferences.
- **How to Configure**: Codex scans from root to current directory, concatenating files up to 32 KiB (configurable via project_doc_max_bytes). Customize fallbacks in ~/.codex/config.toml (e.g., add "TEAM_GUIDE.md"). Set CODEX_HOME for profiles. Verify with Codex commands or logs.
- **Relevant Documentation**: [Custom Instructions with AGENTS.md](https://developers.openai.com/codex/guides/agents-md/).

#### Amp Code: Flexible AGENTS.md with YAML Enhancements
Amp Code prioritizes AGENTS.md for guiding structure, builds, and conventions, with fallbacks to AGENT.md or CLAUDE.md for migration ease. It supports advanced features like conditional globs.

- **Project Level**: Place at workspace root, current directory, or subtrees; Amp loads based on edited files.
- **User/System Level**: Use $HOME/.config/amp/AGENTS.md or $HOME/.config/AGENTS.md for universal rules.
- **How to Configure**: Generate via Amp if absent, or update with prompts. Use @-mentions for external files (e.g., @doc/*.md) and YAML front matter for globs (e.g., apply to **/*.ts). Migrate from other formats with symlinks. Debug with /agent-files.
- **Relevant Documentation**: [Amp Owner’s Manual](https://ampcode.com/manual).

#### Opencode: AGENTS.md Integrated with JSON Config
Opencode uses AGENTS.md for custom LLM instructions, similar to CLAUDE.md, with options to reference external files.

- **Project Level**: Root AGENTS.md for directory-specific rules.
- **User/System Level**: ~/.config/opencode/AGENTS.md for all sessions.
- **How to Configure**: Initialize with /init. Combine local/global files. Use opencode.json's "instructions" array for extras (e.g., CONTRIBUTING.md). Reference files with @filename, loading lazily.
- **Relevant Documentation**: [Opencode Rules](https://opencode.ai/docs/rules/).

#### Gemini CLI: GEMINI.md or AGENT.md with Contextual Overrides
Google's Gemini CLI supports GEMINI.md for detailed context, falling back to AGENT.md at the root. Files build a memory system with specificity overriding generality.

- **Project Level**: Root or subdirectories; parent files apply to subdirs.
- **User/System Level**: ~/.gemini/GEMINI.md for all projects.
- **How to Configure**: Write in Markdown; use @FILENAME for extras. Load hierarchy from working directory up.
- **Relevant Documentation**: [Gemini Code Assist Agent Mode](https://developers.google.com/gemini-code-assist/docs/use-agentic-chat-pair-programmer).

#### Qwen-Code CLI: QWEN.md or AGENTS.md Alignment
Alibaba's Qwen-Code CLI has sparse official details on guideline files, but integrations suggest QWEN.md for context, potentially compatible with AGENTS.md standards. Focus is on CLI optimization for Qwen3-Coder models.

- **Project Level**: Place QWEN.md at root; no subdir hierarchy specified.
- **User/System Level**: Not explicitly documented; may rely on global config.
- **How to Configure**: Set up via MCP server and context files; align with AGENTS.md for cross-tool use.
- **Relevant Documentation**: [Qwen Code GitHub](https://github.com/QwenLM/qwen-code).

#### Warp Terminal: Rules System with WARP.md
Warp's AI features use a Rules system for guidelines, with WARP.md as the project file; it can link to AGENTS.md. Rules ensure tailored responses without core changes.

- **Project Level**: WARP.md at root or subdirs; auto-applies based on directory.
- **User/System Level**: Global rules via settings or Warp Drive.
- **How to Configure**: Create/edit via /init, Slash Commands, or Drive. Precedence: subdir > root > global.
- **Relevant Documentation**: [Warp Rules](https://docs.warp.dev/knowledge-and-collaboration/rules).

#### Cursor: AGENTS.md as Alternative to Structured Rules
Cursor supports AGENTS.md for simple instructions, but prefers .cursor/rules for metadata-rich rules. It includes team enforcement.

- **Project Level**: AGENTS.md at root/subdirs or .cursor/rules with RULE.md.
- **User/System Level**: Settings for user rules; dashboard for team.
- **How to Configure**: Use globs/alwaysApply in frontmatter; precedence: team > project > user. Migrate from .cursorrules.
- **Relevant Documentation**: [Cursor Rules](https://cursor.com/docs/context/rules).

#### Comparison of Configuration Approaches
To highlight differences, the following table summarizes key aspects across agents:

| Agent          | Primary File     | Project Level Placement | User/System Level Placement | Hierarchy/Overrides | Auto-Generation | External File Referencing | Documentation Link |
|----------------|------------------|-------------------------|-----------------------------|---------------------|-----------------|---------------------------|--------------------|
| Claude Code   | CLAUDE.md       | Root, parents, children | ~/.claude/CLAUDE.md        | Yes, on-demand     | /init          | Limited                  | [Link](https://www.anthropic.com/engineering/claude-code-best-practices) |
| Codex         | AGENTS.md       | Root, subdirs           | ~/.codex/AGENTS.md         | Precedence merge   | No             | Via config.toml          | [Link](https://developers.openai.com/codex/guides/agents-md/) |
| Amp Code      | AGENTS.md       | Root, CWD, subtrees     | ~/.config/amp/AGENTS.md    | Yes, conditional   | Yes            | @-mentions, globs        | [Link](https://ampcode.com/manual) |
| Opencode      | AGENTS.md       | Root                    | ~/.config/opencode/AGENTS.md | Combine local/global | /init        | opencode.json array      | [Link](https://opencode.ai/docs/rules/) |
| Gemini CLI    | GEMINI.md/AGENT.md | Root, subdirs, parents | ~/.gemini/GEMINI.md        | Specificity overrides | No          | @FILENAME                | [Link](https://developers.google.com/gemini-code-assist/docs/use-agentic-chat-pair-programmer) |
| Qwen-Code CLI | QWEN.md         | Root                    | Not specified              | Limited            | No             | Via integrations         | [Link](https://github.com/QwenLM/qwen-code) |
| Warp Terminal | WARP.md (Rules) | Root, subdirs           | Settings/Warp Drive        | Subdir > root > global | /init       | Linking (e.g., to AGENTS.md) | [Link](https://docs.warp.dev/knowledge-and-collaboration/rules) |
| Cursor        | AGENTS.md or .cursor/rules | Root, subdirs        | Settings (user), Dashboard (team) | Team > project > user | Command     | Remote GitHub, plugins   | [Link](https://cursor.com/docs/context/rules) |

This table underscores the trend toward standardization while accommodating tool-specific needs. For best results, start with auto-generation where available, then refine based on agent feedback.

In practice, adopting these files streamlines AI integration, reducing errors and enhancing productivity. For cross-agent compatibility, prioritize AGENTS.md and use symlinks for aliases.

### Key Citations
- [Claude Code: Best practices for agentic coding](https://www.anthropic.com/engineering/claude-code-best-practices)
- [Improve your AI code output with AGENTS.md](https://www.builder.io/blog/agents-md)
- [AGENTS.md](https://agents.md/)
- [How I Use Every Claude Code Feature](https://blog.sshh.io/p/how-i-use-every-claude-code-feature)
- [Custom instructions with AGENTS.md](https://developers.openai.com/codex/guides/agents-md/)
- [Owner's Manual - Amp Code](https://ampcode.com/manual)
- [Rules | OpenCode](https://opencode.ai/docs/rules/)
- [Config | OpenCode](https://opencode.ai/docs/config/)
- [Gemini CLI configuration](https://geminicli.com/docs/get-started/configuration/)
- [Use the Gemini Code Assist agent mode](https://developers.google.com/gemini-code-assist/docs/use-agentic-chat-pair-programmer)
- [Qwen Code - ToolUniverse Documentation](https://zitniklab.hms.harvard.edu/ToolUniverse/guide/building_ai_scientists/qwen_code.html)
- [Qwen Code GitHub](https://github.com/QwenLM/qwen-code)
- [Rules - Warp documentation](https://docs.warp.dev/knowledge-and-collaboration/rules)
- [Rules | Cursor Docs](https://cursor.com/docs/context/rules)
