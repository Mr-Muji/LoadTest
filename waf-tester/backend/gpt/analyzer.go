package gpt

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// AnalysisResult는 GPT가 분석한 API 경로 추천 결과를 담는 구조체
type AnalysisResult struct {
	RecommendedPaths []PathRecommendation `json:"recommendedPaths"` // 추천 경로 목록
}

// PathRecommendation은 개별 경로에 대한 추천 정보
type PathRecommendation struct {
	Path        string `json:"path"`        // API 경로
	Method      string `json:"method"`      // HTTP 메서드
	Priority    int    `json:"priority"`    // 우선순위 (1: 높음, 5: 낮음)
	Reason      string `json:"reason"`      // 추천 이유
	RPS         int    `json:"rps"`         // 추천 초당 요청 수
	Description string `json:"description"` // 설명
}

// RecommendHotEndpoints는 API 경로 목록을 GPT에게 전달하여
// 부하 테스트에 가장 적합한 경로와 설정을 추천받는 함수
func RecommendHotEndpoints(pathList []string) (*AnalysisResult, error) {
	// OPENAI_API_KEY 환경변수 확인
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY 환경변수가 설정되지 않았습니다")
	}

	// OpenAI 클라이언트 생성
	client := openai.NewClient(apiKey)
	ctx := context.Background()

	// API 경로 정보를 문자열로 변환
	pathsInfo := strings.Join(pathList, "\n- ")
	pathsInfo = "- " + pathsInfo

	// GPT에 전달할 프롬프트 작성
	prompt := fmt.Sprintf(`You are an API performance expert. Analyze these API endpoints and identify those most likely to cause high server load:

%s

For each endpoint, consider:
1. Database operations (heavy queries)
2. Computational complexity
3. Data transfer size
4. Cache-friendliness
5. Authentication overhead

Return a JSON array of recommendations with the following structure:
{
  "recommendedPaths": [
    {
      "path": "/api/endpoint",
      "method": "GET",
      "priority": 1,
      "reason": "Likely contains heavy database joins",
      "rps": 30,
      "description": "Start with this endpoint as it's likely to cause load"
    }
  ]
}

Sort the results by priority (1=highest, 5=lowest). Limit to 5 endpoints maximum.`, pathsInfo)

	// ChatGPT API 호출
	response, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an API performance testing expert. Return only valid JSON responses.",
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

	// 응답에서 JSON 부분 추출
	content := response.Choices[0].Message.Content
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return nil, fmt.Errorf("응답에서 유효한 JSON을 찾을 수 없습니다: %s", content)
	}

	jsonStr := content[jsonStart : jsonEnd+1]

	// JSON 파싱
	var result AnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("JSON 파싱 중 오류: %v, JSON: %s", err, jsonStr)
	}

	return &result, nil
}
