package ansi

import (
	"strings"
	"unicode"
)

// SanitizeFileLine désinfecte une ligne de code brute pour un affichage terminal sûr.
// Les tabulations sont converties en 4 espaces et les octets de contrôle potentiellement
// perturbateurs pour le terminal (caractères d'échappement ANSI non désirés, NUL, cloche)
// sont éliminés.
func SanitizeFileLine(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\t' {
			b.WriteString("    ")
		} else if r == '\n' || r == '\r' || !unicode.IsControl(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
