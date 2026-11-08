package desktop

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Entry is the parsed [Desktop Entry] group of a .desktop file, see https://specifications.freedesktop.org/desktop-entry-spec/latest/
type Entry struct {
	ID          string // desktop file id, e.g. org.gnome.Nautilus.desktop
	Path        string // absolute path of the .desktop file
	Type        string
	Name        string
	GenericName string
	Comment     string
	Icon        string
	Exec        string // raw exec value, see ParseExec / Command
	TryExec     string
	WorkDir     string // Path key
	Terminal    bool
	NoDisplay   bool
	Hidden      bool
	Categories  []string
	Keywords    []string
	OnlyShowIn  []string
	NotShowIn   []string
}

// ParseEntry parses the [Desktop Entry] group, localized keys are resolved using the given locale candidates (most specific first)
func ParseEntry(r io.Reader, locales []string) (*Entry, error) {
	values := map[string]string{}
	localized := map[string]map[string]string{}

	inEntry := false
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inEntry = line == "[Desktop Entry]"
			continue
		}
		if !inEntry {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if base, locale, isLocalized := strings.Cut(key, "["); isLocalized {
			locale = strings.TrimSuffix(locale, "]")
			if localized[base] == nil {
				localized[base] = map[string]string{}
			}
			localized[base][locale] = value
			continue
		}
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	get := func(key string) string {
		for _, l := range locales {
			if v, ok := localized[key][l]; ok {
				return v
			}
		}
		return values[key]
	}

	return &Entry{
		Type:        values["Type"],
		Name:        unescapeString(get("Name")),
		GenericName: unescapeString(get("GenericName")),
		Comment:     unescapeString(get("Comment")),
		Icon:        unescapeString(get("Icon")),
		Exec:        unescapeString(values["Exec"]),
		TryExec:     unescapeString(values["TryExec"]),
		WorkDir:     unescapeString(values["Path"]),
		Terminal:    values["Terminal"] == "true",
		NoDisplay:   values["NoDisplay"] == "true",
		Hidden:      values["Hidden"] == "true",
		Categories:  splitList(values["Categories"]),
		Keywords:    splitList(get("Keywords")),
		OnlyShowIn:  splitList(values["OnlyShowIn"]),
		NotShowIn:   splitList(values["NotShowIn"]),
	}, nil
}

// ParseFile parses a .desktop file from disk
func ParseFile(path string, locales []string) (*Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	entry, err := ParseEntry(f, locales)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	entry.Path = path
	return entry, nil
}

// unescapeString resolves the escape sequences of the string value type (\s \n \t \r \\)
func unescapeString(value string) string {
	if !strings.Contains(value, `\`) {
		return value
	}

	var b strings.Builder
	for i := 0; i < len(value); i++ {
		if value[i] == '\\' && i+1 < len(value) {
			i++
			switch value[i] {
			case 's':
				b.WriteByte(' ')
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case '\\':
				b.WriteByte('\\')
			default:
				b.WriteByte('\\')
				b.WriteByte(value[i])
			}
			continue
		}
		b.WriteByte(value[i])
	}
	return b.String()
}

// splitList splits a ;-separated list value, \; escapes a literal semicolon
func splitList(value string) []string {
	if value == "" {
		return nil
	}

	var result []string
	var cur strings.Builder
	for i := 0; i < len(value); i++ {
		switch {
		case value[i] == '\\' && i+1 < len(value) && value[i+1] == ';':
			cur.WriteByte(';')
			i++
		case value[i] == ';':
			if s := strings.TrimSpace(cur.String()); s != "" {
				result = append(result, unescapeString(s))
			}
			cur.Reset()
		default:
			cur.WriteByte(value[i])
		}
	}
	if s := strings.TrimSpace(cur.String()); s != "" {
		result = append(result, unescapeString(s))
	}
	return result
}

// ParseExec tokenizes the Exec value and expands field codes, file/url arguments (%f %F %u %U) are dropped as no files are passed
func ParseExec(exec string, name string, icon string, path string) ([]string, error) {
	// tokenize, double quotes group arguments and \ escapes " ` $ \ within quotes
	var tokens []string
	var cur strings.Builder
	inQuote, hasToken := false, false
	for i := 0; i < len(exec); i++ {
		c := exec[i]
		switch {
		case inQuote && c == '\\' && i+1 < len(exec) && strings.IndexByte("\"`$\\", exec[i+1]) >= 0:
			cur.WriteByte(exec[i+1])
			i++
		case c == '"':
			inQuote = !inQuote
			hasToken = true
		case !inQuote && (c == ' ' || c == '\t'):
			if hasToken {
				tokens = append(tokens, cur.String())
				cur.Reset()
				hasToken = false
			}
		default:
			cur.WriteByte(c)
			hasToken = true
		}
	}
	if inQuote {
		return nil, fmt.Errorf("unterminated quote in exec value: %s", exec)
	}
	if hasToken {
		tokens = append(tokens, cur.String())
	}

	// expand field codes
	var args []string
	for _, token := range tokens {
		if token == "%i" {
			if icon != "" {
				args = append(args, "--icon", icon)
			}
			continue
		}

		var b strings.Builder
		onlyFieldCodes := token != ""
		for i := 0; i < len(token); i++ {
			if token[i] != '%' || i+1 >= len(token) {
				b.WriteByte(token[i])
				onlyFieldCodes = false
				continue
			}
			i++
			switch token[i] {
			case '%':
				b.WriteByte('%')
				onlyFieldCodes = false
			case 'c':
				b.WriteString(name)
				onlyFieldCodes = false
			case 'k':
				b.WriteString(path)
				onlyFieldCodes = false
			default:
				// %f %F %u %U and deprecated %d %D %n %N %v %m expand to nothing
			}
		}
		if onlyFieldCodes && b.Len() == 0 {
			continue
		}
		args = append(args, b.String())
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("empty exec value")
	}
	return args, nil
}

// ShellCommand joins the arguments into a command line that is safe to pass to sh -c
func ShellCommand(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		quoted[i] = shellQuoteIfNeeded(a)
	}
	return strings.Join(quoted, " ")
}

func shellQuoteIfNeeded(value string) string {
	if value == "" {
		return "''"
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("_@%+=:,./-", r)) {
			return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
		}
	}
	return value
}

// localeCandidates returns the locale keys to try for localized values, most specific first (lang_COUNTRY@MODIFIER, lang_COUNTRY, lang@MODIFIER, lang)
func localeCandidates() []string {
	var locale string
	for _, env := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := os.Getenv(env); v != "" {
			locale = v
			break
		}
	}
	locale, modifier, _ := strings.Cut(locale, "@")
	locale, _, _ = strings.Cut(locale, ".") // strip encoding
	if locale == "" || locale == "C" || locale == "POSIX" {
		return nil
	}
	lang, country, _ := strings.Cut(locale, "_")

	var result []string
	if country != "" && modifier != "" {
		result = append(result, lang+"_"+country+"@"+modifier)
	}
	if country != "" {
		result = append(result, lang+"_"+country)
	}
	if modifier != "" {
		result = append(result, lang+"@"+modifier)
	}
	return append(result, lang)
}
