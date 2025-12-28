// reader.go provides file reading utilities including zstd decompression.
//
// For large files, use newStreamingScanner which reads incrementally with
// constant memory. For compressed .zst files, use OpenZstdFile or WrapZstd.

package pgn

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/klauspost/compress/zstd"
)

// openFile opens a PGN file, automatically detecting and handling .zst compression.
// The returned closer must be called when done.
func openFile(path string) (io.Reader, io.Closer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	if strings.HasSuffix(strings.ToLower(path), ".zst") {
		dec, err := zstd.NewReader(f)
		if err != nil {
			f.Close()
			return nil, nil, err
		}
		closer := &multiCloser{dec, f}
		return dec, closer, nil
	}

	return f, f, nil
}

type multiCloser struct {
	dec *zstd.Decoder
	f   *os.File
}

func (m *multiCloser) Close() error {
	m.dec.Close()
	return m.f.Close()
}

// wrapZstd wraps a reader with zstd decompression.
func wrapZstd(r io.Reader) (io.Reader, error) {
	return zstd.NewReader(r)
}

// streamingScanner provides memory-efficient PGN parsing for large files.
// It reads incrementally and doesn't buffer the entire file.
type streamingScanner struct {
	reader   *bufio.Reader
	game     Game
	gs       *GameState
	startPos *GameState
	tagMap   map[string]string
	lineBuf  []byte
	err      error
	done     bool
}

func newStreamingScanner(r io.Reader) *streamingScanner {
	startPos, _ := NewGame(startingFEN)
	return &streamingScanner{
		reader:   bufio.NewReaderSize(r, 64*1024),
		startPos: startPos,
		game:     Game{Moves: make([]Mv, 0, 100)},
		tagMap:   make(map[string]string, 16),
		lineBuf:  make([]byte, 0, 1024),
	}
}

func (s *streamingScanner) Next() bool {
	return !s.done && s.err == nil
}

func (s *streamingScanner) Err() error {
	return s.err
}

func (s *streamingScanner) Scan() (*Game, error) {
	for k := range s.tagMap {
		delete(s.tagMap, k)
	}
	s.game.Tags = s.tagMap
	s.game.Moves = s.game.Moves[:0]

	if err := s.parseTags(); err != nil {
		if err == io.EOF {
			s.done = true
			if len(s.game.Tags) == 0 {
				return nil, io.EOF
			}
		} else {
			s.err = err
			return nil, err
		}
	}

	if fen, ok := s.game.Tags["FEN"]; ok {
		var err error
		s.gs, err = NewGame(fen)
		if err != nil {
			s.err = err
			return nil, err
		}
	} else {
		gs := *s.startPos
		s.gs = &gs
	}

	if err := s.parseMoves(); err != nil && err != io.EOF {
		s.err = err
		return nil, err
	}

	if len(s.game.Tags) == 0 && len(s.game.Moves) == 0 {
		s.done = true
		return nil, io.EOF
	}

	return &s.game, nil
}

func (s *streamingScanner) parseTags() error {
	// Check if we have a buffered line from previous game
	if len(s.lineBuf) > 0 {
		line := trimSpace(s.lineBuf)
		s.lineBuf = s.lineBuf[:0]
		if len(line) > 0 && line[0] == '[' {
			s.parseTagLine(line)
		}
	}

	for {
		line, err := s.readLine()
		if err != nil {
			return err
		}

		line = trimSpace(line)
		if len(line) == 0 {
			continue
		}

		if line[0] == '[' {
			s.parseTagLine(line)
		} else {
			// Any non-tag, non-empty line is move content (includes digits, moves, results like * or 0-1)
			s.lineBuf = append(s.lineBuf[:0], line...)
			return nil
		}
	}
}

func (s *streamingScanner) parseTagLine(line []byte) {
	if len(line) < 4 {
		return
	}
	i := 1
	for i < len(line) && line[i] != ' ' && line[i] != '"' {
		i++
	}
	tagName := internTagName(line[1:i])

	for i < len(line) && line[i] != '"' {
		i++
	}
	if i >= len(line) {
		return
	}
	i++
	start := i
	for i < len(line) && line[i] != '"' {
		i++
	}
	s.game.Tags[tagName] = string(line[start:i])
}

