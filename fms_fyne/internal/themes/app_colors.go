// Package themes는 애플리케이션의 색상 및 테마를 정의합니다.
package themes

import "image/color"

// 기본 색상 정의
var Colors = map[string]color.Color{
	// 기본 색상
	"blue":   color.RGBA{R: 0, G: 123, B: 255, A: 255},
	"green":  color.RGBA{R: 40, G: 167, B: 69, A: 255},
	"red":    color.RGBA{R: 220, G: 53, B: 69, A: 255},
	"yellow": color.RGBA{R: 255, G: 193, B: 7, A: 255},
	"orange": color.RGBA{R: 255, G: 128, B: 0, A: 255},
	"purple": color.RGBA{R: 128, G: 0, B: 255, A: 255},
	"cyan":   color.RGBA{R: 0, G: 188, B: 212, A: 255},
	"pink":   color.RGBA{R: 233, G: 30, B: 99, A: 255},

	// 회색 계열
	"gray":      color.RGBA{R: 100, G: 100, B: 100, A: 255}, // ButtonSecondary
	"darkgray":  color.RGBA{R: 80, G: 80, B: 80, A: 255},    // ButtonDark
	"lightgray": color.RGBA{R: 200, G: 200, B: 200, A: 255},
	"black":     color.RGBA{R: 40, G: 40, B: 40, A: 255}, // ButtonBlack
	"white":     color.White,
}

// 색상 이름으로 색상을 반환합니다. 없으면 흰색을 반환합니다.
func GetColor(name string) color.Color {
	if c, ok := Colors[name]; ok {
		return c
	}
	return color.White
}
