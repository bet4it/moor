package reader

import (
	"regexp"
	"strings"

	"github.com/rivo/uniseg"
	"github.com/walles/moor/v2/internal/linemetadata"
	"github.com/walles/moor/v2/internal/textstyles"
	"github.com/walles/moor/v2/twin"
)

type NumberedLine struct {
	Index  linemetadata.Index
	Number linemetadata.Number
	Line   *Line
}

// Warning: This is slow. If that turns out to be a problem, start profiling.
func (nl *NumberedLine) Plain() string {
	styled := nl.Line.HighlightedTokens(twin.StyleDefault, twin.StyleDefault, nil, &nl.Index).StyledRunes

	var b strings.Builder
	b.Grow(len(styled))
	for i := range styled {
		b.WriteRune(styled[i].Rune)
	}
	return b.String()
}

func (nl *NumberedLine) HighlightedTokens(plainTextStyle twin.Style, searchHitStyle twin.Style, search *regexp.Regexp) textstyles.StyledRunesWithTrailer {
	return nl.Line.HighlightedTokens(plainTextStyle, searchHitStyle, search, &nl.Index)
}

func (nl *NumberedLine) DisplayWidth() int {
	width := 0
	for _, r := range nl.Plain() {
		width += uniseg.StringWidth(string(r))
	}
	return width
}
