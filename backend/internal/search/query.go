package search

import (
	"strings"
	"unicode"
)

// Tokens splits q on whitespace after trimming.
func Tokens(q string) []string {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil
	}
	var out []string
	for _, f := range strings.Fields(q) {
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

// EscapeFtsToken wraps a token for FTS5 phrase/prefix queries (double-quote escaping).
func EscapeFtsToken(s string) string {
	s = strings.TrimSpace(s)
	return strings.ReplaceAll(s, `"`, `""`)
}

// BuildPrefixMatchOR builds MATCH for prefix on all columns: ("tok1*" OR "tok2*").
func BuildPrefixMatchOR(tokens []string) string {
	if len(tokens) == 0 {
		return ""
	}
	parts := make([]string, 0, len(tokens))
	for _, t := range tokens {
		e := EscapeFtsToken(t)
		if e == "" {
			continue
		}
		parts = append(parts, `"`+e+`*"`)
	}
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

// BuildTrigramMatchOR builds MATCH for substring search (trigram tokenizer): ("tok1" OR "tok2").
// Tokens shorter than minRunes are skipped (trigram is ineffective for very short needles).
func BuildTrigramMatchOR(tokens []string, minRunes int) string {
	if minRunes < 1 {
		minRunes = 3
	}
	var parts []string
	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if len([]rune(t)) < minRunes {
			continue
		}
		e := EscapeFtsToken(t)
		if e == "" {
			continue
		}
		parts = append(parts, `"`+e+`"`)
	}
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

// NormalizeFold lowercases and trims; used for fuzzy comparison.
func NormalizeFold(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

// IsAlnumWord reports whether s is a single token suitable for strict fuzzy (letters/digits/hyphen).
func IsAlnumWord(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}
