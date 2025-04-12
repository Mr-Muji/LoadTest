Puppeteer 사용을 위한 설치
# 필요한 패키지 설치
npm init -y
npm install puppeteer

# 실행
node api-extractor.js https://example.com

# 로깅툴
zap 사용
go get -u go.uber.org/zap

ai 입력
curl -X POST http://localhost:8080/advanced-auto-test \
   -H "Content-Type: application/json" \
   -d '{"url": "https://kakaotech.my"}'