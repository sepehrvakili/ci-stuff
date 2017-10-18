package texter

import (
	"bytes"
	"log"
	"strings"
)

// MessageChunker interface for all chunking fun.
type MessageChunker interface {
	Split(toSplit string) []string
}

// NewMessageChunker returns a new message chunker.
func NewMessageChunker() MessageChunker {
	return WordBoundaryChunker{}
}

// SimpleChunker will implement the message chunking interface
// in the simplest way possible. We will split based on the
// mod 160 character count backtracking through words.
type SimpleChunker struct{}

// Split will chunk a message up into parts.
func (c SimpleChunker) Split(toSplit string) []string {
	messageCount := len(toSplit) / 160
	if len(toSplit)%160 > 0 {
		messageCount++
	}

	if messageCount > 1 {
		var start, end int
		var chunks []string
		for chunk := 0; chunk < messageCount; chunk++ {
			if chunk == 0 {
				start = 0
			} else {
				start = (chunk * 160)
			}

			if chunk < messageCount-1 {
				end = ((chunk + 1) * 160)
			} else {
				end = len(toSplit)
			}

			chunks = append(chunks, toSplit[start:end])
		}
		return chunks
	}

	return []string{toSplit}
}

type WordBoundaryChunker struct{}

func (c WordBoundaryChunker) Split(toSplit string) []string {
	runes := bytes.Runes([]byte(toSplit))
	messageCount := (len(runes) / 160)
	if len(runes)%160 > 0 {
		messageCount++
	}

	if messageCount > 1 {
		var chunks []string

		start := 0
		end := 160
		done := false
		for !done {
			backtrack := end
			for end-backtrack < 25 && !(end == len(runes)) {
				peek := runes[backtrack]
				if isWhitespace(peek) {
					end = backtrack
					break
				}

				backtrack--
			}

			log.Printf("start %v end %v", start, end)

			message := strings.TrimSpace(string(runes[start:end]))
			log.Printf("message %v", message)

			if len(message) > 0 {
				chunks = append(chunks, message)
			}

			if end == len(runes) {
				done = true
			} else {
				start = end
				end = min(len(runes), start+160)

				log.Printf("New start %v new end %v", start, end)
			}
		}

		return chunks
	}

	return []string{toSplit}
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
