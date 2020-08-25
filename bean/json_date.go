package bean

import (
	"fmt"
	"time"
)

// 自定义类型
type JsonDate time.Time

// 设置时间格式
const (
	timeFormat = "2019-07-18"
)

// JsonDate反序列化
func (t *JsonDate) UnmarshalJSON(data []byte) (err error) {
	if len(data) != 12 {
		return
	}
	newTime, err := time.ParseInLocation("\""+timeFormat+"\"", string(data), time.Local)
	*t = JsonDate(newTime)
	return
}

// JsonDate序列化
func (t JsonDate) MarshalJSON() ([]byte, error) {
	timeStr := fmt.Sprintf("\"%s\"", time.Time(t).Format(timeFormat))
	return []byte(timeStr), nil
}

// string方法
func (t JsonDate) String() string {
	return time.Time(t).Format(timeFormat)
}
