const fs = require('fs');
const { chromium } = require('playwright');
const systemChrome = process.env.CHROME_PATH || ['C:/Program Files/Google/Chrome/Application/chrome.exe','C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(fs.existsSync);
(async()=>{
 const browser=await chromium.launch({headless:true,...(systemChrome?{executablePath:systemChrome}:{})});
 try{
  const page=await browser.newPage({viewport:{width:1440,height:900}});
  await page.goto('http://127.0.0.1:8090/');
  const title=await page.locator('h1').textContent();
  await page.screenshot({path:'artifacts/frontend-login-desktop.png',fullPage:true});
  await page.locator('[data-auth="register"]').click();
  const registerNameVisible=await page.locator('#nameField').isVisible();
  await page.evaluate(()=>{localStorage.setItem('issuepm_token','preview-token');localStorage.setItem('issuepm_user',JSON.stringify({id:'preview',name:'Preview User',role:'user'}))});
  await page.reload();
  await page.locator('#themeBtn').click();
  const theme=await page.locator('html').getAttribute('data-theme');
  await page.screenshot({path:'artifacts/frontend-dashboard-dark.png',fullPage:true});
  await page.evaluate(()=>{localStorage.removeItem('issuepm_token');localStorage.removeItem('issuepm_user')});
  await page.setViewportSize({width:390,height:844});
  await page.reload();
  await page.screenshot({path:'artifacts/frontend-login-mobile.png',fullPage:true});
  console.log(JSON.stringify({title,registerNameVisible,theme}));
 } finally {await browser.close()}
})();

