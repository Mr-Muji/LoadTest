package tester

import (
	"bytes"
	"fmt"
	"io"
	"net/http" // 요청 보낼 때 사용
	"os"
	"strings"
	"sync" // 병렬 처리할 때 결과를 안전하게 저장하려고 mutex 사용
	"time" // 타이머 제어(Duration, Ticker 등)

	// 루트 디렉토리의 logger 패키지
	"github.com/Mr-Muji/LoadTest/waf-tester/backend/config"
	"go.uber.org/zap"         // zap 로깅 라이브러리
	"go.uber.org/zap/zapcore" // zap 설정을 위한 패키지
)

// 전역 로거 변수 선언
var log *zap.SugaredLogger

// InitLogger zap 로거를 초기화하는 함수
func InitLogger(logPath string) {
	// 로그 설정 구성
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, // 대문자로 로그 레벨 표시 (INFO, ERROR 등)
		EncodeTime:     zapcore.ISO8601TimeEncoder,  // ISO8601 시간 포맷 사용
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 로그 파일 설정
	var core zapcore.Core
	if logPath != "" {
		// 로그 디렉토리 생성
		logDir := logPath[:strings.LastIndex(logPath, "/")]
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("로그 디렉토리 생성 실패: %v\n", err)
		}

		// 로그 파일 열기
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("로그 파일 열기 실패: %v, 콘솔에만 출력합니다.\n", err)
			// 콘솔에만 출력하는 설정
			core = zapcore.NewCore(
				zapcore.NewConsoleEncoder(encoderConfig),
				zapcore.AddSync(os.Stdout),
				zap.InfoLevel,
			)
		} else {
			// 파일과 콘솔 모두에 출력
			fileWriter := zapcore.AddSync(logFile)
			consoleWriter := zapcore.AddSync(os.Stdout)
			multiWriter := zapcore.NewMultiWriteSyncer(fileWriter, consoleWriter)

			core = zapcore.NewCore(
				zapcore.NewConsoleEncoder(encoderConfig),
				multiWriter,
				zap.InfoLevel,
			)
		}
	} else {
		// 로그 파일 경로가 없으면 콘솔에만 출력
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zap.InfoLevel,
		)
	}

	// 로거 생성
	zapLogger := zap.New(core, zap.AddCaller())
	log = zapLogger.Sugar() // SugaredLogger는 사용하기 더 편리한 API 제공

	log.Info("로깅 시스템 초기화 완료")
}

