package parser

import (
	"bufio"
	"context"
	"os"
	"regexp"
)

// Parser Struct that is capable of parsing single-line files using a regex
type Parser struct {
	// Lines that match this regex will be ignored. Usually used to ignore
	// comments
	Ignore *regexp.Regexp

	// Lines that match this regex will be captured and the captures returned.
	// The capturing groups must be named
	Capture *regexp.Regexp
}

// ContinueFunc is a function that is executed to determine if the parser should
// continue. The input of a function is the parsed result and if the function
// returns true the parser will continue, otherwise it will stop
type ContinueFunc func(map[string]string) bool

// WithParsedResults Executes a ContinueFunc for each parsed result within a
// given file. It also requires a context to allow the scanning to be cancelled
// if it is taking too long
func (p *Parser) WithParsedResults(ctx context.Context, scanner *bufio.Scanner, f ContinueFunc) error {
	var err error
	var line string
	var captureNames []string

	// Load the capture names in advance
	if p.Capture != nil {
		captureNames = p.Capture.SubexpNames()
	}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			// Allow the scanning to be cancelled
			return ctx.Err()
		default:
			line = scanner.Text()

			// Skip if we are to ignore this line
			if p.Ignore != nil {
				if p.Ignore.MatchString(line) {
					continue
				}
			}

			// Capture lines
			if p.Capture != nil {
				matches := p.Capture.FindStringSubmatch(line)

				if len(matches) > 0 {
					lineMatches := make(map[string]string)

					for i, match := range matches {
						matchName := captureNames[i]
						if matchName != "" {
							lineMatches[matchName] = match
						}
					}

					// Execute the ContinueFunc to determine whether or not we
					// should keep going
					if !f(lineMatches) {
						return nil
					}
				}
			}
		}
	}

	if scanner.Err() != nil {
		return err
	}

	return nil
}

// Parse will parse a file's lines. Lines will be first checked to see if they
// match the Ignore regex, if they do they will be ignored. Else they will be
// checked against the Capture regex. The function returns a slice where each
// element is a map of capture group name to captured value
func (p *Parser) Parse(ctx context.Context, file *os.File) ([]map[string]string, error) {
	parsedResults := make([]map[string]string, 0)

	scanner := bufio.NewScanner(file)

	err := p.WithParsedResults(ctx, scanner, func(m map[string]string) bool {
		parsedResults = append(parsedResults, m)

		return true
	})

	return parsedResults, err
}
