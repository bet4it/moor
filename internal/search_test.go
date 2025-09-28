package internal

import (
	"testing"

	"github.com/walles/moor/v2/internal/linemetadata"
	"github.com/walles/moor/v2/internal/reader"
	"github.com/walles/moor/v2/twin"
	"gotest.tools/v3/assert"
)

func modeName(pager *Pager) string {
	switch pager.mode.(type) {
	case PagerModeViewing:
		return "Viewing"
	case PagerModeNotFound:
		return "NotFound"
	case *PagerModeSearch:
		return "Search"
	case *PagerModeGotoLine:
		return "GotoLine"
	default:
		panic("Unknown pager mode")
	}
}

// Create a pager with three screen lines reading from a six lines stream
func createThreeLinesPager(t *testing.T) *Pager {
	reader := reader.NewFromTextForTesting("", "a\nb\nc\nd\ne\nf\n")

	screen := twin.NewFakeScreen(20, 3)
	pager := NewPager(reader)

	pager.screen = screen

	assert.Equal(t, "Viewing", modeName(pager), "Initial pager state")

	return pager
}

func TestScrollToNextSearchHit_StartAtBottom(t *testing.T) {
	// Create a pager scrolled to the last line
	pager := createThreeLinesPager(t)
	pager.scrollToEnd()

	// Set the search to something that doesn't exist in this pager
	pager.searchString = "xxx"
	pager.searchPattern = toPattern(pager.searchString)

	// Scroll to the next search hit
	pager.scrollToNextSearchHit()

	assert.Equal(t, "NotFound", modeName(pager))
}

func TestScrollToNextSearchHit_StartAtTop(t *testing.T) {
	// Create a pager scrolled to the first line
	pager := createThreeLinesPager(t)

	// Set the search to something that doesn't exist in this pager
	pager.searchString = "xxx"
	pager.searchPattern = toPattern(pager.searchString)

	// Scroll to the next search hit
	pager.scrollToNextSearchHit()

	assert.Equal(t, "NotFound", modeName(pager))
}

func TestScrollToNextSearchHit_WrapAfterNotFound(t *testing.T) {
	// Create a pager scrolled to the last line
	pager := createThreeLinesPager(t)
	pager.scrollToEnd()

	// Search for "a", it's on the first line (ref createThreeLinesPager())
	pager.searchString = "a"
	pager.searchPattern = toPattern(pager.searchString)

	// Scroll to the next search hit, this should take us into _NotFound
	pager.scrollToNextSearchHit()
	assert.Equal(t, "NotFound", modeName(pager))

	// Scroll to the next search hit, this should wrap the search and take us to
	// the top
	pager.scrollToNextSearchHit()
	assert.Equal(t, "Viewing", modeName(pager))
	assert.Assert(t, pager.lineIndex().IsZero())
}

func TestScrollToNextSearchHit_WrapAfterFound(t *testing.T) {
	// Create a pager scrolled to the last line
	pager := createThreeLinesPager(t)
	pager.scrollToEnd()

	// Search for "f", it's on the last line (ref createThreeLinesPager())
	pager.searchString = "f"
	pager.searchPattern = toPattern(pager.searchString)

	// Scroll to the next search hit, this should take us into _NotFound
	pager.scrollToNextSearchHit()
	assert.Equal(t, "NotFound", modeName(pager))

	// Scroll to the next search hit, this should wrap the search and take us
	// back to the bottom again
	pager.scrollToNextSearchHit()
	assert.Equal(t, "Viewing", modeName(pager))
	assert.Equal(t, 4, pager.lineIndex().Index())
}

// setText sets the text of the inputBox and triggers the onTextChanged callback.
func (b *InputBox) setText(text string) {
	b.text = text
	b.moveCursorEnd()
	if b.onTextChanged != nil {
		b.onTextChanged(b.text)
	}
}

// Ref: https://github.com/walles/moor/issues/152
func Test152(t *testing.T) {
	// Show a pager on a five lines terminal
	reader := reader.NewFromTextForTesting("", "a\nab\nabc\nabcd\nabcde\nabcdef\n")
	screen := twin.NewFakeScreen(20, 5)
	pager := NewPager(reader)
	pager.screen = screen
	assert.Equal(t, "Viewing", modeName(pager), "Initial pager state")

	searchMode := NewPagerModeSearch(pager, SearchDirectionForward, pager.scrollPosition)
	pager.mode = searchMode
	// Search for the first not-visible hit
	searchMode.inputBox.setText("abcde")

	assert.Equal(t, "Search", modeName(pager))
	assert.Equal(t, 2, pager.lineIndex().Index())
}

func assertScreenHasRow(t *testing.T, screen *twin.FakeScreen, expected string) {
	screenshot := ""

	_, screenHeight := screen.Size()
	for rowIndex := 0; rowIndex < screenHeight; rowIndex++ {
		row := screen.GetRow(rowIndex)
		rowString := rowToString(row)
		if rowString == expected {
			return
		}

		if rowIndex > 0 {
			screenshot += "\n"
		}
		screenshot += rowString
	}

	t.Fatalf("Expected screen to contain row '%s', screenshot:\n%s", expected, screenshot)
}

