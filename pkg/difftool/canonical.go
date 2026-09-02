package difftool

import (
	"strings"
	"unicode"
)

// CanonicalLineKey transforme une ligne brute ou diff en une clé canonique d'appariement.
// Elle convertit les tabulations en 4 espaces (alignée sur la désinfection d'affichage),
// supprime les caractères de contrôle non imprimables et tronque les espaces d'extrémité.
func CanonicalLineKey(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\t' {
			b.WriteString("    ")
		} else if !unicode.IsControl(r) {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

// MatchChangedLines compare une liste ordonnée de lignes de document avec un ensemble
// de lignes modifiées (provenant d'un patch unifié ou d'un outil d'édition) et retourne
// une carte booléenne indiquant les index de lignes (0-indexés) modifiés.
func MatchChangedLines(lines []string, changed []string) map[int]bool {
	if len(lines) == 0 || len(changed) == 0 {
		return nil
	}
	changedSet := make(map[string]struct{}, len(changed))
	for _, c := range changed {
		k := CanonicalLineKey(c)
		if k != "" {
			changedSet[k] = struct{}{}
		}
	}
	if len(changedSet) == 0 {
		return nil
	}
	matches := make(map[int]bool, len(lines))
	for i, line := range lines {
		k := CanonicalLineKey(line)
		if k != "" {
			if _, ok := changedSet[k]; ok {
				matches[i] = true
			}
		}
	}
	return matches
}
