package hostfile

import (
	"bufio"
	"io"
	"strings"
)

func New() Hostfile {
	return []Block{}
}

type parserState func(p parser) parserState

type parser struct {
	currentBlock Block
	line string
	reader bufio.Reader
}

func (p *parser) NextLine() (string, error) {
	if p.line != nil {
		return p.line, nil
	}
	return p.readLine()
}

func (p *parser) readLine() (string, error) {
	bline, prefix, e := b.ReadLine();
	for prefix && e == nil {
		var blinerest []byte
		blinerest, prefix, e = b.ReadLine()
		bline = append(bline, blinerest...)
	}
	return string(bline), e
}

func (p *parser) Undo(line string) {
	if p.line != nil {
		panic("multiple undos")
	}
	p.line = line
}

func ParseString(content string) (Hostfile, error) {
	return Parse(strings.NewReader(content))
}

func Parse(r io.Reader) (Hostfile, error) {
	h := New()
	p := &parser{
		reader: bufio.NewReader(r),
	}

	return h, nil
}

func readEmptyLine(line string) parserState {
	switch {
		case e == io.EOF:
			return endState
		case e != nil:
			errorStateError = e
			return errorState
		}
		case strings.TrimSpace(line) == "":

	}
	p.Undo(line)
	return readComment
}

func readComment(p parser) parserState {
	for line, e := p.NextLine(); e == nil; line, e = p.NextLine() {

	}
	if e != nil {
		errorStateError = e
		return errorState
	}
	p.Undo(line)
}

var errorStateError error
func errorState(p parser) parserState {}

func endState(p parser) parserState {}
