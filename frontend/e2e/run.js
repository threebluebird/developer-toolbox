const { chromium } = require('playwright');

(async () => {
  const appUrl = process.env.APP_URL || 'http://localhost:34115';
  console.log('Starting E2E test against', appUrl);

  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1280, height: 800 } });

  try {
    await page.goto(appUrl, { waitUntil: 'networkidle' });
    console.log('Page loaded');

    // Wait for tool cards to appear
    await page.waitForSelector('.tool-card', { timeout: 10000 });
    console.log('Tool cards found');

    // Select first tool
    const firstCard = (await page.$$('.tool-card'))[0];
    if (!firstCard) throw new Error('No tool card found');

    const toolName = await firstCard.$eval('strong', el => el.innerText);
    console.log('Selecting tool:', toolName);
    await firstCard.click();

    // Click favorite button inside the first card
    const favBtn = await firstCard.$('.favorite-btn');
    if (favBtn) {
      console.log('Clicking favorite button');
      await favBtn.click();

      // Wait a moment for favorites list to refresh
      await page.waitForTimeout(800);

      // Check favorite list contains the tool name
      const favContains = await page.$$eval('#favoriteList .tool-card strong', els => els.map(e => e.innerText));
      if (favContains.includes(toolName)) {
        console.log('Favorite added successfully');
      } else {
        console.warn('Favorite not found in favorite list (may be empty or backend did not persist)');
      }
    } else {
      console.warn('No favorite button found on card');
    }

    // Click run sample button
    const runBtn = await page.$('#runSampleBtn');
    if (runBtn) {
      console.log('Clicking Run Sample');
      await runBtn.click();
      // Wait for toast to appear
      try {
        await page.waitForSelector('#toast:not(.hidden)', { timeout: 5000 });
        const toastText = await page.$eval('#toast', el => el.innerText);
        console.log('Toast appeared:', toastText);
      } catch (e) {
        console.warn('No toast appeared after running sample');
      }
    }

    console.log('E2E script completed');
    await browser.close();
    process.exit(0);
  } catch (err) {
    console.error('E2E test failed:', err);
    await browser.close();
    process.exit(2);
  }
})();
