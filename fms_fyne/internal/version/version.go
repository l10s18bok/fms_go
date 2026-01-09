// Package version은 애플리케이션 버전 정보를 관리합니다.
package version

// 애플리케이션 버전 정보
const (
	AppName    = "FMS"
	AppVersion = "1.1.0"
)

// 앱 이름과 버전을 포함한 문자열을 반환합니다.
func GetVersionString() string {
	return AppName + " v" + AppVersion
}
