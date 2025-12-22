export type ParsedArgs = {
  flags: Set<string>;
  unknown: string[];
  help: boolean;
};

export function parseArgs(argv: string[], allowedFlags: Set<string>): ParsedArgs {
  const flags = new Set<string>();
  const unknown: string[] = [];
  let help = false;

  for (const arg of argv) {
    if (typeof arg !== "string" || arg.trim() === "") {
      continue;
    }

    if (arg === "--help" || arg === "-h") {
      help = true;
      continue;
    }

    if (arg.startsWith("--")) {
      const [flag] = arg.split("=", 1);
      if (allowedFlags.has(flag)) {
        flags.add(flag);
      } else {
        unknown.push(flag);
      }
      continue;
    }

    if (arg.startsWith("-") && arg.length > 1) {
      const shorts = arg.slice(1).split("");
      for (const short of shorts) {
        const flag = `-${short}`;
        if (allowedFlags.has(flag)) {
          flags.add(flag);
        } else {
          unknown.push(flag);
        }
      }
      continue;
    }

    unknown.push(arg);
  }

  return { flags, unknown, help };
}
