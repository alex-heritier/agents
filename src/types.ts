export type FileSpec = {
  file: string;
  dir: string;
};

export type ProviderConfig = {
  name?: string;
  guidelines?: FileSpec;
  commands?: FileSpec;
};

export type SourcesConfig = {
  guidelines: string;
  commands: string;
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
