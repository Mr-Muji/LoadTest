// logger/logger.go
package logger // 최상위 디렉토리에 위치한 logger 패키지입니다.

// go.uber.org/zap 라이브러리를 사용하여 고성능, 구조화 된 로그를 제공합니다.
import (
	"go.uber.org/zap" // zap 라이브러리 임포트
)

// Logger 전역 변수 선언
// 이 변수는 모든 패키지에서 동일한 로거 인스턴스를 공유하기 위해 사용됩니다.
var Logger *zap.SugaredLogger

// Init 함수는 zap 로거를 초기화하며,
// 초기화된 로거를 전역 변수 Logger에 할당하여 다른 패키지에서도 사용 가능하게 만듭니다.
func Init() {
	// Production 환경에 맞춰 zap 로거 생성
	// 개발 단계에서는 zap.NewDevelopment()를 사용할 수 있습니다.
	l, err := zap.NewProduction()
	if err != nil {
		// 로거 초기화에 실패할 경우, 패닉을 발생시켜 애플리케이션을 종료합니다.
		panic(err)
	}
	// zap.Logger를 SugaredLogger 형태로 변환 후, 전역 변수에 저장합니다.
	Logger = l.Sugar()
}