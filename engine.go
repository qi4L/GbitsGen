package main

import (
	"bytes"
	"io"
	"math/rand"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type GhostBitsEngine struct{}

func (e *GhostBitsEngine) canEncodeGB2312(r rune) bool {
	encoder := simplifiedchinese.GBK.NewEncoder()
	reader := transform.NewReader(strings.NewReader(string(r)), encoder)
	data, err := io.ReadAll(reader)
	if err != nil {
		return false
	}
	if len(data) == 0 {
		return false
	}
	b := data[0]
	return b >= 0xB0 && b <= 0xD7
}

func (e *GhostBitsEngine) GetGhostChar(char rune, mode int) rune {
	if char > 255 {
		return char
	}

	ll := int(char)
	var validHHs []int
	var fallbackHHs []int

	switch mode {
	case 0:
		for hh := 0x4E; hh <= 0x9F; hh++ {
			codepoint := (hh << 8) | ll
			if codepoint >= 0x4E00 && codepoint <= 0x9FA5 {
				fallbackHHs = append(fallbackHHs, hh)
				if e.canEncodeGB2312(rune(codepoint)) {
					validHHs = append(validHHs, hh)
				}
			}
		}
		if len(validHHs) == 0 {
			validHHs = fallbackHHs
		}

	case 1:
		validHHs = []int{0x01, 0x02, 0x03, 0x04}

	default:
		for hh := 0x01; hh <= 0xFF; hh++ {
			if hh >= 0xD8 && hh <= 0xDF {
				continue
			}
			validHHs = append(validHHs, hh)
		}
	}

	if len(validHHs) == 0 {
		return char
	}

	chosenHH := validHHs[rand.Intn(len(validHHs))]
	return rune((chosenHH << 8) | ll)
}

func (e *GhostBitsEngine) Generate(baseText string, repeats int, tail string, mode int, exemptChars string, asUnicode bool) string {
	baseText = strings.ReplaceAll(baseText, "[CRLF]", "\r\n")
	baseText = strings.ReplaceAll(baseText, "\\r", "\r")
	baseText = strings.ReplaceAll(baseText, "\\n", "\n")

	var fullPayload strings.Builder

	for i := 0; i < repeats; i++ {
		for _, char := range baseText {
			if strings.ContainsRune(exemptChars, char) {
				fullPayload.WriteRune(char)
			} else {
				ghostChar := e.GetGhostChar(char, mode)
				if asUnicode {
					fullPayload.WriteString(formatUnicode(ghostChar))
				} else {
					fullPayload.WriteRune(ghostChar)
				}
			}
		}
	}

	fullPayload.WriteString(tail)
	return fullPayload.String()
}

func formatUnicode(r rune) string {
	var buf bytes.Buffer
	buf.WriteString("\\u")
	for _, d := range []byte{byte((r >> 12) & 0xF), byte((r >> 8) & 0xF), byte((r >> 4) & 0xF), byte(r & 0xF)} {
		if d < 10 {
			buf.WriteByte('0' + d)
		} else {
			buf.WriteByte('a' + d - 10)
		}
	}
	return buf.String()
}
