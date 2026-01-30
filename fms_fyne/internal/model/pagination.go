package model

// PageRequest 페이지네이션 요청
type PageRequest struct {
	Offset  int    // 시작 위치 (0-based)
	Limit   int    // 가져올 건수 (= UI PageSize)
	Filter  string // 유형 필터 ("firewall", "program", "" = 전체)
	Keyword string // 검색 키워드 ("" = 검색 없음)
}

// PageResult 페이지네이션 응답 (제네릭)
type PageResult[T any] struct {
	Items      []*T // 조회된 데이터
	TotalCount int  // 필터/검색 적용 후 전체 건수
}
