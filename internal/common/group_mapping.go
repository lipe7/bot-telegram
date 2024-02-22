package common

import (
	"fmt"
	"os"
)

var GroupSuffixMap = map[int64]string{
	-1002114057976: "1",
	-1002073907096: "2",
	// Adicione mais mapeamentos conforme necessário
}

func GetGroupSuffix(groupID int64) (string, error) {
	suffix, ok := GroupSuffixMap[groupID]
	if !ok {
		return "", fmt.Errorf("Sufixo não mapeado para o GroupID: %d", groupID)
	}
	return suffix, nil
}

func GetEnvVariable(suffix, variableName string) string {
	return os.Getenv(fmt.Sprintf("%s_%s", variableName, suffix))
}
