package utils

import "time"

// 커스텀 시간 포맷을 위한 타입입니다.
type JSONTime time.Time

const jsonTimeFormat = "2006-01-02 15:04:05"

// JSONTime을 JSON으로 변환합니다.
func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := time.Time(t).Format(jsonTimeFormat)
	return []byte(`"` + stamp + `"`), nil
}

// JSON을 JSONTime으로 변환합니다.
func (t *JSONTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	s = s[1 : len(s)-1] // 따옴표 제거

	// 새로운 포맷 시도 (초 포함, 로컬 타임존으로 파싱)
	parsed, err := time.ParseInLocation(jsonTimeFormat, s, time.Local)
	if err == nil {
		*t = JSONTime(parsed)
		return nil
	}

	// 기존 포맷 시도 (초 없음, 기존 데이터 호환)
	parsed, err = time.ParseInLocation("2006-01-02 15:04", s, time.Local)
	if err == nil {
		*t = JSONTime(parsed)
		return nil
	}

	// 기존 RFC3339 포맷 시도 (기존 데이터 호환)
	parsed, err = time.Parse(time.RFC3339, s)
	if err == nil {
		*t = JSONTime(parsed)
		return nil
	}

	return err
}

// JSONTime을 time.Time으로 변환합니다.
func (t JSONTime) Time() time.Time {
	return time.Time(t)
}

// t가 u보다 이후인지 확인합니다.
func (t JSONTime) After(u JSONTime) bool {
	return time.Time(t).After(time.Time(u))
}

// 현재 시간을 JSONTime으로 반환합니다.
func Now() JSONTime {
	return JSONTime(time.Now())
}
