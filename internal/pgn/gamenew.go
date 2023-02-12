package pgn

import (
	"bufio"
	"os"
	"strings"
	"unicode"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type GameRaw struct {
	Tags    []string
	BodyRaw string
}

func WalkPgnFile(
	filepath string,
	onGame func(GameRaw) error,
) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	var tags []string
	var body = &strings.Builder{}
	var hasBody bool

	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var line = scanner.Text()
		if strings.HasPrefix(line, "[") {
			if hasBody {
				if len(tags) != 0 && body.Len() != 0 {
					var err = onGame(GameRaw{
						Tags:    tags,
						BodyRaw: body.String(),
					})
					if err != nil {
						return err
					}
				}
				hasBody = false
				tags = nil
				body.Reset()
			}
			tags = append(tags, line)
		} else {
			hasBody = true
			body.WriteString(line)
			body.WriteString(" ")
		}
	}
	if hasBody && len(tags) != 0 && body.Len() != 0 {
		var err = onGame(GameRaw{
			Tags:    tags,
			BodyRaw: body.String(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func ParsePgnBody(bodyRaw string) []Token {
	var result []Token
	var inComment = false
	var body string
	for _, rune := range bodyRaw {
		if inComment {
			if rune == '}' {
				if len(result) != 0 {
					result[len(result)-1].Comment = body
				}
				inComment = false
				body = ""
			} else {
				body = body + string(rune)
			}
		} else if rune == '.' {
			body = ""
		} else if unicode.IsSpace(rune) {
			if body != "" {
				result = append(result, Token{Value: body})
				body = ""
			}
		} else if rune == '{' {
			if body != "" {
				result = append(result, Token{Value: body})
				body = ""
			}
			inComment = true
			body = ""
		} else {
			body = body + string(rune)
		}
	}
	if body != "" {
		result = append(result, Token{Value: body})
	}
	return result
}

func ParseMoves(startFen string, tokens []Token) ([]common.Position, error) {
	if startFen == "" {
		startFen = common.InitialPositionFen
	}

	var pos, err = common.NewPositionFromFEN(startFen)
	if err != nil {
		return nil, err
	}

	var result []common.Position
	result = append(result, pos)

	for i := range tokens {
		var san = tokens[i].Value

		var curPosition = &result[len(result)-1]
		var move = common.ParseMoveSAN(curPosition, san)
		if move == common.MoveEmpty {
			break
		}
		var child common.Position
		if !curPosition.MakeMove(move, &child) {
			break
		}
		result = append(result, child)
	}

	return result, nil
}
