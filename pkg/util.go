package pkg

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-openapi/swag"
)

func ResovleBackupFromPathSuffix(pathSuffix string) (string, error) {
	s := strings.Split(pathSuffix, "-")
	if len(s) < 2 {
		return "", errors.New("len of path suffix should be more than 2")
	}

	timestamp, err := swag.ConvertInt64(s[len(s)-2])
	if err != nil {
		return "", err
	}

	tm := time.Unix(timestamp, 0)

	backupName := fmt.Sprintf("scheduled-backup-%s", tm.Format(time.RFC3339))
	return backupName, nil
}
