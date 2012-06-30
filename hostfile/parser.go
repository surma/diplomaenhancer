package hostfile

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

var (
	parserError error
)

func New() Hostfile {
	return []Block{}
}

type parserState func(p *parser, line string) parserState

type parser struct {
	blocks          []Block
	line            string
	hasBufferedLine bool
	reader          *bufio.Reader
}

func (p *parser) NewBlock() {
	if p.blocks == nil {
		p.blocks = []Block{}
	}
	p.blocks = append(p.blocks, Block{})
}

func (p *parser) CurrentBlock() *Block {
	if p.blocks == nil || len(p.blocks) == 0 {
		panic("Blocklist is empty")
	}
	return &p.blocks[len(p.blocks)-1]
}

func (p *parser) NextLine() (string, error) {
	if p.hasBufferedLine {
		p.hasBufferedLine = false
		return p.line, nil
	}
	return p.readLine()
}

func (p *parser) readLine() (string, error) {
	bline, prefix, e := p.reader.ReadLine()
	for prefix && e == nil {
		var blinerest []byte
		blinerest, prefix, e = p.reader.ReadLine()
		bline = append(bline, blinerest...)
	}
	return string(bline), e
}

func (p *parser) Undo(line string) {
	if p.hasBufferedLine {
		panic("multiple undos")
	}
	p.hasBufferedLine = true
	p.line = line
}

func ParseString(content string) (Hostfile, error) {
	return Parse(strings.NewReader(content))
}

func Parse(r io.Reader) (Hostfile, error) {
	p := &parser{
		reader: bufio.NewReader(r),
	}
	state := emptyLineState
	for {
		line, e := p.NextLine()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		line = strings.TrimSpace(line)
		state = state(p, line)
		if state == nil {
			return nil, parserError
		}
	}
	return Hostfile(p.blocks), nil
}

// States
func emptyLineState(p *parser, line string) parserState {
	switch {
	case line == "":
		return emptyLineState
	case strings.HasPrefix(line, "#"):
		p.Undo(line)
		p.NewBlock()
		return commentLineState
	default:
		p.Undo(line)
		p.NewBlock()
		return hostLineState
	}
	return nil
}

func commentLineState(p *parser, line string) parserState {
	switch {
	case line == "":
		return emptyLineState
	case strings.HasPrefix(line, "#"):
		block := p.CurrentBlock()
		block.Comment = append(block.Comment, line[1:])
		return commentLineState
	default:
		p.Undo(line)
		return hostLineState
	}
	return nil
}

func hostLineState(p *parser, line string) parserState {
	switch {
	case line == "":
		return emptyLineState
	case strings.HasPrefix(line, "#"):
		p.Undo(line)
		p.NewBlock()
		return commentLineState
	default:
		block := p.CurrentBlock()
		fields := strings.Fields(line)
		if len(fields) <= 1 {
			parserError = fmt.Errorf("Invalid host line: %s", line)
			return nil
		}
		entry := Entry{}
		entry.IP = fields[0]
		entry.Hostnames = fields[1:]
		block.Entries = append(block.Entries, entry)
		return hostLineState
	}
	return nil
}
