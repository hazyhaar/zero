package vtparser

import (
	"bytes"
	"testing"
)

func TestOSC52_Clipboard(t *testing.T) {
	g := &CursorGrid{}
	g.Reset(80, 24)
	p := &Parser{}
	p.Reset(g)

	var wroteTarget string
	var wroteData []byte
	var queryTarget string

	p.OnClipboardWrite = func(target string, data []byte) {
		wroteTarget = target
		wroteData = data
	}
	p.OnClipboardQuery = func(target string) {
		queryTarget = target
	}

	// 1. Écriture presse-papier standard (OSC 52 ; c ; Base64 BEL)
	// "Hello World" -> "SGVsbG8gV29ybGQ="
	p.Feed([]byte("\x1b]52;c;SGVsbG8gV29ybGQ=\x07"))
	if wroteTarget != "c" || string(wroteData) != "Hello World" {
		t.Fatalf("OSC 52 write mismatch: target=%q data=%q", wroteTarget, string(wroteData))
	}

	// 2. Écriture presse-papier primary avec ST (\x1b\\)
	// "Alpha-55" -> "QWxwaGEtNTU="
	p.Feed([]byte("\x1b]52;p;QWxwaGEtNTU=\x1b\\"))
	if wroteTarget != "p" || string(wroteData) != "Alpha-55" {
		t.Fatalf("OSC 52 primary write mismatch: target=%q data=%q", wroteTarget, string(wroteData))
	}

	// 3. Requête presse-papier (query ?)
	p.Feed([]byte("\x1b]52;c;?\x07"))
	if queryTarget != "c" {
		t.Fatalf("OSC 52 query mismatch: target=%q, want 'c'", queryTarget)
	}

	// 4. Flux fragmenté sur frontière de blocs
	wroteTarget = ""
	wroteData = nil
	p.Feed([]byte("\x1b]52;c;SGVs"))
	p.Feed([]byte("bG8=\x07")) // "Hello"
	if wroteTarget != "c" || string(wroteData) != "Hello" {
		t.Fatalf("OSC 52 fragmented write mismatch: target=%q data=%q", wroteTarget, string(wroteData))
	}
}

func TestOSC8_Hyperlinks(t *testing.T) {
	g := &CursorGrid{}
	g.Reset(80, 24)
	p := &Parser{}
	p.Reset(g)

	var lastID, lastURL string
	var lastActive bool
	var eventCount int

	p.OnHyperlink = func(id, url string, active bool) {
		lastID = id
		lastURL = url
		lastActive = active
		eventCount++
	}

	// 1. Ouverture d'un hyperlien avec ID explicite
	p.Feed([]byte("\x1b]8;id=doc123:env=prod;https://example.com/spec\x1b\\"))
	if !lastActive || lastID != "doc123" || lastURL != "https://example.com/spec" || p.CurLinkID != "doc123" || p.CurLinkURL != "https://example.com/spec" {
		t.Fatalf("OSC 8 open mismatch: active=%v id=%q url=%q", lastActive, lastID, lastURL)
	}

	// 2. Écriture du texte ancré dans la grille
	p.Feed([]byte("LIEN_DOCUMENT"))
	if cellAt(g, 0, 0).Rune != 'L' || cellAt(g, 4, 0).Rune != '_' {
		t.Fatalf("texte ancré non dessiné dans la grille")
	}

	// 3. Fermeture de l'hyperlien
	p.Feed([]byte("\x1b]8;;\x1b\\"))
	if lastActive || p.CurLinkID != "" || p.CurLinkURL != "" {
		t.Fatalf("OSC 8 close mismatch: active=%v id=%q url=%q", lastActive, p.CurLinkID, p.CurLinkURL)
	}

	if eventCount != 2 {
		t.Fatalf("nombre d'évènements OSC 8 attendu: 2, got: %d", eventCount)
	}
}

