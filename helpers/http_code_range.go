package helpers

import (
	"strconv"
	"strings"
)

type HTTPCodeRanges [][2]int

func NewHTTPCodeRanges(strBlocks []string) (HTTPCodeRanges, error) {
	var blocks HTTPCodeRanges
	for _, block := range strBlocks {
		codes := strings.Split(block, "-")
		if len(codes) == 1 {
			codes = append(codes, codes[0])
		}
		lowCode, err := strconv.Atoi(codes[0])
		if err != nil {
			return nil, err
		}
		highCode, err := strconv.Atoi(codes[1])
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, [2]int{lowCode, highCode})
	}
	return blocks, nil
}

func (h HTTPCodeRanges) Contains(statusCode int) bool {
	for _, block := range h {
		if statusCode >= block[0] && statusCode <= block[1] {
			return true
		}
	}
	return false
}