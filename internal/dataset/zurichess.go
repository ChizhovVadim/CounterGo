package dataset

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ChizhovVadim/CounterGo/internal/domain"
)

// Separate validation dataset gives us general "unit of measurement" for cost error
type ZurichessDatasetProvider struct {
	FilePath string
}

func (dp *ZurichessDatasetProvider) Load(
	ctx context.Context,
	dataset chan<- domain.DatasetItem,
) error {
	file, err := os.Open(dp.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var s = scanner.Text()

		var index = strings.Index(s, "\"")
		if index < 0 {
			return fmt.Errorf("zurichessParser failed %v", s)
		}

		var fen = s[:index]
		var strScore = s[index+1:]

		var prob float64
		if strings.HasPrefix(strScore, "1/2-1/2") {
			prob = 0.5
		} else if strings.HasPrefix(strScore, "1-0") {
			prob = 1.0
		} else if strings.HasPrefix(strScore, "0-1") {
			prob = 0.0
		} else {
			return fmt.Errorf("zurichessParser failed %v", s)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case dataset <- domain.DatasetItem{
			Fen:    fen,
			Target: prob,
		}:
		}
	}

	return nil
}
