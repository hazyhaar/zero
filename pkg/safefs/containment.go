package safefs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InBoundsNonWrapping vérifie de manière non débordante qu'une plage mémoire
// [offset, offset+len) est strictement contenue dans la limite [0, limit).
func InBoundsNonWrapping(offset, length, limit uint64) bool {
	return offset < limit && length <= limit-offset
}

// InRootBounds valide qu'un chemin cible résolu se situe physiquement sous la racine donnée.
func InRootBounds(root, target string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("safefs: resolve root: %w", err)
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("safefs: resolve target: %w", err)
	}
	rel, err := filepath.Rel(absRoot, absTarget)
	if err != nil || strings.HasPrefix(rel, "..") || rel == ".." {
		return "", fmt.Errorf("safefs: path escape: %s hors de la racine %s", target, root)
	}
	return absTarget, nil
}

// WriteFileAtomic écrit un fichier de manière atomique via un fichier temporaire
// suivi d'un renommage atomique, interdisant aux lecteurs concurrents de voir des écritures partielles.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("safefs: create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("safefs: write: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("safefs: chmod: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("safefs: close: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("safefs: rename: %w", err)
	}
	return nil
}
