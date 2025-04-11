package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Mr-Muji/LoadTest/waf-tester/backend/config"
	"github.com/Mr-Muji/LoadTest/waf-tester/backend/tester"
)

// StartTestHandler는 /start-test 라우터에 대응되는 HTTP 핸들러다.
// 프론트엔드나 curl에서 설정을 JSON으로 받아서 부하 테스트를 실행한다.
func StartTestHandler(w http.ResponseWriter, r *http.Request) {
	// 요청은 반드시 POST 방식이어야 함
	if r.Method != http.MethodPost {
		http.Error(w, "허용되지 않은 메서드", http.StatusMethodNotAllowed)
		return
	}

	// 요청 본문(JSON)을 구조체로 파싱
	var req config.TestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "잘못된 요청 형식", http.StatusBadRequest)
		return
	}

	log.Printf("테스트 시작: Target=%s | RPS=%d | Duration=%d초\n", req.Target, req.RPS, req.Duration)

	// 트래픽 테스트 실행
	result, err := tester.RunLoadTest(req)
	if err != nil {
		http.Error(w, "테스트 실행 중 오류 발생", http.StatusInternalServerError)
		log.Printf("❌ RunLoadTest 에러: %v\n", err)
		return
	}

	// 결과를 JSON으로 응답
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("❌ 결과 응답 실패: %v\n", err)
	}
}
