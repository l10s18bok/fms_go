// Package translations는 한국어 번역 데이터를 임베드합니다.
package translations

import _ "embed"

//go:embed base.ko.json
var KoJSON []byte
