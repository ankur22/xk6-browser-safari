/**
 * **Safari WebDriver extension for k6**
 *
 * @module browser_safari
 */
export as namespace browser_safari;

/**
 * Viewport dimensions
 */
export interface Viewport {
  width: number;
  height: number;
}

/**
 * Options for browser.newPage()
 */
export interface NewPageOptions {
  /**
   * Viewport dimensions (default: { width: 1280, height: 720 })
   */
  viewport?: Viewport;
}

/**
 * Browser context represents an isolated session
 */
export interface BrowserContext {
  /**
   * Create a new page in this browser context
   * @example
   * const page = await context.newPage();
   */
  newPage(): Promise<Page>;
  
  /**
   * Get all cookies for this browser context
   * @returns Promise<Cookie[]>
   */
  cookies(): Promise<Cookie[]>;
}

/**
 * Cookie object
 */
export interface Cookie {
  name: string;
  value: string;
  domain?: string;
  path?: string;
  expires?: number;
  httpOnly?: boolean;
  secure?: boolean;
  sameSite?: 'Strict' | 'Lax' | 'None';
}

/**
 * Safari browser instance
 */
export interface Browser {
  /**
   * Create a new browser context with optional configuration
   * @param options Context creation options
   * @example
   * const context = browser.newContext();
   * const context = browser.newContext({ viewport: { width: 1920, height: 1080 } });
   */
  newContext(options?: NewPageOptions): BrowserContext;
  
  /**
   * Create a new page in the browser (uses default context)
   * @param options Page creation options
   * @example
   * const page = await browser.newPage();
   * const page = await browser.newPage({ viewport: { width: 1920, height: 1080 } });
   */
  newPage(options?: NewPageOptions): Promise<Page>;
  
  /**
   * Close the browser and all its pages
   */
  close(): Promise<void>;
}

/**
 * Navigation options for page.goto()
 */
export interface GotoOptions {
  /**
   * When to consider navigation succeeded.
   * - 'load': Wait for the load event (default)
   * - 'domcontentloaded': Wait for DOMContentLoaded event
   * - 'networkidle': Wait for document.readyState === 'complete' + 500ms
   */
  waitUntil?: 'load' | 'domcontentloaded' | 'networkidle';
}

/**
 * Options for locator.waitFor()
 */
export interface WaitForOptions {
  /**
   * State to wait for
   * - 'attached': Wait for element to be present in DOM
   * - 'detached': Wait for element to not be present in DOM
   * - 'visible': Wait for element to be visible (default)
   * - 'hidden': Wait for element to be hidden
   */
  state?: 'attached' | 'detached' | 'visible' | 'hidden';
}

/**
 * Locator represents a way to find element(s) on the page at any moment
 */
export interface Locator {
  /**
   * Click on the element matched by the locator
   */
  click(): Promise<void>;
  
  /**
   * Get the number of elements matching the locator
   */
  count(): Promise<number>;
  
  /**
   * Get all elements matching the locator as an array of Locators
   */
  all(): Promise<Locator[]>;
  
  /**
   * Wait for the element to reach a specific state
   * @param options Wait options
   * @example
   * await page.locator('button').waitFor({ state: 'visible' });
   * await page.locator('div.loading').waitFor({ state: 'hidden', timeout: 5000 });
   */
  waitFor(options?: WaitForOptions): Promise<void>;

  /**
   * Get the text content of the element
   * @returns Promise that resolves to the text content
   * @example
   * const text = await locator.textContent();
   * console.log('Element text:', text);
   */
  textContent(): Promise<string>;

  /**
   * Type text into the element character by character
   * @param text Text to type
   * @param options Typing options
   * @example
   * // Type text into an input field
   * await page.locator('input[name="email"]').type('user@example.com');
   * 
   * // Type with delay between keystrokes (for realistic typing simulation)
   * await page.locator('input[name="search"]').type('search query', { delay: 100 });
   */
  type(text: string, options?: { delay?: number }): Promise<void>;
}

/**
 * Browser page instance
 */
export interface Page {
  /**
   * Navigate to a URL
   * @param url The URL to navigate to
   * @param options Navigation options
   */
  goto(url: string, options?: GotoOptions): Promise<void>;
  
  /**
   * Get the current page URL
   */
  url(): string;
  
  /**
   * Create a locator for finding element(s) on the page
   * @param selector Selector for the element(s) (supports all selector strategies)
   * @example
   * const button = page.locator('button.submit');
   * await button.click();
   * 
   * const count = await page.locator('div.item').count();
   * const allItems = await page.locator('div.item').all();
   */
  locator(selector: string): Locator;
  
  /**
   * Get the current page title
   */
  title(): Promise<string>;
  
  /**
   * Execute JavaScript in the page context
   * @param script The JavaScript code to execute
   */
  evaluate(script: string): Promise<any>;
  
