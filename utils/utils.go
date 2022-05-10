package utils

import (
	"strconv"
	"time"
)

func Uint32ToString(val uint32) string {
	return strconv.FormatUint(uint64(val), 10)
}

func StringToUint32(str string) (uint32, error) {
	tmp, err := strconv.ParseUint(str, 10, 32)
	return uint32(tmp), err
}

func Int32ToString(val int) string {
	return strconv.FormatInt(int64(val), 10)
}

func StringToInt(str string) (int, error) {
	tmp, err := strconv.ParseInt(str, 10, 32)
	return int(tmp), err
}

var timeTemplates = []string{
	"2006-01-02 15:04:05", //常规类型
	"2006/01/02 15:04:05",
	"2006-01-02",
	"2006/01/02",
	"15:04:05",
}

func StringToTime(tm string) time.Time {
	for i := range timeTemplates {
		t, err := time.ParseInLocation(timeTemplates[i], tm, time.Local)
		if nil == err && !t.IsZero() {
			return t
		}
	}
	return time.Time{}
}

func CeilDivUint32(a, b uint32) uint32 {
	return (a + b - 1) / b
}

func ArrayContainsString(item string, arr []string) bool {
	for _, v := range arr {
		if v == item {
			return true
		}
	}
	return false
}
