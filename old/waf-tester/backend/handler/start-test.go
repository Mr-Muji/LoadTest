package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Mr-Muji/LoadTest/logger"
	"github.com/Mr-Muji/LoadTest/waf-tester/backend/config"
	"github.com/Mr-Muji/LoadTest/waf-tester/backend/gpt"
	"github.com/Mr-Muji/LoadTest/waf-tester/backend/tester"
)

// log 변수 선언
var log = logger.Logger

// StartAdvancedAutoTestHandler는 URL만 입력받아 전체 과정을 자동화하는 핸들러
// 변경된 주석
// func StartAdvancedAutoTestHandler(w http.ResponseWriter, r *http.Request) {

// POST 방식만 허용
func StartTestHandler(w http.ResponseWriter, r *http.Request) {
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
		log.Errorw("잘못된 요청 형식", "error", err)
		http.Error(w, "잘못된 요청 형식", http.StatusBadRequest)
		return
	}

	// URL 검증
	if req.URL == "" {
		log.Errorw("URL이 필요합니다")
		http.Error(w, "URL이 필요합니다", http.StatusBadRequest)
		return
	}

	// 1. 기본 테스트 실행하여 경로 추출
	autoTest, err := gpt.RunFullTest(req.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf("테스트 실행 중 오류: %v", err), http.StatusInternalServerError)
		return
	}

	// 2. 웹사이트 분석 - 통합된 함수 사용
	analysisResult, err := gpt.AnalyzeWebsite(req.URL, autoTest.ExtractedPaths)
	if err != nil {
		http.Error(w, fmt.Sprintf("웹사이트 분석 중 오류: %v", err), http.StatusInternalServerError)
		return
	}

	// 3. 첫 번째 권장 테스트 실행 (1차 테스트)
	var firstTestResult interface{}
	if len(analysisResult.RecommendedTests) > 0 {
		testReq := config.TestRequest{
			Target:   req.URL,
			Method:   analysisResult.RecommendedTests[0].Method,
			RPS:      analysisResult.RecommendedTests[0].RPS,
			Duration: analysisResult.RecommendedTests[0].Duration,
			PathList: analysisResult.RecommendedTests[0].Paths,
			Silent:   true,
		}

		firstTestResult, _ = tester.RunLoadTest(testReq)
	}

	// 4. 결과 반환
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	result := map[string]interface{}{
		"url":             req.URL,
		"analysis":        analysisResult.Analysis,
		"extractedPaths":  autoTest.ExtractedPaths,
		"recommendations": analysisResult.RecommendedTests,
	}

	if firstTestResult != nil {
		result["firstTestResult"] = firstTestResult
	}

	json.NewEncoder(w).Encode(result)
}
