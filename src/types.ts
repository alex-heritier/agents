export interface GuidelineFile {
  path: string;      // full path
  dir: string;       // directory containing the file
  agent: string;     // AGENTS, CLAUDE, CURSOR
  file: string;      // filename
  isSymlink: boolean;
  size: number;      // file size in bytes
}
