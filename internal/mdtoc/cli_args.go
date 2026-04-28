package mdtoc

import (
	"fmt"
	"strings"
)

// inputSourceKind distinguishes the three CLI input states the runner cares
// about after argument normalization.
type inputSourceKind int

const (
	// inputSourceNone means that the CLI invocation did not specify a file and
	// stdin was not yet considered.
	inputSourceNone inputSourceKind = iota
	// inputSourceFile means the invocation explicitly named a file input source.
	inputSourceFile
	// inputSourceStdin means the invocation should read document content from
	// stdin and write transformed output to stdout.
	inputSourceStdin
)

// inputSource carries the normalized document input selected from argv and
// later finalized with knowledge about whether stdin is interactive.
type inputSource struct {
	kind    inputSourceKind
	file    string
	viaFlag bool
}

// argumentSpec declares which flags consume a following value and which flags
// are boolean toggles for one CLI parsing mode.
type argumentSpec struct {
	valueFlags map[string]bool
	boolFlags  map[string]bool
}

// normalizedCLIArgs is the result of the preprocessing pass that separates the
// single allowed input source from the remaining flags that can still be handed
// to a flag.FlagSet.
type normalizedCLIArgs struct {
	parseArgs []string
	input     inputSource
}

// generateCommandArgSpec returns the accepted flag shape for `generate`.
func generateCommandArgSpec() argumentSpec {
	return argumentSpec{
		valueFlags: map[string]bool{
			"--numbering": true,
			"-n":          true,
			"--min-level": true,
			"--max-level": true,
			"--anchor":    true,
			"-a":          true,
			"--toc":       true,
			"--bullets":   true,
			"-b":          true,
			"--file":      true,
			"-f":          true,
		},
		boolFlags: map[string]bool{
			"--verbose": true,
			"-v":        true,
			"--help":    true,
			"-h":        true,
		},
	}
}

// regenCommandArgSpec returns the accepted flag shape for `regen`.
func regenCommandArgSpec() argumentSpec {
	return argumentSpec{
		valueFlags: map[string]bool{
			"--file": true,
			"-f":     true,
		},
		boolFlags: map[string]bool{
			"--verbose": true,
			"-v":        true,
			"--help":    true,
			"-h":        true,
		},
	}
}

// stripCommandArgSpec returns the accepted flag shape for `strip`.
func stripCommandArgSpec() argumentSpec {
	return argumentSpec{
		valueFlags: map[string]bool{
			"--file": true,
			"-f":     true,
		},
		boolFlags: map[string]bool{
			"--raw":     true,
			"--verbose": true,
			"-v":        true,
			"--help":    true,
			"-h":        true,
		},
	}
}

// checkCommandArgSpec returns the accepted flag shape for `check`.
func checkCommandArgSpec() argumentSpec {
	return argumentSpec{
		valueFlags: map[string]bool{
			"--file": true,
			"-f":     true,
		},
		boolFlags: map[string]bool{
			"--verbose": true,
			"-v":        true,
			"--help":    true,
			"-h":        true,
		},
	}
}

// rootCommandArgSpec returns the top-level flag shape. It intentionally accepts
// the generate control flags because root convenience mode may dispatch to
// `generate`.
func rootCommandArgSpec() argumentSpec {
	return argumentSpec{
		valueFlags: map[string]bool{
			"--numbering": true,
			"-n":          true,
			"--min-level": true,
			"--max-level": true,
			"--anchor":    true,
			"-a":          true,
			"--toc":       true,
			"--bullets":   true,
			"-b":          true,
			"--file":      true,
			"-f":          true,
		},
		boolFlags: map[string]bool{
			"--help":    true,
			"-h":        true,
			"--version": true,
			"--verbose": true,
			"-v":        true,
		},
	}
}

