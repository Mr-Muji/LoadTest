// api-extractor.js
const puppeteer = require('puppeteer');

/**
 * 대상 URL에서 API 요청 경로를 추출하는 함수
 * @param {string} targetUrl - 분석할 웹사이트 URL
 * @returns {Promise<Array>} - 추출된 API 경로 배열
 */
async function extractApiEndpoints(targetUrl) {
  const apiEndpoints = new Set();
  
  // 더 많은 옵션으로 브라우저 시작
  const browser = await puppeteer.launch({
    headless: "new",
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-web-security'],
    ignoreHTTPSErrors: true
  });
  
  try {
    const page = await browser.newPage();
    
    // 모든 요청 (XHR, Fetch 등)을 감시하는 방식으로 변경
    // 요청 인터셉션 비활성화 (안정성 향상)
    page.on('request', (request) => {
      try {
        const url = new URL(request.url());
        const path = url.pathname;
        
        // 디버깅 로그 추가
        console.log(`요청: ${request.method()} ${path}`);
        
        // 필터링 기준 완화: 더 다양한 패턴 허용 (확장자 없는 경로 포함)
        if (
          // 일반적인 API 패턴
          path.includes('/api/') || 
          path.match(/\/v\d+\//) || 
          path.includes('/graphql') ||
          path.includes('/rest/') ||
          path.includes('/service/') ||
          // .js, .css, .jpg 등 리소스 파일 제외
          (!path.match(/\.(js|css|png|jpg|jpeg|gif|svg|ico|woff|woff2|ttf|eot)$/i) && 
           path !== '/' && 
           path.length > 1)
        ) {
          apiEndpoints.add(`${request.method()} ${path}`);
        }
      } catch (error) {
        console.log(`요청 처리 오류: ${error.message}`);
      }
    });
    
    // 네트워크 응답도 모니터링
    page.on('response', async (response) => {
      try {
        const url = new URL(response.url());
        const contentType = response.headers()['content-type'] || '';
        
        // JSON 응답은 API일 가능성이 높음
        if (contentType.includes('application/json')) {
          apiEndpoints.add(`${response.request().method()} ${url.pathname}`);
        }
      } catch (e) {}
    });
    
    // 타임아웃 증가
    await page.setDefaultNavigationTimeout(60000);
    
    console.log(`🔍 사이트 분석 중: ${targetUrl}`);
    
    // 페이지 로딩 및 추가 대기 시간
    await page.goto(targetUrl, { waitUntil: 'networkidle2' });
    await new Promise(resolve => setTimeout(resolve, 5000));
    
    // JavaScript 추적을 통한 API 호출 캡처
    await page.evaluate(() => {
      // 원본 fetch 함수 저장
      const originalFetch = window.fetch;
      
      // fetch 함수 오버라이드
      window.fetch = function(...args) {
        try {
          const url = args[0];
          console.log('Fetch 호출:', url);
        } catch (e) {}
        
        return originalFetch.apply(this, args);
      };
      
      // XHR 요청 가로채기
      const originalXHROpen = XMLHttpRequest.prototype.open;
      XMLHttpRequest.prototype.open = function(...args) {
        try {
          const method = args[0];
          const url = args[1];
          console.log('XHR 호출:', method, url);
        } catch (e) {}
        
        return originalXHROpen.apply(this, args);
      };
    });
    
    console.log('💡 페이지 상호작용 시뮬레이션...');
    
    // 더 안정적인 상호작용
    await simulateUserInteraction(page);
    
    // 결과 반환
    return Array.from(apiEndpoints);
  } finally {
    await browser.close();
  }
}

async function simulateUserInteraction(page) {
  try {
    // 더 안전한 상호작용 방식
    await page.evaluate(async () => {
      // 클릭 가능한 요소 찾기
      const buttons = Array.from(document.querySelectorAll('button, a, [role="button"], .btn'));
      
      // 각 요소에 대해 시도
      for (let i = 0; i < Math.min(buttons.length, 8); i++) {
        try {
          const button = buttons[i];
          // 화면에 보이는 요소만 클릭
          if (button.offsetParent !== null) {
            console.log('클릭:', button.textContent || button.innerText);
            button.click();
            // 잠시 대기
            await new Promise(r => setTimeout(r, 500));
          }
        } catch (e) {}
      }
      
      // 스크롤
      window.scrollTo(0, document.body.scrollHeight / 2);
      await new Promise(r => setTimeout(r, 1000));
      window.scrollTo(0, document.body.scrollHeight);
    });
    
    // 추가 대기
    await new Promise(resolve => setTimeout(resolve, 3000));
    
  } catch (error) {
    console.log('⚠️ 페이지 상호작용 중 오류:', error.message);
  }
}

// 명령줄에서 실행 시 사용 예제
async function main() {
  if (process.argv.length < 3) {
    console.log('사용법: node api-extractor.js https://example.com');
    process.exit(1);
  }
  
  const targetUrl = process.argv[2];
  try {
    const endpoints = await extractApiEndpoints(targetUrl);
    
    console.log('\n🎯 발견된 API 엔드포인트:');
    if (endpoints.length > 0) {
      endpoints.forEach(endpoint => console.log(`- ${endpoint}`));
      
      // GET 요청 경로만 필터링하여 부하테스트용으로 준비
      const pathList = endpoints
        .filter(endpoint => endpoint.startsWith('GET'))
        .map(endpoint => endpoint.substring(4)); // "GET " 접두사 제거
      
      // JSON 형식으로 저장
      const result = {
        target: targetUrl,
        pathList: pathList,
        rps: 10,
        duration: 10,
        method: "GET"
      };
      
      console.log('\n✅ 부하테스트 구성:');
      console.log(JSON.stringify(result, null, 2));
    } else {
      console.log('API 엔드포인트를 찾지 못했습니다.');
    }
  } catch (error) {
    console.error('🚨 오류 발생:', error);
  }
}

// 스크립트 직접 실행 시
if (require.main === module) {
  main();
}

module.exports = { extractApiEndpoints };