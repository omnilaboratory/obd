package bean

import (
	"fmt"
	"time"
)

//
type JsonDate time.Time

// set time formate
const (
	timeFormat = "2006-01-02"
)

// JsonDate deserialize
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
