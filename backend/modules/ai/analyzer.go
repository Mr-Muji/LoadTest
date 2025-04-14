package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

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
	// OpenAI API 키 확인
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY 환경변수가 설정되지 않았습니다")
	}

	// OpenAI 클라이언트 생성
	client := openai.NewClient(apiKey)
	ctx := context.Background()

	// 경로 정보를 문자열로 변환
	pathsInfo := strings.Join(extractedPaths, "\n- ")
	pathsInfo = "- " + pathsInfo

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

	// API 요청 시작 시간 기록
	requestStart := time.Now()

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

	if err != nil {
		return nil, fmt.Errorf("OpenAI API 호출 중 오류: %v", err)
	}

	// API 요청 시간 계산 및 로그 출력
	requestDuration := time.Since(requestStart)
	fmt.Printf("OpenAI API 요청 소요 시간: %v\n", requestDuration)

	// 응답 내용
	content := response.Choices[0].Message.Content

	// 응답에서 JSON 부분 추출
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return nil, fmt.Errorf("응답에서 유효한 JSON을 찾을 수 없습니다: %s", content)
	}

	jsonStr := content[jsonStart : jsonEnd+1]

	// JSON 파싱
	var result WebsiteAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("JSON 파싱 중 오류: %v, JSON: %s", err, jsonStr)
	}

	return &result, nil
}
