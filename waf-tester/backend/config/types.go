package config

// TestRequest는 /start-test API로부터 받은 테스트 설정을 담는 구조체
type TestRequest struct {
	Target   string              `json:"target"`             // 테스트 대상 도메인 (예: https://example.com)
	RPS      int                 `json:"rps"`                // 초당 요청 수 (Requests Per Second)
	Duration int                 `json:"duration"`           // 테스트 시간 (초)
	Method   string              `json:"method"`             // 요청 메서드 (GET, POST 등)
	Headers  map[string][]string `json:"headers,omitempty"`  // 사용할 HTTP 헤더 세트 (랜덤 선택용)
	PathList []string            `json:"pathList,omitempty"` // 다양한 요청 경로 리스트
}

// TestResult는 트래픽 실행 후 응답 상태를 요약한 결과 구조체(백이 프론트한테 보냄)
type TestResult struct {
	TotalRequests int         `json:"totalRequests"` // 총 요청 수
	SuccessCount  int         `json:"successCount"`  // 200 응답 수
	FailCount     int         `json:"failCount"`     // 200 외 응답 수 (403, 429 등)
	StatusMap     map[int]int `json:"statusMap"`     // 응답 코드별 개수 (예: 200:123, 429:4)
}
