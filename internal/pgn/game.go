package pgn

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

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
						Tags:    parseTags(tags),
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
			Tags:    parseTags(tags),
			BodyRaw: body.String(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func parseTags(tags []string) []Tag {
	var result []Tag
	for _, tag := range tags {
		tag = strings.TrimLeft(tag, "[")
		tag = strings.TrimRight(tag, "]")
		var i0 = strings.Index(tag, "\"")
		var i1 = strings.LastIndex(tag, "\"")
		if i0 == -1 || i1 == -1 {
			continue
		}
		var name = strings.TrimSpace(tag[:i0])
		var val = tag[i0+1 : i1]
		result = append(result, Tag{Key: name, Value: val})
	}
	return result
}

func ParseGame(gameRaw GameRaw) (Game, error) {
	var result Game
	var ok bool
	result.Result, ok = TagValue(gameRaw.Tags, "Result")
	if !ok {
		return Game{}, fmt.Errorf("no game result")
	}
	result.Fen, _ = TagValue(gameRaw.Tags, "FEN")

	var startFen = result.Fen
	if startFen == "" {
		startFen = common.InitialPositionFen
	}
	var pos, err = common.NewPositionFromFEN(startFen)
	if err != nil {
		return Game{}, err
	}

	var tokens = parsePgnBody(gameRaw.BodyRaw)

	for i := range tokens {
		var san = tokens[i].Value
		var move = common.ParseMoveSAN(&pos, san)
		if move == common.MoveEmpty {
			break
		}
		var child common.Position
		if !pos.MakeMove(move, &child) {
			break
		}
		pos = child

		var comment, _ = parseComment(tokens[i].Comment)
		result.Items = append(result.Items, Item{
			Move:    move,
			Comment: comment,
		})
	}

	return result, nil
}

func TagValue(tags []Tag, key string) (string, bool) {
	for _, tag := range tags {
		if tag.Key == key {
			return tag.Value, true
		}
	}
	return "", false
}

type token struct {
	Value   string
	Comment string
}

func parsePgnBody(bodyRaw string) []token {
	var result []token
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
				result = append(result, token{Value: body})
				body = ""
			}
		} else if rune == '{' {
			if body != "" {
				result = append(result, token{Value: body})
				body = ""
			}
			inComment = true
			body = ""
		} else {
			body = body + string(rune)
		}
	}
	if body != "" {
		result = append(result, token{Value: body})
	}
	return result
}

// Примеры комментариев:
// {-1.46/10 0.004s}
// {-12.05/10 0.014s, Black wins by adjudication}
// TODO парсить без ошибок '0s'
func parseComment(comment string) (Comment, error) {
	comment = strings.TrimLeft(comment, "{")
	comment = strings.TrimRight(comment, "}")

	if comment == "book" {
		return Comment{}, nil
	}

	var fields = strings.Fields(comment)
	if len(fields) >= 2 {
		var s string
		if strings.HasPrefix(fields[0], "(") {
			s = fields[1]
		} else {
			s = fields[0]
		}
		if s != "" {
			var index = strings.Index(s, "/")
			if index >= 0 {
				var sScore = s[:index]
				var sDepth = s[index+1:]

				var uciScore common.UciScore
				if strings.Contains(sScore, "M") {
					sScore = strings.Replace(sScore, "M", "", 1)
					score, err := strconv.Atoi(sScore)
					if err != nil {
						return Comment{}, err
					}
					uciScore = common.UciScore{Mate: score}
				} else {
					score, err := strconv.ParseFloat(sScore, 64)
					if err != nil {
						return Comment{}, err
					}
					uciScore = common.UciScore{Centipawns: int(100 * score)}
				}

				depth, err := strconv.Atoi(sDepth)
				if err != nil {
					return Comment{}, err
				}
				return Comment{
					Score: uciScore,
					Depth: depth,
				}, nil
			}
		}
	}
	return Comment{}, errParseComment
}

var errParseComment = errors.New("parse comment failed")
