/**
 * Type definitions for the agents CLI
 */

export interface GuidelineConfig {
  file: string;
  dir: string;
  source: string;
  global_paths: string[];
}

export interface CommandsConfig {
  dir: string;
  extension: string;
  source_dir: string;
  global_paths: string[];
}

export interface Provider {
  name: string;
  display_name: string;
  guideline: GuidelineConfig;
  commands: CommandsConfig;
}

export interface Config {
  version: string;
  providers: Record<string, Provider>;
}

export interface AgentConfig {
  name: string;
  file: string;
  dir: string;
}

export interface GuidelineFile {
  path: string;
  dir: string;
  agent: string;
  file: string;
  isSymlink: boolean;
  size: number;
}

export interface CommandFile {
  path: string;
  dir: string;
  provider: string;
  name: string;
  file: string;
  isSymlink: boolean;
  size: number;
}