func (s *streamingScanner) parseMoves() error {
	moveText := s.lineBuf

	for {
		line, err := s.readLine()
		if err != nil {
			if err == io.EOF {
				s.processMoveText(moveText)
				return nil
			}
			return err
		}

		line = trimSpace(line)
		if len(line) == 0 {
			continue
		}

		if line[0] == '[' {
			s.processMoveText(moveText)
			s.lineBuf = append(s.lineBuf[:0], line...)
			return nil
		}

		moveText = append(moveText, ' ')
		moveText = append(moveText, line...)
	}
}

func (s *streamingScanner) processMoveText(text []byte) {
	pos := 0
	n := len(text)

	for pos < n {
		for pos < n && isWhitespaceOrDot(text[pos]) {
			pos++
		}
		if pos >= n {
			break
		}

		c := text[pos]

		if c == '{' {
			for pos < n && text[pos] != '}' {
				pos++
			}
			if pos < n {
				pos++
			}
			continue
		}

		if c == '(' {
			depth := 1
			pos++
			for pos < n && depth > 0 {
				if text[pos] == '(' {
					depth++
				} else if text[pos] == ')' {
					depth--
				}
				pos++
			}
			continue
		}

		if c == '1' || c == '0' || c == '*' {
			start := pos
			for pos < n && !isWhitespaceOrDot(text[pos]) {
				pos++
			}
			token := string(text[start:pos])
			if isResult(token) {
				return
			}
			continue
		}

		if c == '$' || c == '!' || c == '?' {
			for pos < n && !isWhitespaceOrDot(text[pos]) {
				pos++
			}
			continue
		}

		if isMoveLetter(c) {
			start := pos
			for pos < n && isMoveChar(text[pos]) {
				pos++
			}
			if start < pos {
				mv, err := ParseSANBytes(s.gs, text[start:pos])
				if err == nil {
					s.game.Moves = append(s.game.Moves, mv)
					MakeMove(s.gs, mv)
				}
			}
			continue
		}

		pos++
	}
}

func (s *streamingScanner) readLine() ([]byte, error) {
	line, err := s.reader.ReadBytes('\n')
	if err != nil && len(line) == 0 {
		return nil, err
	}
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}
	return line, nil
}

func trimSpace(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\t') {
		b = b[1:]
	}
	for len(b) > 0 && (b[len(b)-1] == ' ' || b[len(b)-1] == '\t') {
		b = b[:len(b)-1]
	}
	return b
}

func isWhitespaceOrDot(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '.'
}

func isMoveLetter(c byte) bool {
	return (c >= 'a' && c <= 'h') || (c >= 'A' && c <= 'Z') || c == 'O'
}

func isMoveChar(c byte) bool {
	return (c >= 'a' && c <= 'h') || (c >= '1' && c <= '8') ||
		(c >= 'A' && c <= 'Z') || c == '-' || c == '=' ||
		c == '+' || c == '#' || c == 'x'
}

// Games returns a parser that yields games from a PGN file.
// Handles .zst compressed files automatically.
//
// Workers defaults to NumCPU-1 (leaving one core for decompression).
//
// Example:
//
//	for game := range pgn.Games("huge.pgn.zst").Games {
//	    fmt.Println(game.Tags["Event"])
//	}
func Games(path string, workers ...int) *GameParser {
	w := runtime.NumCPU() - 1
	if len(workers) > 0 && workers[0] > 0 {
		w = workers[0]
	}
	if w < 1 {
		w = 1
	}
	ch := make(chan *Game, w*2)
	gp := &GameParser{
		Games:    ch,
		cancel:   make(chan struct{}),
		closeErr: make(chan struct{}),
	}

	go func() {
		defer close(ch)
		defer close(gp.closeErr)

		r, closer, err := openFile(path)
		if err != nil {
			gp.err = err
			return
		}
		defer closer.Close()

		gp.err = streamToChannel(r, w, ch, gp.cancel)
	}()

	return gp
}

