// package main - Go 프로그램의 진입점을 정의하는 패키지
// Go 애플리케이션은 반드시 main 패키지와 main 함수가 있어야 실행 가능합니다
package main

import (
	// Go에서는 필요한 기능을 패키지로 가져와 사용합니다
	"fmt"           // 콘솔에 문자열 출력할 때 사용 (printf, println 등)
	stdlog "log"    // 표준 log 패키지를 stdlog로 별칭
	"net/http"      // HTTP 서버/클라이언트 기능 제공 (웹 서버 만들 때 필요)
	"os"            // 파일 시스템 접근용
	"path/filepath" // 경로 처리용
	"time"          // 날짜 포맷팅용

	// 패키지 경로 수정 (service-test/ 제거)
	api "github.com/Mr-Muji/LoadTest/backend/api/load-test"          // 로드 테스트 API 핸들러
	loadtest "github.com/Mr-Muji/LoadTest/backend/modules/load-test" // 로드 테스트 모듈
	"github.com/Mr-Muji/LoadTest/libs/logger"                        // 로깅 모듈
	"github.com/joho/godotenv"
	"go.uber.org/zap" // zap 로거 패키지
)

// 짧은 별칭 변수 선언 - logger.Logger 대신 log 사용
var log *zap.SugaredLogger

// TestRequest - 클라이언트로부터 받을 테스트 요청 정보를 담는 구조체
// Go에서 구조체(struct)는 관련 데이터를 하나로 묶는 자료형입니다
type TestRequest struct {
	// `json:"필드명"` 형태의 태그는 JSON과 Go 구조체 간 변환 규칙을 정의합니다
	Target   string `json:"target"`   // 테스트할 URL (예: https://example.com)
	RPS      int    `json:"rps"`      // 초당 요청 수 (Requests Per Second)
	Duration int    `json:"duration"` // 테스트 지속 시간(초 단위)
}

func init() {
	// .env 파일 로드
	envFiles := []string{
		".env",          // 현재 디렉토리
		"../../.env",    // 프로젝트 루트
		"../../../.env", // 더 상위 디렉토리
	}

	for _, file := range envFiles {
		err := godotenv.Load(file)
		if err == nil {
			stdlog.Println("환경 변수 로드 성공:", file)
			break
		}
	}

	// 로그 디렉토리 설정
	logDir := "./logs"

	// 현재 실행 경로 기준으로 로그 디렉토리 생성
	executable, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(executable)
		logDir = filepath.Join(execDir, "logs")
	}

	// 로그 디렉토리 생성
	if err := os.MkdirAll(logDir, 0755); err != nil {
		stdlog.Printf("로그 디렉토리 생성 실패: %v", err)
	}

	// 날짜별 로그 파일 경로 생성
	timestamp := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, fmt.Sprintf("loadtest_%s.log", timestamp))

	// zap 로거 초기화 (load-test 모듈에 구현)
	loadtest.InitLogger(logPath)

	// 루트 디렉토리의 logger 패키지 초기화
	logger.Init()

	// logger.Logger를 log 변수에 할당
	log = logger.Logger

	// 초기화 완료 로그
	log.Info("애플리케이션 시작, 로깅 시스템 초기화 완료")
}

// main - 프로그램의 진입점이 되는 함수
func main() {
	// 종료 시 로그 버퍼 비우기
	defer log.Sync()

	// 서버 시작 로그
	log.Info("서버 실행 중: http://localhost:8080")
	// fmt.Println("🚀 서버 실행 중 : http://localhost:8080")

	// API 라우트 설정
	http.HandleFunc("/test", api.HandleStartTest)
	http.HandleFunc("/advanced-auto-test", api.HandleAdvancedAutoTest)

	// 8080 포트에서 HTTP 서버 시작
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalw("서버 실행 실패", "error", err)
	}
}