func TestDEC2026_SynchronizedUpdates(t *testing.T) {
	g := &CursorGrid{}
	g.Reset(80, 24)
	p := &Parser{}
	p.Reset(g)

	var syncStates []bool
	p.OnSyncUpdate = func(active bool) {
		syncStates = append(syncStates, active)
	}

	// 1. Début de trame atomique (CSI ? 2026 h)
	p.Feed([]byte("\x1b[?2026h"))
	if !p.SyncUpdate {
		t.Fatalf("p.SyncUpdate attendu true")
	}

	// 2. Fin de trame atomique (CSI ? 2026 l)
	p.Feed([]byte("\x1b[?2026l"))
	if p.SyncUpdate {
		t.Fatalf("p.SyncUpdate attendu false")
	}

	if len(syncStates) != 2 || !syncStates[0] || syncStates[1] {
		t.Fatalf("historique des états sync invalide: %v", syncStates)
	}

	// 3. Test de helper d'encodage
	startSeq := EncodeSyncUpdate(true)
	endSeq := EncodeSyncUpdate(false)
	if !bytes.Equal(startSeq, []byte("\x1b[?2026h")) || !bytes.Equal(endSeq, []byte("\x1b[?2026l")) {
		t.Fatalf("EncodeSyncUpdate mismatch: start=%q end=%q", startSeq, endSeq)
	}
}

func TestSGR1006_MouseReporting(t *testing.T) {
	g := &CursorGrid{}
	g.Reset(80, 24)
	p := &Parser{}
	p.Reset(g)

	// 1. Activation mode souris bouton + format SGR étendu
	p.Feed([]byte("\x1b[?1002;1006h"))
	if p.MouseMode != MouseModeButton || !p.MouseSGR {
		t.Fatalf("modes souris mismatch: mode=%d, sgr=%v", p.MouseMode, p.MouseSGR)
	}

	// 2. Désactivation mode souris
	p.Feed([]byte("\x1b[?1002l"))
	if p.MouseMode != MouseModeNone {
		t.Fatalf("mode souris non désactivé: %d", p.MouseMode)
	}

	// 3. Test d'encodage d'évènements souris
	// Clic bouton gauche (btn=0) à x=10, y=5 (1-indexed -> 11, 6)
	evPress := EncodeMouseEvent(0, 10, 5, false, true)
	if string(evPress) != "\x1b[<0;11;6M" {
		t.Fatalf("EncodeMouseEvent press mismatch: %q", string(evPress))
	}

	// Relâchement bouton gauche
	evRelease := EncodeMouseEvent(0, 10, 5, true, true)
	if string(evRelease) != "\x1b[<0;11;6m" {
		t.Fatalf("EncodeMouseEvent release mismatch: %q", string(evRelease))
	}

	// Format X10 standard
	evX10 := EncodeMouseEvent(0, 10, 5, false, false)
	wantX10 := []byte{0x1b, '[', 'M', byte(32 + 0), byte(32 + 11), byte(32 + 6)}
	if !bytes.Equal(evX10, wantX10) {
		t.Fatalf("EncodeMouseEvent X10 mismatch: got %v, want %v", evX10, wantX10)
	}
}

func TestBracketedPaste(t *testing.T) {
	g := &CursorGrid{}
	g.Reset(80, 24)
	p := &Parser{}
	p.Reset(g)

	var pasteMode bool
	p.OnBracketedPaste = func(active bool) {
		pasteMode = active
	}

	p.Feed([]byte("\x1b[?2004h"))
	if !p.BracketedPaste || !pasteMode {
		t.Fatalf("Bracketed paste non activé")
	}

	p.Feed([]byte("\x1b[?2004l"))
	if p.BracketedPaste || pasteMode {
		t.Fatalf("Bracketed paste non désactivé")
	}

	wrapped := EncodeBracketedPaste([]byte("code_snippet"))
	if string(wrapped) != "\x1b[200~code_snippet\x1b[201~" {
		t.Fatalf("EncodeBracketedPaste mismatch: %q", string(wrapped))
	}
}

func TestDSR_CPR(t *testing.T) {
	g := &CursorGrid{}
	g.Reset(80, 24)
	p := &Parser{}
	p.Reset(g)

	var lastResp []byte
	var dsrQuery, dsrCol, dsrRow int
	p.OnResponse = func(resp []byte) {
		lastResp = resp
	}
	p.OnDSR = func(query int, col, row int) {
		dsrQuery = query
		dsrCol = col
		dsrRow = row
	}

	// 1. Positionnement curseur à ligne 10, colonne 15 (1-based -> y=9, x=14)
	p.Feed([]byte("\x1b[10;15H"))

	// 2. Requête DSR 6 (Cursor Position Report)
	p.Feed([]byte("\x1b[6n"))
	if dsrQuery != 6 || dsrRow != 10 || dsrCol != 15 {
		t.Fatalf("OnDSR mismatch: query=%d row=%d col=%d (want 6, 10, 15)", dsrQuery, dsrRow, dsrCol)
	}
	if string(lastResp) != "\x1b[10;15R" {
		t.Fatalf("DSR 6 response mismatch: got %q, want %q", string(lastResp), "\x1b[10;15R")
	}

	// 3. Requête DSR 5 (Status Report)
	lastResp = nil
	p.Feed([]byte("\x1b[5n"))
	if string(lastResp) != "\x1b[0n" {
		t.Fatalf("DSR 5 response mismatch: got %q, want %q", string(lastResp), "\x1b[0n")
	}
}