// normalizeCLIArgs performs the input-source preprocessing needed for issue
// #53. The key job here is to make positional files and `--file/-f`
// interchangeable while still leaving a clean flag-only argument list for the
// existing flag.FlagSet based parsers.
func normalizeCLIArgs(args []string, spec argumentSpec) (normalizedCLIArgs, error) {
	normalized := normalizedCLIArgs{
		parseArgs: make([]string, 0, len(args)),
	}
	positionalFiles := []string{}
	fileValues := []string{}

	for i := 0; i < len(args); i++ {
		token := args[i]

		if token == "--" {
			positionalFiles = append(positionalFiles, args[i+1:]...)
			break
		}

		if flagName, hasInlineValue, inlineValue, ok := classifyValueFlagToken(token, spec); ok {
			if flagName == "--file" || flagName == "-f" {
				value, nextIndex, err := resolveFlagValue(args, i, hasInlineValue, inlineValue)
				if err != nil {
					return normalizedCLIArgs{}, err
				}
				fileValues = append(fileValues, value)
				i = nextIndex
				continue
			}

			normalized.parseArgs = append(normalized.parseArgs, token)
			if !hasInlineValue {
				if i+1 >= len(args) {
					return normalizedCLIArgs{}, fmt.Errorf("flag needs an argument: %s", token)
				}
				normalized.parseArgs = append(normalized.parseArgs, args[i+1])
				i++
			}
			continue
		}

		if isKnownBoolFlagToken(token, spec) {
			normalized.parseArgs = append(normalized.parseArgs, token)
			continue
		}

		if strings.HasPrefix(token, "-") {
			normalized.parseArgs = append(normalized.parseArgs, token)
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				normalized.parseArgs = append(normalized.parseArgs, args[i+1])
				i++
			}
			continue
		}

		positionalFiles = append(positionalFiles, token)
	}

	if len(positionalFiles) > 1 {
		return normalizedCLIArgs{}, fmt.Errorf("provide exactly one input source: positional file, --file, or stdin")
	}
	if len(fileValues) > 1 {
		return normalizedCLIArgs{}, fmt.Errorf("provide exactly one input source: positional file, --file, or stdin")
	}
	if len(positionalFiles) == 1 && len(fileValues) == 1 {
		return normalizedCLIArgs{}, fmt.Errorf("provide exactly one input source: positional file, --file, or stdin")
	}

	switch {
	case len(positionalFiles) == 1:
		normalized.input = inputSource{kind: inputSourceFile, file: positionalFiles[0]}
	case len(fileValues) == 1:
		normalized.input = inputSource{kind: inputSourceFile, file: fileValues[0], viaFlag: true}
	default:
		normalized.input = inputSource{kind: inputSourceNone}
	}

	return normalized, nil
}

// classifyValueFlagToken recognizes known value flags in both `--flag value`
// and `--flag=value` spellings.
func classifyValueFlagToken(token string, spec argumentSpec) (string, bool, string, bool) {
	if spec.valueFlags[token] {
		return token, false, "", true
	}
	if strings.HasPrefix(token, "-") {
		if flagName, value, ok := splitInlineFlagToken(token); ok && spec.valueFlags[flagName] {
			return flagName, true, value, true
		}
	}
	return "", false, "", false
}

// isKnownBoolFlagToken reports whether the token names one of the boolean flags
// supported in the active parsing mode.
func isKnownBoolFlagToken(token string, spec argumentSpec) bool {
	return spec.boolFlags[token]
}

// splitInlineFlagToken extracts `--flag` and `value` from `--flag=value`.
func splitInlineFlagToken(token string) (string, string, bool) {
	if !strings.HasPrefix(token, "-") {
		return "", "", false
	}
	parts := strings.SplitN(token, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// resolveFlagValue returns the effective value for a known value flag and the
// updated loop index to continue scanning from.
func resolveFlagValue(args []string, index int, hasInlineValue bool, inlineValue string) (string, int, error) {
	if hasInlineValue {
		return inlineValue, index, nil
	}
	if index+1 >= len(args) {
		return "", index, fmt.Errorf("flag needs an argument: %s", args[index])
	}
	return args[index+1], index + 1, nil
}

// hasExplicitGenerateControlFlag reports whether the root invocation explicitly
// requested one of the generate-shaping flags that force root dispatch to
// `generate`.
func hasExplicitGenerateControlFlag(args []string) bool {
	for _, token := range args {
		flagName := token
		if splitName, _, ok := splitInlineFlagToken(token); ok {
			flagName = splitName
		}
		flagName = canonicalGenerateControlFlag(flagName)
		switch flagName {
		case "--numbering", "-n", "--min-level", "--max-level", "--anchor", "-a", "--toc", "--bullets", "-b":
			return true
		}
	}
	return false
}

// canonicalGenerateControlFlag normalizes the Go flag package's accepted
// one-dash long-flag spellings such as `-toc` to their canonical double-dash
// form so root-mode dispatch sees the same override semantics as flag parsing.
func canonicalGenerateControlFlag(flagName string) string {
	switch flagName {
	case "-numbering":
		return "--numbering"
	case "-min-level":
		return "--min-level"
	case "-max-level":
		return "--max-level"
	case "-anchor":
		return "--anchor"
	case "-toc":
		return "--toc"
	case "-bullets":
		return "--bullets"
	default:
		return flagName
	}
}

// isVerboseOnlyInvocation reports whether the normalized argument list contains
// only the verbose flag spelling supported by the current command.
func isVerboseOnlyInvocation(args []string) bool {
	return len(args) == 1 && (args[0] == "--verbose" || args[0] == "-v")
}

// hasFlag reports whether the exact token is present in the argument list. The
// helper remains intentionally small because tests still use it to probe basic
// flag-token behavior directly.
func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}
