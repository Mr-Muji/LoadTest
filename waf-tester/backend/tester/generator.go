package tester // 이 파일이 tester 패키지에 속함을 선언

import (
	"math/rand" // 랜덤 숫자 생성을 위한 패키지
)

// init 함수는 프로그램 시작 시 자동으로 호출되는 특별한 함수입니다.
// 패키지가 처음 로드될 때 단 한 번만 실행됩니다.
func init() {
	// Go 1.20 이상에서는 전역 랜덤 생성기에 자동으로 시드가 설정되므로
	// 명시적인 시드 설정이 필요 없습니다.

	// Go 1.20 미만 버전을 사용하는 경우, 아래와 같이 사용:
	// r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 또는 글로벌 시드를 직접 설정:
	// rand.New(rand.NewSource(time.Now().UnixNano()))
}

// GetRandomPath 함수는 주어진 경로 목록(paths) 중에서 무작위로 하나를 선택하여 반환합니다.
// 부하 테스트를 수행할 때 다양한 URL 경로로 요청을 보내기 위해 사용됩니다.
// 매개변수:
// - paths []string: 선택할 경로들의 문자열 배열
// 반환값:
// - string: 선택된 랜덤 경로
func GetRandomPath(paths []string) string {
	if len(paths) == 0 { // 경로 목록이 비어있는 경우
		return "/" // 기본값으로 루트 경로("/")를 반환
	}
	return paths[rand.Intn(len(paths))] // rand.Intn(n)은 0부터 n-1 사이의 랜덤 정수를 반환
	// 이를 인덱스로 사용해 배열에서 랜덤 항목 선택
}

// GetRandomHeaderSet 함수는 HTTP 요청에 사용할 랜덤 헤더 세트를 생성합니다.
// 입력으로 받은 헤더맵에서 각 헤더 이름마다 가능한 여러 값 중 하나를 무작위로 선택합니다.
// 매개변수:
// - headerMap map[string][]string: 헤더 이름을 키로, 가능한 값들의 배열을 값으로 가지는 맵
// 반환값:
// - map[string][]string: 각 헤더 이름마다 하나의 값만 선택된 결과 맵
func GetRandomHeaderSet(headerMap map[string][]string) map[string][]string {
	result := make(map[string][]string) // 결과를 저장할 빈 맵 생성

	for key, values := range headerMap { // headerMap의 모든 키-값 쌍을 순회
		if len(values) == 0 { // 해당 헤더에 가능한 값이 없으면
			continue // 다음 헤더로 넘어감
		}
		// 해당 헤더에 대해 가능한 값들 중 하나를 랜덤으로 선택
		v := values[rand.Intn(len(values))]
		result[key] = []string{v} // 선택된 값을 배열로 감싸서 결과 맵에 저장
		// HTTP 헤더는 여러 값을 가질 수 있어 배열 형태로 저장됨
	}
	return result // 완성된 랜덤 헤더 맵 반환
}