// Start at the top and type a word that has been wrapped off screen. That word
// should become visible.
func TestScrollToNextSearchHit_SubLineHits1(t *testing.T) {
	reader := reader.NewFromTextForTesting("", "1miss 2miss 3miss 4miss 5träff 6miss 7miss 8träff 9miss")

	screen := twin.NewFakeScreen(10, 3)
	pager := NewPager(reader)
	pager.WrapLongLines = true
	pager.ShowStatusBar = false
	pager.ShowLineNumbers = false
	pager.screen = screen

	pager.searchString = "träff"
	searchMode := PagerModeSearch{pager: pager}
	pager.mode = searchMode

	// Scroll to the next search hit
	searchMode.updateSearchPattern()

	// The first hit should be visible
	pager.redraw("")
	assertScreenHasRow(t, screen, "5träff")
}

// Start at the top and go to the next search hit, which has been wrapped off
// screen. The search hit currently visible should be ignored, and the next one
// should become visible.
func TestScrollToNextSearchHit_SubLineHits2(t *testing.T) {
	reader := reader.NewFromTextForTesting("", "1miss 2träff 3miss 4miss 5träff 6miss 7miss 8träff 9miss")

	screen := twin.NewFakeScreen(10, 3)
	pager := NewPager(reader)
	pager.WrapLongLines = true
	pager.ShowStatusBar = false
	pager.ShowLineNumbers = false
	pager.screen = screen

	pager.searchString = "träff"
	searchMode := PagerModeSearch{pager: pager}
	pager.mode = searchMode

	// This will highlight the currently visible hit
	searchMode.updateSearchPattern()

	// This should take us to the next hit
	pager.mode = PagerModeViewing{pager: pager}
	pager.scrollToNextSearchHit()

	// The second hit should be visible
	pager.redraw("")
	assertScreenHasRow(t, screen, "5träff")

	// Go for the last hit as well
	pager.scrollToNextSearchHit()
	pager.redraw("")
	assertScreenHasRow(t, screen, "8träff")
}

// Start at the bottom and type a backwards search for a word that is above us
// in the wrapped line. That word should become visible.
func TestScrollToNextSearchHit_SubLineHits3(t *testing.T) {
	reader := reader.NewFromTextForTesting("", "1miss 2träff 3miss 4miss 5träff 6miss 7miss 8miss 9miss")
	bottom := NewScrollPositionFromIndex(linemetadata.IndexFromZeroBased(99), "end")

	screen := twin.NewFakeScreen(10, 3)
	pager := NewPager(reader)
	pager.WrapLongLines = true
	pager.ShowStatusBar = false
	pager.ShowLineNumbers = false
	pager.screen = screen

	// Precondition: Check that our bottom position actually is
	pager.scrollPosition = bottom
	pager.redraw("")
	assertScreenHasRow(t, screen, "9miss")

	// Start typing a backwards search
	pager.searchString = "träff"
	searchMode := PagerModeSearch{
		pager:                 pager,
		direction:             SearchDirectionBackward,
		initialScrollPosition: bottom,
	}
	pager.mode = searchMode

	// Scroll to the previous search hit
	searchMode.updateSearchPattern()

	// The first hit should be visible
	pager.redraw("")
	assertScreenHasRow(t, screen, "5träff")
}

// Start at the bottom and search for a word that is above us in the wrapped
// line. The search hit currently visible should be ignored, and the next one
// should become visible.
func TestScrollToNextSearchHit_SubLineHits4(t *testing.T) {
	reader := reader.NewFromTextForTesting("", "1miss 2träff 3miss 4miss 5träff 6miss 7miss 8träff 9miss")
	bottom := NewScrollPositionFromIndex(linemetadata.IndexFromZeroBased(99), "end")

	screen := twin.NewFakeScreen(10, 3)
	pager := NewPager(reader)
	pager.WrapLongLines = true
	pager.ShowStatusBar = false
	pager.ShowLineNumbers = false
	pager.screen = screen

	// Precondition: Check that our bottom position actually is
	pager.scrollPosition = bottom
	pager.redraw("")
	assertScreenHasRow(t, screen, "9miss")

	// Start typing a backwards search
	pager.searchString = "träff"
	searchMode := PagerModeSearch{
		pager:                 pager,
		direction:             SearchDirectionBackward,
		initialScrollPosition: bottom,
	}
	pager.mode = searchMode

	// This will highlight the currently visible hit
	searchMode.updateSearchPattern()
	assertScreenHasRow(t, screen, "8träff")

	// This should take us to the previous hit
	pager.mode = PagerModeViewing{pager: pager}
	pager.scrollToPreviousSearchHit()

	// The second (backwards) hit should be visible
	pager.redraw("")
	assertScreenHasRow(t, screen, "5träff")

	// Go for the last (backwards) hit as well
	pager.scrollToPreviousSearchHit()
	pager.redraw("")
	assertScreenHasRow(t, screen, "2träff")
}
