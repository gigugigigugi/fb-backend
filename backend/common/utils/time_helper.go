package utils

import "time"

// ParseTime 统一处理 ISO8601 格式的时间解析
func ParseTime(layout string) (time.Time, error) {
	return time.Parse(time.RFC3339, layout)
}
