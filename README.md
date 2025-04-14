# 웹 서비스 테스트 플랫폼

## Puppeteer 크롤러 사용

### 설치
```bash
cd workers/scrappers
npm install
```

### 단독 실행
```bash
node workers/scrappers/api-extractor.js https://example.com
```

## 백엔드 서버 실행
```bash
cd backend
go run main.go
```

## API 사용 예시
```bash
# 고급 자동 테스트 (크롤링부터 GPT 분석, 부하 테스트까지 자동 수행)
curl -X POST http://localhost:8080/advanced-auto-test \
   -H "Content-Type: application/json" \
   -d '{"url": "https://example.com"}'

# 단순 부하 테스트
curl -X POST http://localhost:8080/test \
   -H "Content-Type: application/json" \
   -d '{"url": "https://example.com"}'
```

## 사용된 주요 라이브러리
- 백엔드: Go (zap 로깅)
- 크롤러: Node.js (Puppeteer)
- 분석: OpenAI GPT