func RunLoadTest(req config.TestRequest) (config.TestResult, error) {
	// 로그 시스템이 초기화되지 않은 경우를 대비
	if log == nil {
		// 기본 콘솔 로거 생성
		zapLogger, _ := zap.NewProduction()
		log = zapLogger.Sugar()
		defer zapLogger.Sync()
	}

	// 테스트 시작 로깅
	log.Infow("부하 테스트 시작",
		"target", req.Target,
		"rps", req.RPS,
		"duration", req.Duration,
	)

	// 결과를 저장할 구조체 생성
	result := config.TestResult{
		StatusMap: make(map[int]int),
	}
	// 평균 응답 시간 누적을 위한 변수
	var totalLatencySum float64

	// 요청 수를 안전하게 업데이트하기 위한 mutex(병렬 접근 대비)
	var mu sync.Mutex

	//요청을 보낼 초당 주기 설정(예: rps 30이면 초당 30)
	ticker := time.NewTicker(time.Second / time.Duration(req.RPS))
	defer ticker.Stop()

	// 테스트 시간 설정
	timeout := time.After(time.Duration(req.Duration) * time.Second)

	// WaitGroup : 모든 요청이 끝날 때까지 기다릴 수 있게 함.
	var wg sync.WaitGroup

	// 상태 업데이트를 위한 타이머 (10초마다)
	statusTicker := time.NewTicker(10 * time.Second)
	defer statusTicker.Stop()

	// 상태 업데이트 고루틴
	go func() {
		for {
			select {
			case <-statusTicker.C:
				mu.Lock()
				log.Infow("테스트 진행 상황",
					"요청수", result.TotalRequests,
					"성공", result.SuccessCount,
					"실패", result.FailCount,
				)
				mu.Unlock()
			case <-timeout:
				return
			}
		}
	}()

loop:
	for {
		select {
		case <-timeout:
			log.Infow("테스트 시간 종료", "duration", req.Duration)
			break loop
		case <-ticker.C:
			wg.Add(1)
			go func() {
				defer wg.Done()

				//경로 + 헤더 랜덤 선택
				selectedPath := GetRandomPath(req.PathList)
				url := strings.TrimRight(req.Target, "/") + "/" + strings.TrimLeft(selectedPath, "/")

				headers := GetRandomHeaderSet(req.Headers)

				// 요청 본문 설정
				var bodyReader io.Reader = nil
				if strings.ToUpper(req.Method) == "POST" && req.Body != "" {
					bodyReader = bytes.NewBuffer([]byte(req.Body))
				}

				httpReq, err := http.NewRequest(req.Method, url, bodyReader)
				if err != nil {
					log.Errorw("요청 생성 실패",
						"url", url,
						"error", err,
					)
					return
				}

				//랜덤 헤더 적용
				for k, vs := range headers {
					for _, v := range vs {
						httpReq.Header.Add(k, v)
					}
				}

				startTime := time.Now()

				// 요청 보내기
				timeoutDuration := 10 * time.Second
				if req.Timeout > 0 {
					timeoutDuration = time.Duration(req.Timeout) * time.Second
				}

				client := &http.Client{
					Timeout: timeoutDuration,
					Transport: &http.Transport{
						MaxIdleConns:        100,
						MaxIdleConnsPerHost: 100,
						IdleConnTimeout:     30 * time.Second,
					},
				}

				// 타임아웃 발생 시 처리
				resp, err := client.Do(httpReq)
				if err != nil {
					mu.Lock()
					result.TotalRequests++
					result.FailCount++

					// 타임아웃 오류 감지
					if os.IsTimeout(err) || strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
						result.TimeoutCount++
						if result.StatusMap[-1] == 0 {
							result.StatusMap[-1] = 1
						} else {
							result.StatusMap[-1]++
						}
						log.Warnw("요청 타임아웃",
							"url", url,
						)
					} else {
						log.Errorw("요청 실패",
							"url", url,
							"error", err,
						)
					}
					mu.Unlock()
					return
				}
				defer resp.Body.Close()

				latency := time.Since(startTime)
				latencyMs := float64(latency.Milliseconds())

				// 응답 코드 저장
				mu.Lock()
				result.TotalRequests++
				totalLatencySum += latencyMs

				// 응답 코드 처리
				if resp.StatusCode == 200 {
					result.SuccessCount++
					log.Debugw("요청 성공",
						"url", url,
						"statusCode", resp.StatusCode,
						"latencyMs", latencyMs,
					)
				} else {
					result.FailCount++
					statusCode := resp.StatusCode
					if statusCode >= 400 && !req.Silent {
						log.Warnw("요청 실패",
							"url", url,
							"statusCode", statusCode,
							"latencyMs", latencyMs,
						)
					}
					result.StatusMap[resp.StatusCode]++
				}

				// latency 통계 누적
				if latencyMs > result.MaxLatencyMs {
					result.MaxLatencyMs = latencyMs
					log.Infow("새로운 최대 응답시간 기록",
						"url", url,
						"latencyMs", latencyMs,
					)
				}
				if latencyMs > 500 {
					result.SlowCountOver500++
					log.Warnw("느린 응답",
						"url", url,
						"latencyMs", latencyMs,
					)
				}
				mu.Unlock()

				// 로깅 조건부 실행
				if !req.Silent {
					log.Infow("요청 결과", "status", resp.StatusCode, "latency", latencyMs)
				}
			}()
		}
	}

	// 모든 goroutine이 끝날 때까지 기다림
	log.Info("모든 요청 완료 대기 중...")
	wg.Wait()

	// 평균 응답 시간 계산
	if result.TotalRequests > 0 {
		result.AvgLatencyMs = totalLatencySum / float64(result.TotalRequests)
	}

	// 테스트 결과 요약 로깅
	log.Infow("테스트 완료",
		"총요청", result.TotalRequests,
		"성공", result.SuccessCount,
		"실패", result.FailCount,
		"타임아웃", result.TimeoutCount,
		"평균응답시간", fmt.Sprintf("%.2fms", result.AvgLatencyMs),
	)

	return result, nil
}
