package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Mr-Muji/LoadTest/backend/config"
	"github.com/Mr-Muji/LoadTest/backend/modules/ai"
	loadtest "github.com/Mr-Muji/LoadTest/backend/modules/load-test"
	"github.com/Mr-Muji/LoadTest/backend/modules/orchestrator"
	"github.com/Mr-Muji/LoadTest/libs/logger"
)

// log 변수 선언
var log = logger.Logger

// HandleStartTest는 기본 부하 테스트를 시작하는 핸들러
func HandleStartTest(w http.ResponseWriter, r *http.Request) {
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

	// 테스트 요청 생성
	testReq := config.TestRequest{
		Target:   req.URL,
		Method:   "GET",
		RPS:      10,
		Duration: 30,
		PathList: []string{"/"},
		Silent:   false,
	}

	// 부하 테스트 실행
	result, err := loadtest.RunLoadTest(testReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("테스트 실행 중 오류: %v", err), http.StatusInternalServerError)
		return
	}

	// 결과 반환
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleAdvancedAutoTest는 URL만 입력받아 전체 과정을 자동화하는 핸들러
func HandleAdvancedAutoTest(w http.ResponseWriter, r *http.Request) {
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

	// 1. 기본 테스트 실행하여 경로 추출 (orchestrator 모듈 사용)
	autoTest, err := orchestrator.RunFullTest(req.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf("테스트 실행 중 오류: %v", err), http.StatusInternalServerError)
		return
	}

	// 2. 웹사이트 분석 - 통합된 함수 사용 (ai 모듈 사용)
	analysisResult, err := ai.AnalyzeWebsite(req.URL, autoTest.ExtractedPaths)
	if err != nil {
		http.Error(w, fmt.Sprintf("웹사이트 분석 중 오류: %v", err), http.StatusInternalServerError)
		return
	}

	// 3. 첫 번째 권장 테스트 실행 (1차 테스트) (orchestrator 모듈 사용)
	var firstTestResult interface{}
	if len(analysisResult.RecommendedTests) > 0 {
		firstTestResult, err = orchestrator.RunRecommendedLoadTest(
			req.URL,
			analysisResult.RecommendedTests[0],
		)
		if err != nil {
			log.Warnw("권장 테스트 실행 중 오류", "error", err)
		}
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
