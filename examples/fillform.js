import { browser } from "k6/browser";
import { browser as browser_safari, compareScreenshots, createDiffImage } from "k6/x/browser_safari";
import { check } from 'https://jslib.k6.io/k6-utils/1.5.0/index.js';
import { Trend } from 'k6/metrics';
import redis from 'k6/experimental/redis';
import encoding from 'k6/encoding';

// Initialize Redis client
const redisClient = new redis.Client('redis://localhost:6379');

// Create custom metrics for screenshot similarity
const screenshotSimilarity1 = new Trend('screenshot_similarity_1', true);
const screenshotSimilarity2 = new Trend('screenshot_similarity_2', true);
const screenshotSimilarity3 = new Trend('screenshot_similarity_3', true);
const screenshotSimilarity4 = new Trend('screenshot_similarity_4', true);

export const options = {
  scenarios: {
    safari: {
      exec: 'safari',
      executor: 'shared-iterations',
    },
    chromium: {
      exec: 'chromium',
      executor: 'shared-iterations',
      options: {
        browser: {
          type: 'chromium',
        },
      },
    },
    compare: {
      exec: 'compare',
      executor: 'shared-iterations',
    },
  },
  thresholds: {
    checks: ["rate==1.0"],
    // Screenshot similarity thresholds
    // Values are between 0.0 and 1.0, where 1.0 means 100% identical
    // We expect at least 95% similarity (0.95) for all screenshots
    'screenshot_similarity_1': ['avg>=0.95'],
    'screenshot_similarity_2': ['avg>=0.95'],
    'screenshot_similarity_3': ['avg>=0.95'],
    'screenshot_similarity_4': ['avg>=0.95'],
  }
}

export function teardown(data) {
  // Delete all screenshot keys
  const screenshotNumbers = [1, 2, 3, 4];
  
  for (const num of screenshotNumbers) {
    try {
      redisClient.del(`screenshot:safari:${num}`);
    } catch (e) {
      console.log(`⚠ Failed to delete screenshot:safari:${num}: ${e}`);
    }
    
    try {
      redisClient.del(`screenshot:chromium:${num}`);
    } catch (e) {
      console.log(`⚠ Failed to delete screenshot:chromium:${num}: ${e}`);
    }
  }
}

export async function compare() {
  try {
    // Helper function to wait for a key to exist in Redis
    async function waitForKey(key, maxAttempts = 30, delayMs = 1000) {
      for (let i = 0; i < maxAttempts; i++) {
        const exists = await redisClient.exists(key);
        if (exists) {
          return true;
        }
        await new Promise(resolve => setTimeout(resolve, delayMs));
      }
      throw new Error(`Timeout waiting for ${key} after ${maxAttempts * delayMs / 1000} seconds`);
    }

    const screenshots = [1, 2, 3, 4];
    const results = [];

    // Wait for all screenshots to be available
    for (const num of screenshots) {
      await waitForKey(`screenshot:safari:${num}`);
      await waitForKey(`screenshot:chromium:${num}`);
    }

    // Map screenshot numbers to their metric objects
    const similarityMetrics = {
      1: screenshotSimilarity1,
      2: screenshotSimilarity2,
      3: screenshotSimilarity3,
      4: screenshotSimilarity4,
    };

    // Compare each screenshot pair
    for (const num of screenshots) {
      // Retrieve screenshots from Redis
      const safariScreenshotBase64 = await redisClient.get(`screenshot:safari:${num}`);
      const chromiumScreenshotBase64 = await redisClient.get(`screenshot:chromium:${num}`);

      // Decode base64 to binary
      const safariScreenshot = encoding.b64decode(safariScreenshotBase64);
      const chromiumScreenshot = encoding.b64decode(chromiumScreenshotBase64);

      // Compare screenshots
      const similarity = compareScreenshots(safariScreenshot, chromiumScreenshot);

      // Record similarity metric
      similarityMetrics[num].add(similarity);

      // Create check for this screenshot
      check(similarity, {
        [`screenshot ${num} similarity >= 95%`]: (s) => s >= 0.95,
      });

      // Create diff image if there are differences
      if (similarity < 1.0) {
        createDiffImage(safariScreenshot, chromiumScreenshot, `screenshots/fillform-diff-${num}.png`);
      }
    }

  } catch (error) {
    console.error('Comparison failed:', error);
    throw error;
  }
}

