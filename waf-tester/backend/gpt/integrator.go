package gpt

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Mr-Muji/LoadTest/waf-tester/backend/config"
	"github.com/Mr-Muji/LoadTest/waf-tester/backend/tester"
	"go.uber.org/zap"
)

// zap 로거 변수 선언
var log *zap.SugaredLogger

func init() {
	// 기본 zap 로거 생성 (만약 다른 곳에서 초기화되지 않았다면)
	if log == nil {
		logger, _ := zap.NewProduction()
		log = logger.Sugar()
	}
}

// AutomatedTest는 URL 기반으로 전체 테스트 과정을 자동화하는 구조체
type AutomatedTest struct {
	TargetURL      string                 // 테스트 대상 URL
	ExtractedPaths []string               // 스크래퍼로 추출한 전체 경로
	TopPaths       []PathRecommendation   // GPT 분석 후 우선순위 높은 경로들
	TestResults    map[string]interface{} // 테스트 결과
}

// RunFullTest는 URL로부터 시작하여 전체 과정을 실행하는 메소드
// 1. 스크래퍼로 경로 추출
// 2. GPT로 중요 경로 분석
// 3. 부하 테스트 실행
// 4. 결과 반환
func RunFullTest(targetURL string) (*AutomatedTest, error) {
	test := &AutomatedTest{
		TargetURL: targetURL,
	}

	// 1. API 경로 추출 (Node.js 스크래퍼 실행)
	if err := test.extractPaths(); err != nil {
		return nil, fmt.Errorf("경로 추출 실패: %v", err)
	}

	// 경로가 없으면 오류 반환
	if len(test.ExtractedPaths) == 0 {
		return nil, fmt.Errorf("추출된 API 경로가 없습니다")
	}

	// 2. GPT 분석 - 부하 가능성 높은 경로 추천
	if err := test.analyzePathsWithGPT(); err != nil {
		return nil, fmt.Errorf("GPT 분석 실패: %v", err)
	}

	// 3. 테스트 구성 생성 및 실행
	if err := test.runLoadTest(); err != nil {
		return nil, fmt.Errorf("부하 테스트 실패: %v", err)
	}

	return test, nil
}

// extractPaths는 Node.js 스크래퍼를 호출하여 API 경로를 추출하는 메소드
func (t *AutomatedTest) extractPaths() error {
	// Node.js 스크래퍼 실행 명령 구성
	cmd := exec.Command("node", "../../scrapper/api-extractor.js", t.TargetURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("스크래퍼 실행 오류: %v, 출력: %s", err, string(output))
	}

	// 출력에서 경로 목록 추출
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	// 발견된 API 엔드포인트 찾기
	var extracting bool
	for _, line := range lines {
		if strings.Contains(line, "발견된 API 엔드포인트") {
			extracting = true
			continue
		}

		if extracting && strings.HasPrefix(line, "- ") {
			// "- GET /path" 형식에서 "/path" 부분 추출
			parts := strings.SplitN(strings.TrimPrefix(line, "- "), " ", 2)
			if len(parts) == 2 {
				t.ExtractedPaths = append(t.ExtractedPaths, parts[1])
			}
		}

		if extracting && strings.Contains(line, "부하테스트 구성") {
			break
		}
	}

	return nil
}

// analyzePathsWithGPT는 GPT를 사용하여 추출된 경로들의 중요도를 분석
func (t *AutomatedTest) analyzePathsWithGPT() error {
	// 로그 추가: GPT 분석 시작
	log.Info("Starting GPT analysis", zap.Any("extractedPaths", t.ExtractedPaths))

	// GPT 분석 호출 (통합된 AnalyzeWebsite 함수 사용)
	result, err := AnalyzeWebsite(t.TargetURL, t.ExtractedPaths)
	if err != nil {
		// 로그 추가: 오류 발생 시
		log.Error("GPT analysis error", zap.Error(err))
		return fmt.Errorf("GPT 분석 오류: %v", err)
	}

	// 추천된 경로들 저장
	t.TopPaths = result.RecommendedPaths

	// 로그 추가: GPT 분석 완료
	log.Info("GPT analysis completed", zap.Any("recommendedPaths", t.TopPaths))
	return nil
}

// runLoadTest는 분석된 경로들로 부하 테스트를 실행
func (t *AutomatedTest) runLoadTest() error {
	// 테스트 요청 구성
	testReq := config.TestRequest{
		Target:   t.TargetURL,
		Method:   "GET",
		RPS:      10,
		Duration: 10,
		PathList: make([]string, 0),
		Silent:   true,
	}

	// GPT 추천 경로만 테스트 대상으로 설정
	for _, path := range t.TopPaths {
		testReq.PathList = append(testReq.PathList, path.Path)

		// 첫 번째 경로의 메소드와 RPS 추천을 사용
		if len(testReq.PathList) == 1 {
			testReq.Method = path.Method
			if path.RPS > 0 {
				testReq.RPS = path.RPS
			}
		}
	}

	// TODO: 완성된 테스트 요청으로 tester.RunLoadTest 호출
	// 여기서는 구현 생략 (이미 waf-tester에 구현되어 있음)

	return nil
}

// RunRecommendedTest는 특정 테스트 권장사항에 따라 테스트를 실행
func RunRecommendedTest(targetURL string, recommendation TestRecommendation) (interface{}, error) {
	// 테스트 요청 구성
	testReq := config.TestRequest{
		Target:   targetURL,
		Method:   recommendation.Method,
		RPS:      recommendation.RPS,
		Duration: recommendation.Duration,
		PathList: recommendation.Paths,
		Silent:   true, // 로깅 비활성화 옵션 추가
	}

	// 추천 테스트 실행 로그
	log.Infow("Running recommended test",
		"targetURL", targetURL,
		"recommendation", recommendation,
	)

	// 테스트 유형에 따라 다른 테스트 실행
	var result interface{}
	var err error
	switch recommendation.Type {
	case "load":
		result, err = tester.RunLoadTest(testReq)
	case "security":
		result, err = simulateSecurityTest(testReq)
	case "functional":
		result, err = simulateFunctionalTest(testReq)
	default:
		result, err = tester.RunLoadTest(testReq)
	}

	// 테스트 실행 완료 로그
	log.Infow("Test execution completed",
		"result", result,
	)

	return result, err
}

// simulateSecurityTest는 보안 테스트를 시뮬레이션 (실제 구현 필요)
func simulateSecurityTest(req config.TestRequest) (map[string]interface{}, error) {
	// 보안 테스트 시뮬레이션
	return map[string]interface{}{
		"testType":    "security",
		"description": "보안 테스트 시뮬레이션 (향후 실제 보안 테스트 구현 예정)",
		"paths":       req.PathList,
		"method":      req.Method,
		"findings":    []string{"구현 예정입니다"},
	}, nil
}

// simulateFunctionalTest는 기능 테스트를 시뮬레이션 (실제 구현 필요)
func simulateFunctionalTest(req config.TestRequest) (map[string]interface{}, error) {
	// 기능 테스트 시뮬레이션
	return map[string]interface{}{
		"testType":    "functional",
		"description": "기능 테스트 시뮬레이션 (향후 실제 기능 테스트 구현 예정)",
		"paths":       req.PathList,
		"method":      req.Method,
		"results":     []string{"구현 예정입니다"},
	}, nil
}
