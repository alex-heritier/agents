export type FileSpec = {
  file: string;
  dir: string;
};

export type ProviderConfig = {
  name?: string;
  guidelines?: FileSpec;
  commands?: FileSpec;
  skills?: FileSpec;
};

export type SourcesConfig = {
  guidelines: string;
  commands: string;
  skills: string;
};

export type ProvidersConfig = {
  sources: SourcesConfig;
  globalGuidelines: string[];
  providers: Record<string, ProviderConfig>;
};

export type ManagedFile = {
  path: string;
  dir: string;
  agent: string;
  file: string;
  isSymlink: boolean;
  size: number;
};

export type SkillMetadata = {
  name: string;
  description: string;
  license?: string;
  allowedTools?: string;
  metadata?: Record<string, unknown>;
};

export type SkillFile = {
  path: string;
  dir: string;
  skillName: string;
  location: "global" | "project";
  metadata?: SkillMetadata;
  error?: string;
};
