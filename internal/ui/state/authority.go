package state

import (
	"sync"
	"sync/atomic"
)

// MonotonicAuthority centralise la gestion des numéros de séquence et de révision
// de sources pour l'ensemble des requêtes asynchrones du TUI.
type MonotonicAuthority struct {
	mu            sync.RWMutex
	liveSeq       atomic.Uint64
	lifetimeToken atomic.Uint64
	pathRevisions map[string]uint64
	generation    atomic.Uint64
}

// NewMonotonicAuthority instancie un gestionnaire d'autorité scellé.
func NewMonotonicAuthority() *MonotonicAuthority {
	m := &MonotonicAuthority{
		pathRevisions: make(map[string]uint64),
	}
	m.lifetimeToken.Store(1)
	m.generation.Store(1)
	return m
}

// NextLifetimeToken incrémente et retourne un nouveau jeton de vie de vue.
func (m *MonotonicAuthority) NextLifetimeToken() uint64 {
	return m.lifetimeToken.Add(1)
}

// NextSeq incrémente et retourne le numéro de séquence actif.
func (m *MonotonicAuthority) NextSeq() uint64 {
	return m.liveSeq.Add(1)
}

// CurrentSeq retourne le numéro de séquence actif sans mutation.
func (m *MonotonicAuthority) CurrentSeq() uint64 {
	return m.liveSeq.Load()
}

// IsSuperseded vérifie si une requête émise avec une séquence passée a été supplantée.
func (m *MonotonicAuthority) IsSuperseded(requestSeq uint64) bool {
	return m.liveSeq.Load() != requestSeq
}

// InvalidatePath incrémente la révision minimale requise pour un chemin donné
// suite à une mutation sur disque (outil, commande bash, sweep git, /rewind).
func (m *MonotonicAuthority) InvalidatePath(path string) uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pathRevisions[path]++
	return m.pathRevisions[path]
}

// RequiredRevision retourne la révision requise pour un chemin donné.
func (m *MonotonicAuthority) RequiredRevision(path string) uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pathRevisions[path]
}

// InvalidateAll incrémente la génération globale du cache (ex. lors d'un changement de thème).
func (m *MonotonicAuthority) InvalidateAll() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	clear(m.pathRevisions)
	return m.generation.Add(1)
}

// Generation retourne la génération courante du cache.
func (m *MonotonicAuthority) Generation() uint64 {
	return m.generation.Load()
}
