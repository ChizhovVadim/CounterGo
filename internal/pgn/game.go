package pgn

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

const (
	GameResultNone     = "*"
	GameResultWhiteWin = "1-0"
	GameResultBlackWin = "0-1"
	GameResultDraw     = "1/2-1/2"
)

type Game struct {
	Tags  []Tag
	Items []Item
}

type Tag struct {
	Key   string
	Value string
}

type Item struct {
	SanMove    string //for debug
	TxtComment string //for debug
	Position   common.Position
	Comment    Comment
}

func (item Item) String() string {
	return fmt.Sprintln(item.SanMove, item.TxtComment, item.Comment)
}

type Comment struct {
	Depth int
	Score common.UciScore
}

func (g *Game) TagValue(key string) (string, bool) {
	return tagValue(g.Tags, key)
}

func LoadPgnsManyFiles(ctx context.Context, files []string, pgns chan<- string) error {
	for _, filepath := range files {
		var err = LoadPgns(ctx, filepath, pgns)
		if err != nil {
			return err
		}
	}
	return nil
}

func LoadPgns(ctx context.Context, filepath string, pgns chan<- string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	var sb = &strings.Builder{}
	var isEmptyPrevLine bool

	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var line = scanner.Text()
		if strings.HasPrefix(line, "[") && isEmptyPrevLine && sb.Len() != 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case pgns <- sb.String():
				sb = &strings.Builder{}
			}
		}
		sb.WriteString(line)
		sb.WriteString("\n")
		isEmptyPrevLine = strings.TrimSpace(line) == ""
	}

	if sb.Len() != 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case pgns <- sb.String():
		}
	}

	return scanner.Err()
}

func ParseGame(pgn string) (Game, error) {
	var tags = parseTags(pgn)

	var curPosition = startPosition
	if fen, fenFound := tagValue(tags, "FEN"); fenFound {
		var err error
		curPosition, err = common.NewPositionFromFEN(fen)
		if err != nil {
			return Game{}, fmt.Errorf("parse FEN tag failed")
		}
	}

	var tokens = parsePgnTokens(pgn)
	var items = make([]Item, 0, len(tokens))
	var addLastPos = false

	for i := range tokens {
		addLastPos = false

		var san = tokens[i].Value
		var move = common.ParseMoveSAN(&curPosition, san)
		if move == common.MoveEmpty {
			break
		}
		var child common.Position
		if !curPosition.MakeMove(move, &child) {
			break
		}

		var comment = Comment{}
		var txtComment = tokens[i].Comment
		if txtComment != "" {
			//var err error
			comment, _ = parseComment(txtComment)
			//if err != nil {
			//	log.Printf("'%v' %v", txtComment, err)
			//}
		}

		items = append(items, Item{
			SanMove:    san,
			TxtComment: txtComment,
			Position:   curPosition,
			Comment:    comment,
		})

		curPosition = child
		addLastPos = true
	}

	if addLastPos {
		items = append(items, Item{Position: curPosition})
	}

	return Game{
		Tags:  tags,
		Items: items,
	}, nil
}

func parseTags(pgn string) []Tag {
	var tags = make([]Tag, 0, 16)
	tagMatches := tagPairRegex.FindAllStringSubmatch(pgn, -1)
	for i := range tagMatches {
		tags = append(tags, Tag{Key: tagMatches[i][1], Value: tagMatches[i][2]})
	}
	return tags
}

func tagValue(tags []Tag, key string) (string, bool) {
	for _, tag := range tags {
		if tag.Key == key {
			return tag.Value, true
		}
	}
	return "", false
}

type Token struct {
	Value   string
	Comment string
}

func parsePgnTokens(pgn string) []Token {
	pgn = tagsRegex.ReplaceAllString(pgn, "")
	pgn = strings.ReplaceAll(pgn, "\n", " ")
	var result []Token
	var inComment = false
	var body string
	for _, rune := range pgn {
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

//TODO парсить без ошибок '0s'
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
var startPosition, _ = common.NewPositionFromFEN(common.InitialPositionFen)

//TODO В идеале парсить и такие теги
//[Variation "Abbazia defence (classical defence, modern defence[!])"]
var (
	tagsRegex    = regexp.MustCompile(`\[[^\]]+\]`)
	tagPairRegex = regexp.MustCompile(`\[(.*)\s\"(.*)\"\]`)
)
