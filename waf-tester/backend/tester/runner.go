package tester

import (
	"bytes"
	"fmt"
	"io"
	"net/http" // 요청 보낼 때 사용
	"strings"
	"sync" // 병렬 처리할 때 결과를 안전하게 저장하려고 mutex 사용
	"time" // 타이머 제어(Duration, Ticker 등)

	"github.com/Mr-Muji/LoadTest/waf-tester/backend/config"
)

func RunLoadTest(req config.TestRequest) (config.TestResult, error) {
	// 결과를 저장할 구조체 생성
	result := config.TestResult{
		StatusMap: make(map[int]int),
	}
	// 평균 응답 시간 누적을 위한 변수
	var totalLatencySum float64

	// 요청 수를 안전하게 업데이트하기 위한 mutex(병렬 접근 대비)
	var mu sync.Mutex

	//요청을 보낼 초당 주기 설정(예: rps 30이면 초당 30)
	ticker := time.NewTicker(time.Second / time.Duration(req.RPS))
	defer ticker.Stop()

	// 테스트 시간 설정
	timeout := time.After(time.Duration(req.Duration) * time.Second)

	// WaitGroup : 모든 요청이 끝날 때까지 기다릴 수 있게 함.
	var wg sync.WaitGroup
loop: // 'loop'는 레이블(label)로, Go에서 특정 반복문에 이름을 붙여 제어하는 기능
	for { // 무한 반복문 시작
		select { // select는 여러 채널 연산 중 준비된 것을 처리하는 Go의 특별 구문
		case <-timeout: // timeout 채널에서 값을 받으면 (시간 초과 발생)
			break loop // loop 레이블이 붙은 반복문을 종료합니다 (일반 break는 select만 빠져나감)
		case <-ticker.C: // ticker의 채널 C에서 값을 받을 때마다 (일정 시간 간격)
			wg.Add(1)   // WaitGroup 카운터 증가 (고루틴 추가)
			go func() { // 새 고루틴(경량 스레드) 시작 - 병렬 처리를 위함
				defer wg.Done() // 함수 종료 시 WaitGroup 카운터 감소

				//경로 + 헤더 랜덤 선택
				selectedPath := GetRandomPath(req.PathList)
				url := strings.TrimRight(req.Target, "/") + "/" + strings.TrimLeft(selectedPath, "/")

				headers := GetRandomHeaderSet(req.Headers) // 헤더 랜덤 선택 수정

				// 요청 본문 설정
				var bodyReader io.Reader = nil // 변수명은 유지, 타입만 변경
				if strings.ToUpper(req.Method) == "POST" && req.Body != "" {
					bodyReader = bytes.NewBuffer([]byte(req.Body)) // 기존 로직 유지
				}

				httpReq, err := http.NewRequest(req.Method, url, bodyReader) // 변수명 유지
				if err != nil {                                              // 요청 객체 생성 실패 시
					fmt.Println("요청 생성 실패:", err) // 오류 출력
					return                        // 고루틴 종료
				}

				//랜덤 헤더 적용
				for k, vs := range headers { // 헤더 맵을 순회 (키와 값 배열)
					for _, v := range vs { // 각 헤더 값 배열을 순회
						httpReq.Header.Add(k, v) // 요청에 헤더 추가
					}
				}
				// resp 앞에 있어야 함
				startTime := time.Now()

				// 요청 보내기
				client := &http.Client{}        // HTTP 클라이언트 생성
				resp, err := client.Do(httpReq) // 요청 전송하고 응답 받기
				if err != nil {                 // 요청 실패 시
					return // 고루틴 종료
				}
				defer resp.Body.Close() // 함수 종료 시 응답 본문 닫기 (리소스 정리)

				latency := time.Since(startTime)
				latencyMs := float64(latency.Milliseconds())

				// 응답 코드 저장
				mu.Lock()              // 뮤텍스 잠금 (동시 접근 방지)
				result.TotalRequests++ // 총 요청 수 증가
				totalLatencySum += latencyMs
				// 응답 코드 처리
				if resp.StatusCode == 200 { // 성공(200) 응답이면
					result.SuccessCount++ // 성공 카운트 증가
				} else { // 그 외 응답 코드는
					result.FailCount++ // 실패 카운트 증가
				}
				result.StatusMap[resp.StatusCode]++ // 응답 코드별 카운트 증가

				// latency 통계 누적
				if latencyMs > result.MaxLatencyMs {
					result.MaxLatencyMs = latencyMs
				}
				if latencyMs > 500 {
					result.SlowCountOver500++
				}
				mu.Unlock() // 뮤텍스 잠금 해제
			}() // 고루틴 함수 종료
		}
	}

	//모든 goroutine이 끝날 때까지 기다림
	wg.Wait() // 모든 고루틴이 작업을 마칠 때까지 대기

	// 평균 응답 시간 계산
	if result.TotalRequests > 0 {
		result.AvgLatencyMs = totalLatencySum / float64(result.TotalRequests)
	}

	return result, nil // 테스트 결과와 nil 에러 반환
}
