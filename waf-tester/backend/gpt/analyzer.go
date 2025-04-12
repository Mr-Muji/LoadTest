package gpt

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

// log 변수 선언
var log = *zap.SugaredLogger

// PathRecommendation은 GPT가 추천하는 경로 정보를 담는 구조체
type PathRecommendation struct {
	Path        string `json:"path"`        // API 경로
	Method      string `json:"method"`      // HTTP 메서드
	Priority    int    `json:"priority"`    // 우선순위 (1: 높음, 5: 낮음)
	Reason      string `json:"reason"`      // 추천 이유
	RPS         int    `json:"rps"`         // 추천 초당 요청 수
	Description string `json:"description"` // 설명
}

// TestRecommendation은 권장 테스트 정보를 담는 구조체
type TestRecommendation struct {
	Type        string   `json:"type"`        // 테스트 유형 (load, security, functional 등)
	Paths       []string `json:"paths"`       // 테스트할 경로들
	Method      string   `json:"method"`      // HTTP 메서드
	RPS         int      `json:"rps"`         // 초당 요청 수
	Duration    int      `json:"duration"`    // 테스트 지속 시간(초)
	Description string   `json:"description"` // 테스트 설명
}

// WebsiteAnalysisResult는 웹사이트 분석 통합 결과를 담는 구조체
type WebsiteAnalysisResult struct {
	Analysis         string               `json:"analysis"`         // 자연어 분석 결과
	RecommendedTests []TestRecommendation `json:"recommendedTests"` // 추천 테스트 전략
	RecommendedPaths []PathRecommendation `json:"recommendedPaths"` // 부하 테스트 우선순위 경로
}

