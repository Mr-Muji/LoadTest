// package main - Go 프로그램의 진입점을 정의하는 패키지
// Go 애플리케이션은 반드시 main 패키지와 main 함수가 있어야 실행 가능합니다
package main

import (
	// Go에서는 필요한 기능을 패키지로 가져와 사용합니다
	"encoding/json" // JSON 문자열을 Go 구조체로 변환하는 데 필요
	"fmt"           // 콘솔에 문자열 출력할 때 사용 (printf, println 등)
	"log"           // 로깅 기능 제공 (에러나 정보를 기록할 때 사용)
	"net/http"      // HTTP 서버/클라이언트 기능 제공 (웹 서버 만들 때 필요)

)

// TestRequest - 클라이언트로부터 받을 테스트 요청 정보를 담는 구조체
// Go에서 구조체(struct)는 관련 데이터를 하나로 묶는 자료형입니다
type TestRequest struct {
	// `json:"필드명"` 형태의 태그는 JSON과 Go 구조체 간 변환 규칙을 정의합니다
	Target   string `json:"target"`   // 테스트할 URL (예: https://example.com)
	RPS      int    `json:"rps"`      // 초당 요청 수 (Requests Per Second)
	Duration int    `json:"duration"` // 테스트 지속 시간(초 단위)
}

// startTestHandler - /start-test 엔드포인트에 대한 HTTP 요청을 처리하는 핸들러 함수
// w: HTTP 응답을 작성하기 위한 객체
// r: 클라이언트로부터 받은 HTTP 요청 정보
func startTestHandler(w http.ResponseWriter, r *http.Request) {
	// 클라이언트 요청을 저장할 변수 선언
	var req TestRequest

	// HTTP 요청 본문(body)을 JSON으로 파싱하여 TestRequest 구조체에 저장
	// json.NewDecoder: JSON 데이터를 읽기 위한 디코더 생성
	// r.Body: HTTP 요청의 본문
	// Decode(&req): JSON 데이터를 req 변수에 저장
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// JSON 파싱에 실패하면 400 Bad Request 오류 응답 반환
		http.Error(w, "잘못된 요청", http.StatusBadRequest)
		return
	}

	// 요청 정보를 로그로 출력 (서버 콘솔에 표시)
	log.Printf("[REQUEST] target=%s rps=%d duration=%d\n", req.Target, req.RPS, req.Duration)

	// 클라이언트에 200 OK 상태 코드 전송
	w.WriteHeader(http.StatusOK)

	// 응답 본문으로 "OK" 문자열 전송
	w.Write([]byte("OK"))
}

// main - 프로그램의 진입점이 되는 함수
func main() {
	// /start-test 경로로 들어오는 HTTP 요청을 startTestHandler 함수로 처리하도록 등록
	http.HandleFunc("/start-test", startTestHandler)

	// 서버가 시작되었음을 콘솔에 출력
	fmt.Println("🚀 서버 실행 중 : http://localhost:8080")

	// 8080 포트에서 HTTP 서버 시작
	// ListenAndServe 함수는 서버가 종료될 때까지 블로킹됨
	// 서버 실행 중 오류가 발생하면 log.Fatal이 오류 메시지를 출력하고 프로그램 종료
	log.Fatal(http.ListenAndServe(":8080", nil))
}
