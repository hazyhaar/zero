package vtgrid

import (
	"testing"
	"unsafe"
)

func TestCellMemoryLayout(t *testing.T) {
	var c C2_cell_t
	if sz := unsafe.Sizeof(c); sz != 8 {
		t.Fatalf("C2_cell_t doit mesurer exactement 8 octets, got: %d", sz)
	}
}

func TestGridPutCellAndWrap(t *testing.T) {
	g := NewGrid(10, 5)
	// Écrire 10 caractères sur la première ligne
	for i := 0; i < 10; i++ {
		g.PutCell(uint32('0'+i), 1)
	}
	// Vérifier que le curseur est passé à la ligne 1, colonne 0
	if g.cursorX != 0 || g.cursorY != 1 {
		t.Fatalf("Curseur attendu en (0, 1) après wrap, got: (%d, %d)", g.cursorX, g.cursorY)
	}
	// Vérifier la première cellule
	if (*g.active)[0].Rune != '0' || (*g.active)[9].Rune != '9' {
		t.Fatalf("Contenu de ligne 0 invalide")
	}
}

func TestGridScrollUpAndDown(t *testing.T) {
	g := NewGrid(10, 4)
	g.SetSGR(7, 0, 0)
	// Remplir 3 lignes sans déclencher le scroll automatique sur la dernière
	for r := 0; r < 3; r++ {
		for c := 0; c < 10; c++ {
			g.PutCell(uint32('A'+r), 1)
		}
	}

	// ScrollUp 1 ligne : 'B', 'C' remontent aux lignes 0 et 1, la ligne 2 et 3 sont vides
	g.ScrollUp(1)
	if (*g.active)[0].Rune != 'B' || (*g.active)[10].Rune != 'C' {
		t.Fatalf("Lignes 0-1 invalides après ScrollUp(1)")
	}
	if (*g.active)[20].Rune != 0 || (*g.active)[30].Rune != 0 {
		t.Fatalf("Lignes inférieures non effacées après ScrollUp(1)")
	}

	// ScrollDown 1 ligne : réinsertion en haut
	g.ScrollDown(1)
	if (*g.active)[0].Rune != 0 {
		t.Fatalf("Première ligne non effacée après ScrollDown(1)")
	}
	if (*g.active)[10].Rune != 'B' || (*g.active)[20].Rune != 'C' {
		t.Fatalf("Lignes 1-2 invalides après ScrollDown(1)")
	}
}

func TestGridSwapScreen(t *testing.T) {
	g := NewGrid(10, 5)
	g.PutCell('P', 1)
	if (*g.active)[0].Rune != 'P' {
		t.Fatalf("Primary cell invalide")
	}

	// Bascule sur Alternate Screen et positionnement curseur en (0, 0)
	g.SwapScreen()
	g.SetCursor(0, 0)
	if (*g.active)[0].Rune != 0 {
		t.Fatalf("Alternate screen attendu vide au départ")
	}
	g.PutCell('A', 1)
	if (*g.active)[0].Rune != 'A' {
		t.Fatalf("Alternate cell invalide")
	}

	// Retour sur Primary
	g.SwapScreen()
	if (*g.active)[0].Rune != 'P' {
		t.Fatalf("Primary cell non restaurée après SwapScreen")
	}
}

func TestGridReflow(t *testing.T) {
	g := NewGrid(10, 5)
	for i := 0; i < 5; i++ {
		g.PutCell(uint32('A'+i), 1)
	}

	// Redimensionnement à 20x10
	g.Reflow(20, 10)
	if g.cols != 20 || g.rows != 10 {
		t.Fatalf("Dimensions invalides après Reflow: %dx%d", g.cols, g.rows)
	}
	if (*g.active)[0].Rune != 'A' || (*g.active)[4].Rune != 'E' {
		t.Fatalf("Contenu corrompu après expansion")
	}

	// Réduction à 4x2
	g.Reflow(4, 2)
	if g.cols != 4 || g.rows != 2 {
		t.Fatalf("Dimensions invalides après shrinking: %dx%d", g.cols, g.rows)
	}
	if (*g.active)[0].Rune != 'A' || (*g.active)[3].Rune != 'D' {
		t.Fatalf("Contenu corrompu après réduction")
	}
}

func TestScrollbackRing(t *testing.T) {
	ring := NewScrollbackRing(5)
	if ring.Count() != 0 {
		t.Fatalf("Anneau initial non vide: %d", ring.Count())
	}

	// Insérer 5 lignes
	for i := 0; i < 5; i++ {
		line := make([]C2_cell_t, 10)
		line[0].Rune = uint32('0' + i)
		ring.Push(line)
	}
	if ring.Count() != 5 {
		t.Fatalf("Count attendu 5, got %d", ring.Count())
	}
	if ring.Get(0)[0].Rune != '0' || ring.Get(4)[0].Rune != '4' {
		t.Fatalf("Lignes 0 ou 4 invalides")
	}

	// Insérer une 6ème ligne (éviction de '0')
	line6 := make([]C2_cell_t, 10)
	line6[0].Rune = '5'
	ring.Push(line6)
	if ring.Count() != 5 {
		t.Fatalf("Count attendu 5 après saturation, got %d", ring.Count())
	}
	// L'élément le plus ancien est maintenant '1'
	if ring.Get(0)[0].Rune != '1' {
		t.Fatalf("Éviction FIFO incorrecte: got %c, want '1'", ring.Get(0)[0].Rune)
	}
	// L'élément le plus récent est '5'
	if ring.Get(4)[0].Rune != '5' {
		t.Fatalf("Dernier élément attendu '5', got %c", ring.Get(4)[0].Rune)
	}

	// Hors-bornes
	if ring.Get(-1) != nil || ring.Get(10) != nil {
		t.Fatalf("Get hors-bornes doit renvoyer nil")
	}
}

func BenchmarkPutCell(b *testing.B) {
	g := NewGrid(80, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		g.PutCell('A', 1)
	}
}

func BenchmarkScrollbackRingPush(b *testing.B) {
	ring := NewScrollbackRing(10000)
	line := make([]C2_cell_t, 80)
	for i := range line {
		line[i].Rune = 'X'
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ring.Push(line)
	}
}
