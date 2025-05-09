package config

// TestRequest는 /start-test API로부터 받은 테스트 설정을 담는 구조체
type TestRequest struct {
	Target   string              `json:"target"`             // 테스트 대상 도메인 (예: https://example.com)
	RPS      int                 `json:"rps"`                // 초당 요청 수 (Requests Per Second)
	Duration int                 `json:"duration"`           // 테스트 시간 (초)
	Method   string              `json:"method"`             // 요청 메서드 (GET, POST 등)
	Headers  map[string][]string `json:"headers,omitempty"`  // 사용할 HTTP 헤더 세트 (랜덤 선택용)
	PathList []string            `json:"pathList,omitempty"` // 다양한 요청 경로 리스트
	Body     string              `json:"body"`               // 요청 본문 (POST 요청에만 사용)
	Timeout  int                 `json:"timeout,omitempty"`  // 요청별 타임아웃(초)
	Silent   bool                `json:"silent,omitempty"`   // true면 요청별 로깅 비활성화
}

// TestResult는 트래픽 실행 후 응답 상태를 요약한 결과 구조체(백이 프론트한테 보냄)
type TestResult struct {
	TotalRequests    int         `json:"totalRequests"`    // 총 요청 수
	SuccessCount     int         `json:"successCount"`     // 200 응답 수
	FailCount        int         `json:"failCount"`        // 200 외 응답 수 (403, 429 등)
	TimeoutCount     int         `json:"timeoutCount"`     // 타임아웃 발생 수
	StatusMap        map[int]int `json:"statusMap"`        // 응답 코드별 개수 (예: 200:123, 429:4)
	AvgLatencyMs     float64     `json:"avgLatencyMs"`     // 평균 응답 시간
	MaxLatencyMs     float64     `json:"maxLatencyMs"`     // 최대 응답 시간
	SlowCountOver500 int         `json:"slowCountOver500"` // 500ms 초과한 요청 개수
}
