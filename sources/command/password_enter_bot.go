package command

import (
	"bytes"
	"io"
	"regexp"
)

const CHUNK_SIZE = 4096

// PasswordEnterBot Monitors an input stream for a password prompt. When this is
// seen, the password is sent to the output stream, followed by a newline
type PasswordEnterBot struct {
	// ReadFrom The stream to read, watching for the prompt
	ReadFrom io.Reader
	// WriteTo The stream that the password should be sent to (usually STDIN)
	WriteTo io.Writer
	// Prompt The prompt to look for
	Prompt *regexp.Regexp
	// Password The password to enter
	Password string

	doneChan chan bool
	err      error
}

// StartWatching Starts the bot watching the In for the password prompt in
// a separate goroutine. This routine will exit once the password has been
// entered
func (p *PasswordEnterBot) StartWatching() {
	p.doneChan = make(chan bool)

	go func() {
		var lineBuffer []byte
		buffer := make([]byte, 16)

		for {
			// Read bytes from the input
			num, err := p.ReadFrom.Read(buffer)

			if err != nil {
				p.err = err
				break
			}

			// Place these into the line buffer (but only the ones that were
			// actually read, hence the num)
			lineBuffer = append(lineBuffer, buffer[:num]...)

			// TODO: See if I can simplify using MatchReader()
			if p.Prompt.Match(lineBuffer) {
				p.WriteTo.Write([]byte(p.Password + "\n"))
				break
			}

			if idx := bytes.IndexRune(lineBuffer, '\n'); idx >= 0 {
				// If there is a newline, pass on everything before the newline
				// to the output and discard from the buffer
				lineBuffer = lineBuffer[idx+1:]
			}
		}

		close(p.doneChan)
	}()
}

// Done Wait for the password entry to be done and return an error of anything
// went wrong
func (p *PasswordEnterBot) Wait() error {
	<-p.doneChan
	return p.err
}
