package tester

import (
	"math/rand"
	"time"
)

// init 함수는 프로그램 시작 시 랜덤 시드를 설정한다.
// 이를 통해 매번 실행 시 다른 랜덤 값이 나오게 한다.
func init() {
	rand.Seed(time.Now().UnixNano())
}

// GetRandomPath는 전달받은 경로 리스트 중 하나를 무작위로 선택한다.
// 만약 리스트가 비어 있으면 기본 경로 "/"를 반환한다.
func GetRandomPath(paths []string) string {
	if len(paths) == 0 {
		return "/"
	}
	return paths[rand.Intn(len(paths))]
}

// GetRandomHeaderSet은 헤더 이름에 대해 주어진 후보 값 중 하나씩 골라서
// 완성된 헤더 세트를 반환한다.
// 예: { "User-Agent": ["A", "B"], "Accept": ["C", "D"] } → {"User-Agent": "A", "Accept": "D"}
func GetRandomHeaderSet(headerMap map[string][]string) map[string][]string {
	result := make(map[string][]string)

	for key, values := range headerMap {
		if len(values) == 0 {
			continue
		}
		// key마다 하나씩 랜덤으로 선택
		v := values[rand.Intn(len(values))]
		result[key] = []string{v}
	}
	return result
}
