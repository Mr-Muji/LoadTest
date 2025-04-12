package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"fmt"
	

	"github.com/Mr-Muji/LoadTest/waf-tester/backend/config"
	"github.com/Mr-Muji/LoadTest/waf-tester/backend/tester"
	"github.com/Mr-Muji/LoadTest/waf-tester/backend/gpt"
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

// StartAutoTestHandler는 URL만 입력받아 전체 과정을 자동화하는 핸들러
func StartAutoTestHandler(w http.ResponseWriter, r *http.Request) {
    // POST 방식만 허용
    if r.Method != http.MethodPost {
        http.Error(w, "허용되지 않은 메서드", http.StatusMethodNotAllowed)
        return
    }

    // URL만 포함된 단순 요청 구조체
    var req struct {
        URL string `json:"url"`
    }

    // 요청 파싱
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "잘못된 요청 형식", http.StatusBadRequest)
        return
    }

    // URL 검증
    if req.URL == "" {
        http.Error(w, "URL이 필요합니다", http.StatusBadRequest)
        return
    }

    // 자동 테스트 실행
    result, err := gpt.RunFullTest(req.URL)
    if err != nil {
        http.Error(w, fmt.Sprintf("테스트 실행 중 오류: %v", err), http.StatusInternalServerError)
        return
    }

    // 결과 반환
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(result)
}