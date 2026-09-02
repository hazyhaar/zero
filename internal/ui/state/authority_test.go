package state

import (
	"sync"
	"testing"
)

func TestMonotonicAuthority_SequenceAndSupersession(t *testing.T) {
	auth := NewMonotonicAuthority()
	seq1 := auth.NextSeq()
	if auth.IsSuperseded(seq1) {
		t.Errorf("seq1 %d should not be superseded yet", seq1)
	}
	seq2 := auth.NextSeq()
	if !auth.IsSuperseded(seq1) {
		t.Errorf("seq1 %d must be superseded after seq2 %d", seq1, seq2)
	}
	if auth.IsSuperseded(seq2) {
		t.Errorf("seq2 %d should be the live sequence", seq2)
	}
}

func TestMonotonicAuthority_PathRevisions(t *testing.T) {
	auth := NewMonotonicAuthority()
	path := "/path/to/file.go"

	if rev := auth.RequiredRevision(path); rev != 0 {
		t.Fatalf("initial revision must be 0, got %d", rev)
	}

	rev1 := auth.InvalidatePath(path)
	if rev1 != 1 {
		t.Fatalf("expected rev 1, got %d", rev1)
	}

	rev2 := auth.InvalidatePath(path)
	if rev2 != 2 {
		t.Fatalf("expected rev 2, got %d", rev2)
	}

	if auth.RequiredRevision(path) != 2 {
		t.Fatalf("expected required revision 2, got %d", auth.RequiredRevision(path))
	}
}

func TestMonotonicAuthority_ConcurrentAccess(t *testing.T) {
	auth := NewMonotonicAuthority()
	var wg sync.WaitGroup
	const goroutines = 50
	const ops = 100

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			path := "file.go"
			for j := 0; j < ops; j++ {
				auth.NextSeq()
				auth.InvalidatePath(path)
				_ = auth.RequiredRevision(path)
				_ = auth.Generation()
			}
		}(i)
	}
	wg.Wait()
}
