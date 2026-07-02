package auth

import (
	_ "embed"
	"strings"
)

//go:embed data/10k-most-common.txt
var mostCommonPasswords string
var commonPasswordSet map[string]struct{}

func init() {
	commonPasswordSet = make(map[string]struct{})
	passwords := strings.SplitSeq(mostCommonPasswords, "\n")
	for password := range passwords {
		commonPasswordSet[strings.TrimSpace(strings.ToLower(password))] = struct{}{}
	}
}

func IsCommonPassword(password string) bool {
	_, isACommonPassword := commonPasswordSet[strings.TrimSpace(strings.ToLower(password))]
	return isACommonPassword
}
