package chess

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// PGN is a movetext record with optional standard tag pairs.
type PGN struct {
	Tags   map[string]string
	Moves  []Move
	Result string
}

var tagPattern = regexp.MustCompile(`^\[([A-Za-z0-9_]+)\s+"((?:\\.|[^"\\])*)"\]$`)

// ParsePGN reads a PGN record and resolves each SAN token from its starting
// position (or the FEN tag when SetUp is "1").
func ParsePGN(text string) (PGN, error) {
	game := PGN{Tags: make(map[string]string), Result: "*"}
	lines := strings.Split(text, "\n")
	var movetext strings.Builder
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			match := tagPattern.FindStringSubmatch(trimmed)
			if match == nil {
				return PGN{}, fmt.Errorf("invalid PGN tag line %q", trimmed)
			}
			value, err := strconv.Unquote(`"` + match[2] + `"`)
			if err != nil {
				return PGN{}, fmt.Errorf("invalid PGN tag value: %w", err)
			}
			game.Tags[match[1]] = value
			continue
		}
		movetext.WriteString(line)
		movetext.WriteByte(' ')
	}
	position := StartingPosition()
	if game.Tags["SetUp"] == "1" {
		var err error
		position, err = ParseFEN(game.Tags["FEN"])
		if err != nil {
			return PGN{}, fmt.Errorf("parse PGN starting FEN: %w", err)
		}
	}
	tokens, err := pgnTokens(movetext.String())
	if err != nil {
		return PGN{}, err
	}
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" || isMoveNumber(token) {
			continue
		}
		if token == "1-0" || token == "0-1" || token == "1/2-1/2" || token == "*" {
			game.Result = token
			break
		}
		token = stripMoveNumber(token)
		if token == "" {
			continue
		}
		move, err := position.ParseSAN(token)
		if err != nil {
			return PGN{}, fmt.Errorf("parse PGN move %q: %w", token, err)
		}
		game.Moves = append(game.Moves, move)
		position, err = position.MakeMove(move)
		if err != nil {
			return PGN{}, err
		}
	}
	if taggedResult := game.Tags["Result"]; taggedResult != "" && game.Result == "*" {
		game.Result = taggedResult
	}
	return game, nil
}

// FormatPGN emits a compact PGN record. Moves are played from the standard
// initial position unless the PGN carries SetUp "1" and a valid FEN tag.
func FormatPGN(game PGN) (string, error) {
	position := StartingPosition()
	if game.Tags["SetUp"] == "1" {
		var err error
		position, err = ParseFEN(game.Tags["FEN"])
		if err != nil {
			return "", fmt.Errorf("parse PGN starting FEN: %w", err)
		}
	}
	var output strings.Builder
	for key, value := range game.Tags {
		quoted := strconv.Quote(value)
		fmt.Fprintf(&output, "[%s %s]\n", key, quoted)
	}
	if len(game.Tags) > 0 {
		output.WriteByte('\n')
	}
	for index, move := range game.Moves {
		if position.SideToMove == White {
			fmt.Fprintf(&output, "%d. ", position.FullmoveNumber)
		} else if index == 0 {
			fmt.Fprintf(&output, "%d... ", position.FullmoveNumber)
		}
		san, err := position.SAN(move)
		if err != nil {
			return "", fmt.Errorf("format PGN move %d: %w", index+1, err)
		}
		output.WriteString(san)
		output.WriteByte(' ')
		position, err = position.MakeMove(move)
		if err != nil {
			return "", err
		}
	}
	result := game.Result
	if result == "" {
		result = "*"
	}
	output.WriteString(result)
	return strings.TrimSpace(output.String()), nil
}

func pgnTokens(text string) ([]string, error) {
	var tokens []string
	var token strings.Builder
	commentDepth, variationDepth := 0, 0
	flush := func() {
		if token.Len() > 0 {
			tokens = append(tokens, token.String())
			token.Reset()
		}
	}
	for i := 0; i < len(text); i++ {
		char := text[i]
		if commentDepth > 0 {
			if char == '{' {
				commentDepth++
			} else if char == '}' {
				commentDepth--
			}
			continue
		}
		if variationDepth > 0 {
			switch char {
			case '(':
				variationDepth++
			case ')':
				variationDepth--
			}
			continue
		}
		switch char {
		case '{':
			flush()
			commentDepth = 1
		case ';':
			flush()
			for i < len(text) && text[i] != '\n' {
				i++
			}
		case '(':
			flush()
			variationDepth = 1
		case ')':
			return nil, fmt.Errorf("unmatched PGN variation close")
		case ' ', '\t', '\r', '\n':
			flush()
		default:
			token.WriteByte(char)
		}
	}
	if commentDepth != 0 || variationDepth != 0 {
		return nil, fmt.Errorf("unterminated PGN comment or variation")
	}
	flush()
	return tokens, nil
}

func isMoveNumber(token string) bool {
	if len(token) == 0 {
		return false
	}
	dots := strings.TrimLeft(token, "0123456789")
	return dots != "" && strings.Trim(dots, ".") == ""
}

func stripMoveNumber(token string) string {
	for i := 0; i < len(token); i++ {
		if token[i] < '0' || token[i] > '9' {
			return strings.TrimLeft(token[i:], ".")
		}
	}
	return ""
}
