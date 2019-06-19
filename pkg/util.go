package pkg

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-openapi/swag"
)

func ResovleBackupFromPodName(podName string) (string, error) {
	s := strings.Split(podName, "-")
	if len(s) != 6 {
		return "", errors.New("len of pod name should be exactly 6")
	}

	timeStamp, err := swag.ConvertInt64(s[len(s)-2])
	if err != nil {
		return "", err
	}

	tm := time.Unix(timeStamp, 0)

	backupName := fmt.Sprintf("scheduled-backup-%s", tm.Format("20060102-150405"))

	return backupName, nil
}
