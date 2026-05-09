package mchart

import (
	"fmt"
	"strings"
	"unicode"
)

func estimateTextWidth(text string, fontSize float64) float64 {
	var width float64
	for _, r := range text {
		switch {
		case unicode.IsSpace(r):
			width += fontSize * 0.33
		case strings.ContainsRune("ilIjtfr.,:;!|'`", r):
			width += fontSize * 0.32
		case strings.ContainsRune("mwMW@#%", r):
			width += fontSize * 0.9
		case unicode.IsUpper(r):
			width += fontSize * 0.65
		default:
			width += fontSize * 0.56
		}
	}
	return width
}

func breakWord(word string, maxWidth, fontSize float64) []string {
	if word == "" {
		return nil
	}

	runes := []rune(word)
	parts := make([]string, 0, 2)
	start := 0

	for start < len(runes) {
		end := start + 1
		for end <= len(runes) {
			piece := string(runes[start:end])
			if estimateTextWidth(piece, fontSize) > maxWidth {
				if end == start+1 {
					end++
				}
				break
			}
			end++
		}

		if end > len(runes)+1 {
			end = len(runes) + 1
		}

		partEnd := end - 1
		if partEnd <= start {
			partEnd = start + 1
		}
		if partEnd > len(runes) {
			partEnd = len(runes)
		}

		parts = append(parts, string(runes[start:partEnd]))
		start = partEnd
	}

	return parts
}

func wrapLabel(label string, maxWidth, fontSize float64) []string {
	trimmed := strings.TrimSpace(label)
	if trimmed == "" {
		return []string{""}
	}

	if maxWidth <= 0 {
		return []string{trimmed}
	}

	words := strings.Fields(trimmed)
	if len(words) == 0 {
		return []string{trimmed}
	}

	lines := make([]string, 0, 2)
	line := ""

	for _, word := range words {
		candidates := []string{word}
		if estimateTextWidth(word, fontSize) > maxWidth {
			candidates = breakWord(word, maxWidth, fontSize)
		}

		for _, part := range candidates {
			if line == "" {
				line = part
				continue
			}

			combined := line + " " + part
			if estimateTextWidth(combined, fontSize) <= maxWidth {
				line = combined
				continue
			}

			lines = append(lines, line)
			line = part
		}
	}

	if line != "" {
		lines = append(lines, line)
	}

	if len(lines) == 0 {
		return []string{trimmed}
	}

	return lines
}

func ptPercent(v float64) string {
	s := fmt.Sprintf("%.2f%%", v)
	return strings.ReplaceAll(s, ".", ",")
}

func ptNumber(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	return strings.ReplaceAll(s, ".", ",")
}

func clamp(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}
