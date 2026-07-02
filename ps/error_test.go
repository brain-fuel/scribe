package ps

import (
	"strings"
	"testing"
)

func wantErr(t *testing.T, src, wantPos, wantSub string) {
	t.Helper()
	_, err := Parse(src)
	if err == nil {
		t.Fatalf("Parse(%q) succeeded, want error", src)
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("error type %T, want *ParseError", err)
	}
	if !strings.HasPrefix(pe.Error(), wantPos+":") {
		t.Errorf("position: got %q, want prefix %q", pe.Error(), wantPos)
	}
	if !strings.Contains(pe.Msg, wantSub) {
		t.Errorf("message %q does not mention %q", pe.Msg, wantSub)
	}
}

func TestErrUnknownWord(t *testing.T) {
	wantErr(t, "1 2 movetoo", "1:5", "unknown word")
}

func TestErrArity(t *testing.T) {
	wantErr(t, "1 moveto", "1:3", "missing number")
}

func TestErrType(t *testing.T) {
	wantErr(t, "1 1 8 8 2 butt roundrect", "1:11", "corner style")
	wantErr(t, "1 setcolor", "1:3", "want color")
}

func TestErrPositionOnLaterLine(t *testing.T) {
	wantErr(t, "% comment\n1 2 moveto\nbogus", "3:1", "unknown word")
}

func TestErrLeftover(t *testing.T) {
	wantErr(t, "1 2 3 moveto", "1:1", "leftover")
}

func TestErrGrestore(t *testing.T) {
	wantErr(t, "gsave grestore grestore", "1:16", "without matching gsave")
}

func TestErrBadColor(t *testing.T) {
	wantErr(t, "#12345 setcolor", "1:1", "#rrggbb")
}

func TestErrBadName(t *testing.T) {
	wantErr(t, "square setjoin", "1:1", "bevel or round")
}