  /**
   * Click an element
   * @param selector Selector for the element. Supports multiple strategies:
   *   - CSS: "button.submit" (default)
   *   - XPath: "xpath=//button[@type='submit']" or "//button"
   *   - Text: "text=Submit Form" (exact text match)
   *   - Visible Text: "visible-text=Submit" (visible elements only)
   *   - Data TestID: "data-testid=submit-button"
   *   - ARIA Label: "aria-label=Close dialog"
   *   - ARIA Role: "role=button"
   *   - ID: "id=submitBtn"
   *   - Class: "class=submit-button"
   *   - Tag: "tag=button"
   *   - Link Text: "link=Click Here" (exact link text)
   *   - Partial Link: "partial-link=Click" (partial link text)
   */
  click(selector: string): Promise<void>;
  
  /**
   * Fill an input field with text
   * @param selector Selector for the input field (see click() for supported formats)
   * @param text Text to fill in the field
   */
  fill(selector: string, text: string): Promise<void>;
  
  /**
   * Take a screenshot of the current page
   * @param options Screenshot options
   * @param options.path Optional path where to save the screenshot
   * @returns Promise that resolves to a buffer containing the screenshot (PNG format)
   * @example
   * // Save to file and get buffer
   * const buffer = await page.screenshot({ path: 'screenshot.png' });
   * console.log('Screenshot size:', buffer.length, 'bytes');
   * 
   * // Just get the buffer without saving
   * const buffer = await page.screenshot();
   */
  screenshot(options?: { path?: string }): Promise<ArrayBuffer>;
  
  /**
   * Wait for a specified amount of time
   * @param milliseconds Number of milliseconds to wait
   * @example
   * await page.waitForTimeout(1000); // Wait for 1 second
   */
  waitForTimeout(milliseconds: number): Promise<void>;
  
  /**
   * Compare two screenshots and return a similarity score
   * @param img1 First screenshot buffer
   * @param img2 Second screenshot buffer
   * @returns Similarity score between 0.0 (completely different) and 1.0 (identical)
   * @example
   * const screenshot1 = await page.screenshot();
   * await page.click('button');
   * const screenshot2 = await page.screenshot();
   * const similarity = page.compareScreenshots(screenshot1, screenshot2);
   * console.log(`Images are ${(similarity * 100).toFixed(2)}% similar`);
   */
  compareScreenshots(img1: ArrayBuffer, img2: ArrayBuffer): number;
  
  /**
   * Count the number of different pixels between two screenshots
   * @param img1 First screenshot buffer
   * @param img2 Second screenshot buffer
   * @param threshold Difference threshold per channel (0-255), default 0
   * @returns Number of pixels that differ by more than the threshold
   * @example
   * const screenshot1 = await page.screenshot();
   * const screenshot2 = await page.screenshot();
   * const diffPixels = page.countPixelDifference(screenshot1, screenshot2, 10);
   * console.log(`${diffPixels} pixels are different`);
   */
  countPixelDifference(img1: ArrayBuffer, img2: ArrayBuffer, threshold: number): number;
  
  /**
   * Close the page
   */
  close(): Promise<void>;
}

/**
 * The global browser instance
 */
export declare const browser: Browser;

/**
 * Compare two screenshots and return a similarity score
 * @param img1 First screenshot buffer
 * @param img2 Second screenshot buffer
 * @returns Similarity score between 0.0 (completely different) and 1.0 (identical)
 * @example
 * import { compareScreenshots } from "k6/x/browser_safari";
 * 
 * const screenshot1 = await page.screenshot();
 * await page.click('button');
 * const screenshot2 = await page.screenshot();
 * const similarity = compareScreenshots(screenshot1, screenshot2);
 * console.log(`Images are ${(similarity * 100).toFixed(2)}% similar`);
 */
export declare function compareScreenshots(img1: ArrayBuffer, img2: ArrayBuffer): number;

/**
 * Create a visual diff image highlighting differences between two screenshots
 * Identical pixels are shown in grayscale, different pixels are highlighted in red
 * @param img1 First screenshot buffer
 * @param img2 Second screenshot buffer
 * @param filePath Optional path to save the diff image (e.g., "diff.png")
 * @returns The diff image as an ArrayBuffer
 * @example
 * import { createDiffImage } from "k6/x/browser_safari";
 * 
 * const screenshot1 = await page.screenshot();
 * const screenshot2 = await page.screenshot();
 * 
 * // Create diff and save to file
 * const diffImage = createDiffImage(screenshot1, screenshot2, "diff.png");
 * console.log(`Diff image saved to diff.png (${diffImage.length} bytes)`);
 * 
 * // Or just get the buffer without saving
 * const diffImage = createDiffImage(screenshot1, screenshot2, "");
 */
export declare function createDiffImage(img1: ArrayBuffer, img2: ArrayBuffer, filePath: string): ArrayBuffer;