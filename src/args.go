package main

import (
	"strings"
)

// ParsedArgs holds the result of parsing command-line arguments
type ParsedArgs struct {
	Flags   map[string]bool
	Unknown []string
	Help    bool
}

// parseArgs parses command-line arguments
func parseArgs(argv []string, allowedFlags map[string]bool) ParsedArgs {
	flags := make(map[string]bool)
	unknown := []string{}
	help := false

	for _, arg := range argv {
		if arg == "" || strings.TrimSpace(arg) == "" {
			continue
		}

		if arg == "--help" || arg == "-h" {
			help = true
			continue
		}

		if strings.HasPrefix(arg, "--") {
			// Long flag
			parts := strings.SplitN(arg, "=", 2)
			flag := parts[0]
			if allowedFlags[flag] {
				flags[flag] = true
			} else {
				unknown = append(unknown, flag)
			}
			continue
		}

		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			// Short flags
			shorts := strings.Split(arg[1:], "")
			for _, short := range shorts {
				flag := "-" + short
				if allowedFlags[flag] {
					flags[flag] = true
				} else {
					unknown = append(unknown, flag)
				}
			}
			continue
		}

		unknown = append(unknown, arg)
	}

	return ParsedArgs{
		Flags:   flags,
		Unknown: unknown,
		Help:    help,
	}
}
