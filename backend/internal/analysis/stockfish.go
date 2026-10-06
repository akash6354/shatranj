package analysis

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/shatranj/backend/internal/chess"
)

const (
	defaultEngineTimeout = 15 * time.Second
	defaultEngineDepth   = 14
)

type Stockfish struct {
	binary  string
	timeout time.Duration
	depth   int
}

func NewStockfish(binary string, timeout time.Duration, depth int) *Stockfish {
	if timeout <= 0 {
		timeout = defaultEngineTimeout
	}
	if depth <= 0 {
		depth = defaultEngineDepth
	}
	return &Stockfish{binary: strings.TrimSpace(binary), timeout: timeout, depth: depth}
}

func (s *Stockfish) Available() bool {
	return s.binary != ""
}

func (s *Stockfish) Evaluate(ctx context.Context, fen string) (Evaluation, error) {
	results, err := s.EvaluateMany(ctx, []string{fen})
	if err != nil {
		return Evaluation{}, err
	}
	return results[0], nil
}

// EvaluateMany reuses one engine process for all positions in a game while
// enforcing a separate search timeout for each position.
func (s *Stockfish) EvaluateMany(ctx context.Context, fens []string) ([]Evaluation, error) {
	if !s.Available() {
		return nil, ErrUnavailable
	}
	if len(fens) == 0 {
		return []Evaluation{}, nil
	}
	positions := make([]chess.Position, len(fens))
	for index, fen := range fens {
		position, err := chess.ParseFEN(fen)
		if err != nil {
			return nil, fmt.Errorf("invalid analysis FEN at position %d: %w", index, err)
		}
		positions[index] = position
	}

	command := exec.CommandContext(ctx, s.binary)
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("open Stockfish stdin: %w", err)
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, fmt.Errorf("open Stockfish stdout: %w", err)
	}
	command.Stderr = io.Discard
	if err := command.Start(); err != nil {
		return nil, fmt.Errorf("start Stockfish: %w", err)
	}
	lines := make(chan string, 128)
	done := make(chan struct{})
	go scanEngineOutput(stdout, lines, done)
	defer func() {
		close(done)
		_ = stdin.Close()
		wait := make(chan error, 1)
		go func() {
			wait <- command.Wait()
		}()
		select {
		case <-wait:
		case <-time.After(500 * time.Millisecond):
			_ = command.Process.Kill()
			<-wait
		}
	}()
	write := func(command string) error {
		_, err := io.WriteString(stdin, command+"\n")
		return err
	}
	if err := write("uci"); err != nil {
		return nil, fmt.Errorf("write Stockfish uci command: %w", err)
	}
	if err := awaitEngineLine(ctx, lines, "uciok", s.timeout); err != nil {
		return nil, withEngineError(err)
	}
	if err := write("setoption name Threads value 1"); err != nil {
		return nil, fmt.Errorf("configure Stockfish: %w", err)
	}
	if err := write("isready"); err != nil {
		return nil, fmt.Errorf("check Stockfish readiness: %w", err)
	}
	if err := awaitEngineLine(ctx, lines, "readyok", s.timeout); err != nil {
		return nil, withEngineError(err)
	}

	evaluations := make([]Evaluation, 0, len(fens))
	for index, fen := range fens {
		if err := write("position fen " + fen); err != nil {
			return nil, fmt.Errorf("set Stockfish position %d: %w", index, err)
		}
		if err := write(fmt.Sprintf("go depth %d", s.depth)); err != nil {
			return nil, fmt.Errorf("start Stockfish search %d: %w", index, err)
		}
		evaluation, err := readSearchResult(ctx, lines, positions[index], s.timeout)
		if err != nil {
			return nil, fmt.Errorf("evaluate position %d: %w", index, withEngineError(err))
		}
		evaluations = append(evaluations, evaluation)
	}
	return evaluations, nil
}

func awaitEngineLine(ctx context.Context, lines <-chan string, expected string, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return context.DeadlineExceeded
		case line := <-lines:
			if strings.HasPrefix(line, "\x00EOF") {
				return fmt.Errorf("Stockfish closed its output: %s", strings.TrimPrefix(line, "\x00EOF"))
			}
			if line == expected {
				return nil
			}
		}
	}
}

func readSearchResult(ctx context.Context, lines <-chan string, position chess.Position, timeout time.Duration) (Evaluation, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	var score int
	scoreSet := false
	for {
		select {
		case <-ctx.Done():
			return Evaluation{}, ctx.Err()
		case <-timer.C:
			return Evaluation{}, context.DeadlineExceeded
		case line := <-lines:
			if strings.HasPrefix(line, "\x00EOF") {
				return Evaluation{}, fmt.Errorf("Stockfish closed its output: %s", strings.TrimPrefix(line, "\x00EOF"))
			}
			if strings.HasPrefix(line, "info ") {
				if cp, ok := parseCentipawnScore(line); ok {
					score, scoreSet = cp, true
				}
				if mate, ok := parseMateScore(line); ok {
					score, scoreSet = mate, true
				}
				continue
			}
			if !strings.HasPrefix(line, "bestmove ") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 2 || fields[1] == "(none)" || fields[1] == "0000" {
				return Evaluation{}, errors.New("Stockfish returned no legal best move")
			}
			if !scoreSet {
				return Evaluation{}, errors.New("Stockfish returned no position evaluation")
			}
			if position.SideToMove == chess.Black {
				score = -score
			}
			return Evaluation{ScoreCP: score, BestMove: fields[1]}, nil
		}
	}
}

func scanEngineOutput(stdout io.Reader, lines chan<- string, done <-chan struct{}) {
	sendLine := func(line string) bool {
		select {
		case <-done:
			return false
		case lines <- line:
			return true
		}
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	for scanner.Scan() {
		if !sendLine(strings.TrimSpace(scanner.Text())) {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		sendLine("\x00EOF" + err.Error())
	} else {
		sendLine("\x00EOF" + io.EOF.Error())
	}
}

func parseCentipawnScore(line string) (int, bool) {
	fields := strings.Fields(line)
	for index := 0; index+2 < len(fields); index++ {
		if fields[index] == "score" && fields[index+1] == "cp" {
			value, err := strconv.Atoi(fields[index+2])
			return value, err == nil
		}
	}
	return 0, false
}

func parseMateScore(line string) (int, bool) {
	fields := strings.Fields(line)
	for index := 0; index+2 < len(fields); index++ {
		if fields[index] == "score" && fields[index+1] == "mate" {
			moves, err := strconv.Atoi(fields[index+2])
			if err != nil {
				return 0, false
			}
			if moves == 0 {
				return 100000, true
			}
			score := 100000 - abs(moves)*100
			if moves < 0 {
				score = -score
			}
			return score, true
		}
	}
	return 0, false
}

func withEngineError(err error) error {
	return fmt.Errorf("Stockfish analysis failed: %w", err)
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
