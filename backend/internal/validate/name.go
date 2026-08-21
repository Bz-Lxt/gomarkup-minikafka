package validate

import (
	"regexp"
	"unicode/utf8"

	"minikafka/internal/apperror"
)

var resourceRe = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,64}$`)

const (
	MinPartitions = 1
	MaxPartitions = 16
	MaxBatch      = 10000
	MaxConsume    = 500
	MaxBrowse     = 200
)

func ResourceName(kind, name string) error {
	if name == "" {
		return apperror.Wrap(apperror.ErrInvalid, "%s 必填", kind)
	}
	if !utf8.ValidString(name) {
		return apperror.Wrap(apperror.ErrInvalid, "%s 须为合法 UTF-8", kind)
	}
	if !resourceRe.MatchString(name) {
		return apperror.Wrap(apperror.ErrInvalid, "%s 仅允许字母数字 . _ - ，最长 64", kind)
	}
	return nil
}

func PartitionCount(n int) (int, error) {
	if n <= 0 {
		return 1, nil
	}
	if n < MinPartitions || n > MaxPartitions {
		return 0, apperror.Wrap(apperror.ErrInvalid, "partitions 须为 %d–%d", MinPartitions, MaxPartitions)
	}
	return n, nil
}

func PartitionID(id, n int) error {
	if id < 0 || id >= n {
		return apperror.Wrap(apperror.ErrInvalid, "partition")
	}
	return nil
}

func ConsumeLimit(n int) int {
	if n <= 0 || n > MaxConsume {
		return 50
	}
	return n
}

func BrowseLimit(n int) int {
	if n <= 0 || n > MaxBrowse {
		return 20
	}
	return n
}

func ResetTarget(to string) error {
	if to != "earliest" && to != "latest" {
		return apperror.Wrap(apperror.ErrInvalid, "to must be earliest|latest")
	}
	return nil
}

func BatchSize(n int) error {
	if n <= 0 {
		return apperror.Wrap(apperror.ErrInvalid, "messages 不能为空")
	}
	if n > MaxBatch {
		return apperror.Wrap(apperror.ErrInvalid, "单批最多 %d 条", MaxBatch)
	}
	return nil
}

func FormatField(field, msg string) map[string]string {
	return map[string]string{"field": field, "message": msg}
}
