import { chromium } from 'playwright';

const appURL = process.env.APP_URL || 'http://127.0.0.1:5173';
const screenshotPath = process.env.HIIT_SCREENSHOT || '';
const browser = await chromium.launch({ headless: true });
const page = await browser.newPage({ viewport: { width: 1280, height: 800 } });
const errors = [];
page.on('pageerror', error => errors.push(error.message));

try {
    await page.goto(appURL, { waitUntil: 'networkidle' });
    await page.getByText('HIIT Timer', { exact: true }).click();
    await page.waitForSelector('.hiit-orb');
    if (screenshotPath) await page.screenshot({ path: screenshotPath, fullPage: true });

    await page.locator('#hiitSound').uncheck();
    await page.locator('#hiitWork').fill('1');
    await page.locator('#hiitWork').press('Tab');
    await page.locator('#hiitRest').fill('1');
    await page.locator('#hiitRest').press('Tab');
    await page.locator('#hiitRounds').fill('2');
    await page.locator('#hiitRounds').press('Tab');

    await page.locator('#hiitStartPause').click();
    await page.waitForTimeout(250);
    await page.locator('#hiitStartPause').click();
    const pausedValue = await page.locator('#hiitTime').innerText();
    await page.waitForTimeout(1100);
    if (await page.locator('#hiitTime').innerText() !== pausedValue) throw new Error('paused timer continued counting');

    await page.locator('#hiitStartPause').click();
    await page.waitForFunction(() => document.querySelector('#hiitStage')?.dataset.phase === 'rest', null, { timeout: 2500 });
    await page.waitForFunction(() => document.querySelector('#hiitStage')?.dataset.phase === 'work' && document.querySelector('#hiitRound')?.textContent?.includes('2 / 2'), null, { timeout: 2500 });
    await page.waitForFunction(() => document.querySelector('#hiitStage')?.dataset.phase === 'complete', null, { timeout: 2500 });

    if (errors.length) throw new Error(errors.join('\n'));
    console.log('HIIT timer E2E passed');
} finally {
    await browser.close();
}