// AnalyzeWebsite는 URL과 경로 목록을 분석하여 모든 정보를 한 번에 반환하는 함수
func AnalyzeWebsite(url string, extractedPaths []string) (*WebsiteAnalysisResult, error) {
	// 함수 시작 로깅
	log.Infow("AnalyzeWebsite 함수 시작",
		"url", url,
		"extractedPathsCount", len(extractedPaths))

	// OpenAI API 키 확인
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Error("OPENAI_API_KEY 환경변수 없음")
		return nil, fmt.Errorf("OPENAI_API_KEY 환경변수가 설정되지 않았습니다")
	}
	log.Debug("OPENAI_API_KEY 환경변수 확인 완료")

	// OpenAI 클라이언트 생성
	client := openai.NewClient(apiKey)
	ctx := context.Background()
	log.Debug("OpenAI 클라이언트 생성 완료")

	// 경로 정보를 문자열로 변환
	pathsInfo := strings.Join(extractedPaths, "\n- ")
	pathsInfo = "- " + pathsInfo
	log.Debug("경로 정보 문자열 변환 완료",
		"경로 수", len(extractedPaths),
		"첫 번째 경로", extractedPaths[0])

	// GPT에 전달할 프롬프트 작성 (영어로)
	prompt := fmt.Sprintf(`Website URL: %s
	
Discovered API endpoints:
%s

You are a web application analysis expert. Analyze the website and API endpoints above and provide:

1. A comprehensive analysis in Korean language (within 500 characters). Explain what kind of service this website is, what API structure it has, and what types of tests would be useful.

2. Recommended test strategies in a JSON structure. Respond with the following structure:

{
  "analysis": "한국어로 된 분석 결과를 이 곳에 작성하세요...",
  "recommendedPaths": [
    {
      "path": "/api/products",
      "method": "GET",
      "priority": 1,
      "reason": "Contains database join operations",
      "rps": 50,
      "description": "Product listing API, high load expected"
    }
  ],
  "recommendedTests": [
    {
      "type": "load",
      "paths": ["/api/products", "/api/users"],
      "method": "GET",
      "rps": 50,
      "duration": 30,
      "description": "Load test for product listing API"
    },
    {
      "type": "security",
      "paths": ["/api/login", "/api/register"],
      "method": "POST",
      "rps": 20,
      "duration": 30,
      "description": "Security test for authentication endpoints"
    }
  ]
}

Specify path priorities from 1 (highest) to 5 (lowest), and provide at least 3 different test types.`, url, pathsInfo)

	log.Infow("GPT 분석 요청 준비 완료",
		"url", url,
		"prompt길이", len(prompt))

	// 프롬프트의 처음과 끝 부분 로깅 (전체 프롬프트가 너무 길 수 있음)
	promptPreview := prompt
	if len(promptPreview) > 500 {
		promptPreview = promptPreview[:250] + "..." + promptPreview[len(promptPreview)-250:]
	}
	log.Debug("프롬프트 내용(일부)", "prompt", promptPreview)

	// API 요청 시작 시간 기록
	requestStart := time.Now()
	log.Infow("ChatGPT API 요청 시작", "시작시간", requestStart)

	// ChatGPT API 호출
	response, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a web security and performance testing expert. You analyze websites and recommend test strategies. Always respond with valid JSON and make sure to write the analysis part in Korean language.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.2, // 낮은 온도로 일관된 응답 유도
		},
	)

	// API 응답 시간 및 결과 로깅
	responseTime := time.Since(requestStart)
	log.Infow("ChatGPT API 응답 수신",
		"소요시간", responseTime,
		"성공여부", err == nil)

	if err != nil {
		log.Errorw("GPT API 호출 오류",
			"error", err,
			"경과시간", responseTime)
		return nil, fmt.Errorf("OpenAI API 호출 중 오류: %v", err)
	}

	// 응답 내용 로깅
	content := response.Choices[0].Message.Content
	contentPreview := content
	if len(contentPreview) > 500 {
		contentPreview = contentPreview[:250] + "..." + contentPreview[len(contentPreview)-250:]
	}
	log.Infow("GPT 응답 내용(일부)", "content", contentPreview)
	log.Debug("GPT 모델", "model", response.Model)
	log.Debug("GPT 토큰 사용량",
		"프롬프트 토큰", response.Usage.PromptTokens,
		"완성 토큰", response.Usage.CompletionTokens,
		"총 토큰", response.Usage.TotalTokens)

	// 응답에서 JSON 부분 추출
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		log.Errorw("JSON 형식 추출 실패",
			"content", content,
			"jsonStart", jsonStart,
			"jsonEnd", jsonEnd)
		return nil, fmt.Errorf("응답에서 유효한 JSON을 찾을 수 없습니다: %s", content)
	}

	jsonStr := content[jsonStart : jsonEnd+1]
	log.Debug("JSON 추출 완료", "jsonLength", len(jsonStr))

	// JSON 파싱 시작
	log.Infow("JSON 파싱 시작")
	var result WebsiteAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Errorw("JSON 파싱 오류",
			"error", err,
			"jsonStr", jsonStr)
		return nil, fmt.Errorf("JSON 파싱 중 오류: %v, JSON: %s", err, jsonStr)
	}

	// 파싱 결과 로깅
	log.Infow("JSON 파싱 성공",
		"recommendedPathsCount", len(result.RecommendedPaths),
		"recommendedTestsCount", len(result.RecommendedTests))

	// 분석 결과 미리보기
	analysisPreview := result.Analysis
	if len(analysisPreview) > 100 {
		analysisPreview = analysisPreview[:97] + "..."
	}
	log.Infow("분석 결과(일부)", "analysis", analysisPreview)

	// 추천 경로 로깅
	for i, path := range result.RecommendedPaths {
		if i < 3 { // 처음 3개만 로깅
			log.Infow("추천 경로",
				"index", i,
				"path", path.Path,
				"method", path.Method,
				"priority", path.Priority,
				"rps", path.RPS)
		}
	}

	// 추천 테스트 로깅
	for i, test := range result.RecommendedTests {
		if i < 3 { // 처음 3개만 로깅
			log.Infow("추천 테스트",
				"index", i,
				"type", test.Type,
				"method", test.Method,
				"rps", test.RPS,
				"duration", test.Duration)
		}
	}

	log.Infow("AnalyzeWebsite 함수 완료", "총소요시간", time.Since(requestStart))
	return &result, nil
}