// GamesFromReader returns a channel that yields parsed games from a reader.
// Workers defaults to NumCPU-1 (leaving one core for I/O).
func GamesFromReader(r io.Reader, workers ...int) *GameParser {
	w := runtime.NumCPU() - 1
	if len(workers) > 0 && workers[0] > 0 {
		w = workers[0]
	}
	if w < 1 {
		w = 1
	}
	ch := make(chan *Game, w*2)
	gp := &GameParser{
		Games:    ch,
		cancel:   make(chan struct{}),
		closeErr: make(chan struct{}),
	}

	go func() {
		defer close(ch)
		defer close(gp.closeErr)
		gp.err = streamToChannel(r, w, ch, gp.cancel)
	}()

	return gp
}

// GameParser streams parsed games from a PGN source.
type GameParser struct {
	// Games yields parsed games. Closed when parsing completes.
	Games    <-chan *Game
	cancel   chan struct{}
	closeErr chan struct{}
	err      error
}

// Err returns any error that occurred during parsing.
// Only valid after Games channel is closed.
func (gp *GameParser) Err() error {
	<-gp.closeErr
	return gp.err
}

// Stop cancels parsing early. Safe to call multiple times.
func (gp *GameParser) Stop() {
	select {
	case <-gp.cancel:
	default:
		close(gp.cancel)
	}
}

func streamToChannel(r io.Reader, workers int, out chan<- *Game, cancel <-chan struct{}) error {
	if workers <= 0 {
		workers = 1
	}

	type gameChunk struct {
		index int
		data  []byte
	}

	type parseResult struct {
		index int
		games []*Game
		err   error
	}

	chunks := make(chan gameChunk, workers*2)
	results := make(chan parseResult, workers*2)
	var wg sync.WaitGroup

	// Start parser workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range chunks {
				games, err := parseChunk(chunk.data)
				select {
				case results <- parseResult{chunk.index, games, err}:
				case <-cancel:
					return
				}
			}
		}()
	}

	// Collector goroutine - maintains order and sends to channel
	collectDone := make(chan error, 1)
	go func() {
		pending := make(map[int]parseResult)
		nextIndex := 0
		for result := range results {
			if result.err != nil {
				collectDone <- result.err
				return
			}
			pending[result.index] = result

			for {
				if r, ok := pending[nextIndex]; ok {
					for _, game := range r.games {
						select {
						case out <- game:
						case <-cancel:
							collectDone <- nil
							return
						}
					}
					delete(pending, nextIndex)
					nextIndex++
				} else {
					break
				}
			}
		}
		collectDone <- nil
	}()

	// Read and chunk the input
	const chunkSize = 4 * 1024 * 1024
	br := bufio.NewReaderSize(r, 256*1024)
	var currentChunk []byte
	chunkIndex := 0
	inTags := false

readLoop:
	for {
		select {
		case <-cancel:
			break readLoop
		default:
		}

		line, err := br.ReadBytes('\n')
		if len(line) > 0 {
			trimmed := trimSpace(line)

			if len(trimmed) == 0 {
				inTags = false
			} else if len(trimmed) > 0 && trimmed[0] == '[' {
				if !inTags && len(currentChunk) >= chunkSize {
					select {
					case chunks <- gameChunk{chunkIndex, currentChunk}:
						chunkIndex++
						currentChunk = nil
					case <-cancel:
						break readLoop
					}
				}
				inTags = true
			} else {
				inTags = false
			}

			currentChunk = append(currentChunk, line...)
		}

		if err != nil {
			break
		}
	}

	// Send final chunk
	if len(currentChunk) > 0 {
		select {
		case chunks <- gameChunk{chunkIndex, currentChunk}:
		case <-cancel:
		}
	}

	close(chunks)
	wg.Wait()
	close(results)

	return <-collectDone
}

func parseChunk(data []byte) ([]*Game, error) {
	var games []*Game
	scanner := newStreamingScanner(bytes.NewReader(data))
	for scanner.Next() {
		game, err := scanner.Scan()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Skip games with parse errors, don't fail the whole chunk
			continue
		}
		if len(game.Tags) == 0 && len(game.Moves) == 0 {
			continue
		}
		gameCopy := &Game{
			Tags:  make(map[string]string, len(game.Tags)),
			Moves: make([]Mv, len(game.Moves)),
		}
		for k, v := range game.Tags {
			gameCopy.Tags[k] = v
		}
		copy(gameCopy.Moves, game.Moves)
		games = append(games, gameCopy)
	}
	return games, nil
}