export async function safari() {
  const context = await browser_safari.newContext({
    viewport: {
      width: 1280,
      height: 720
    }
  });
  const page = await context.newPage();

  try {
    await page.goto('https://quickpizza.grafana.com/test.k6.io/', { waitUntil: 'networkidle' });
    
    const screenshot1 = await page.screenshot({ path: "screenshots/fillform-safari-screenshot-1.png" });
    const screenshot1Base64 = encoding.b64encode(screenshot1);
    await redisClient.set('screenshot:safari:1', screenshot1Base64);

    await page.locator('a[href="/my_messages.php"]').click();

    await page.locator("//h2[text()='Login']").waitFor();

    const screenshot2 = await page.screenshot({ path: "screenshots/fillform-safari-screenshot-2.png" });
    const screenshot2Base64 = encoding.b64encode(screenshot2);
    await redisClient.set('screenshot:safari:2', screenshot2Base64);

    await page.locator('input[name="login"]').type('admin');
    await page.locator('input[name="password"]').type("123");

    const screenshot3 = await page.screenshot({ path: "screenshots/fillform-safari-screenshot-3.png" });
    const screenshot3Base64 = encoding.b64encode(screenshot3);
    await redisClient.set('screenshot:safari:3', screenshot3Base64);

    await page.locator('input[type="submit"]').click();

    await page.locator("//h2[text()='Welcome, admin!']").waitFor();

    await check(page.locator('h2'), {
      'header': async lo => {
        return await lo.textContent() == 'Welcome, admin!'
      }
    });

    // Check whether we receive cookies from the logged site.
    await check(context, {
      'session cookie is set': async ctx => {
        const cookies = await ctx.cookies();
        return cookies.find(c => c.name == 'AWSALB') !== undefined;
      }
    });

    const screenshot4 = await page.screenshot({ path: "screenshots/fillform-safari-screenshot-4.png" });
    const screenshot4Base64 = encoding.b64encode(screenshot4);
    await redisClient.set('screenshot:safari:4', screenshot4Base64);
  } finally {
    await page.close();
    await browser_safari.close();
  }
}

export async function chromium() {
  const context = await browser.newContext({
    viewport: {
      width: 1280,
      height: 720
    }
  });
  const page = await context.newPage();

  try {
    await page.goto('https://quickpizza.grafana.com/test.k6.io/', { waitUntil: 'networkidle' });
    
    const screenshot1 = await page.screenshot({ path: "screenshots/fillform-chromium-screenshot-1.png" });
    const screenshot1Base64 = encoding.b64encode(screenshot1);
    await redisClient.set('screenshot:chromium:1', screenshot1Base64);

    await page.locator('a[href="/my_messages.php"]').click();

    await page.locator("//h2[text()='Login']").waitFor();

    const screenshot2 = await page.screenshot({ path: "screenshots/fillform-chromium-screenshot-2.png" });
    const screenshot2Base64 = encoding.b64encode(screenshot2);
    await redisClient.set('screenshot:chromium:2', screenshot2Base64);

    await page.locator('input[name="login"]').type('admin');
    await page.locator('input[name="password"]').type("123");

    const screenshot3 = await page.screenshot({ path: "screenshots/fillform-chromium-screenshot-3.png" });
    const screenshot3Base64 = encoding.b64encode(screenshot3);
    await redisClient.set('screenshot:chromium:3', screenshot3Base64);

    await page.locator('input[type="submit"]').click();

    await page.locator("//h2[text()='Welcome, admin!']").waitFor();

    await check(page.locator('h2'), {
      'header': async lo => {
        return await lo.textContent() == 'Welcome, admin!'
      }
    });

    // Check whether we receive cookies from the logged site.
    await check(context, {
      'session cookie is set': async ctx => {
        const cookies = await ctx.cookies();
        return cookies.find(c => c.name == 'AWSALB') !== undefined;
      }
    });

    const screenshot4 = await page.screenshot({ path: "screenshots/fillform-chromium-screenshot-4.png" });
    const screenshot4Base64 = encoding.b64encode(screenshot4);
    await redisClient.set('screenshot:chromium:4', screenshot4Base64);
  } finally {
    await page.close();
    await context.close();
  }
}