func TestDeviceAttributes_DA1_DA2(t *testing.T) {
	g := &CursorGrid{}
	g.Reset(80, 24)
	p := &Parser{}
	p.Reset(g)

	var lastResp []byte
	p.OnResponse = func(resp []byte) {
		lastResp = resp
	}

	// Primary DA (\x1b[c)
	p.Feed([]byte("\x1b[c"))
	if string(lastResp) != "\x1b[?62;1;2;6;7;8;9c" {
		t.Fatalf("Primary DA response mismatch: got %q", string(lastResp))
	}

	// Secondary DA (\x1b[>c)
	lastResp = nil
	p.Feed([]byte("\x1b[>c"))
	if string(lastResp) != "\x1b[>0;10;0c" {
		t.Fatalf("Secondary DA response mismatch: got %q", string(lastResp))
	}
}

func TestOSC10_11_DynamicColors(t *testing.T) {
	g := &CursorGrid{}
	g.Reset(80, 24)
	p := &Parser{}
	p.Reset(g)

	var lastResp []byte
	p.OnResponse = func(resp []byte) {
		lastResp = resp
	}

	// OSC 10 query (Foreground)
	p.Feed([]byte("\x1b]10;?\x07"))
	if string(lastResp) != "\x1b]10;rgb:f1f1/f5f5/f9f9\x1b\\" {
		t.Fatalf("OSC 10 query response mismatch: got %q", string(lastResp))
	}

	// OSC 11 query (Background)
	lastResp = nil
	p.Feed([]byte("\x1b]11;?\x1b\\"))
	if string(lastResp) != "\x1b]11;rgb:0202/0606/1717\x1b\\" {
		t.Fatalf("OSC 11 query response mismatch: got %q", string(lastResp))
	}
}

func TestUnicodeEastAsianWidth(t *testing.T) {
	g := &CursorGrid{}
	g.Reset(80, 24)
	p := &Parser{}
	p.Reset(g)

	// Écriture d'un caractère large (⚡ U+26A1 ou 🚀 U+1F680)
	p.Feed([]byte("⚡A"))
	if g.Cells[0].Rune != '⚡' || g.Cells[0].Width != 2 {
		t.Fatalf("Cell 0 wide rune mismatch: rune=%c width=%d", g.Cells[0].Rune, g.Cells[0].Width)
	}
	if g.Cells[1].Width != 0 {
		t.Fatalf("Cell 1 continuation mismatch: width=%d (want 0)", g.Cells[1].Width)
	}
	if g.Cells[2].Rune != 'A' || g.Cells[2].Width != 1 {
		t.Fatalf("Cell 2 next rune mismatch: rune=%c width=%d", g.Cells[2].Rune, g.Cells[2].Width)
	}
	if g.CursorX != 3 {
		t.Fatalf("CursorX mismatch after wide rune: got %d, want 3", g.CursorX)
	}
}

func TestBackgroundColorErase_BCE(t *testing.T) {
	g := &CursorGrid{}
	g.Reset(80, 24)
	p := &Parser{}
	p.Reset(g)

	// Définition d'un fond bleu (SGR 44 -> Bg = 4)
	p.Feed([]byte("\x1b[44m"))
	if p.CurBg != 4 || g.Bg != 4 {
		t.Fatalf("CurBg mismatch: p.CurBg=%d g.Bg=%d", p.CurBg, g.Bg)
	}

	// Effacement de la ligne entière (EL 2)
	p.Feed([]byte("\x1b[2K"))
	for x := 0; x < 80; x++ {
		if g.Cells[x].Bg != 4 {
			t.Fatalf("BCE mismatch at col %d: Bg=%d (want 4)", x, g.Cells[x].Bg)
		}
	}
}
