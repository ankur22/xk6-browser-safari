# xk6-browser-safari

**Safari WebDriver extension for k6**

This k6 extension provides Safari browser automation capabilities using the WebDriver protocol. It allows you to control Safari browser instances, navigate to web pages, interact with elements, and execute JavaScript within k6 test scripts.

## Prerequisites

Before using this extension, you need to:

1. **Enable Safari WebDriver**: 
   - Open Safari
   - Go to Safari > Preferences > Advanced
   - Check "Show Develop menu in menu bar"
   - Go to Develop > Allow Remote Automation

**Note:** The extension automatically starts `safaridriver --port 4444` when you call `launch()` and stops it when the browser is closed. You don't need to manually start safaridriver.

## Features

### Automatic Script Injection

Similar to Playwright, this extension automatically injects a JavaScript helper script (`injection_script.js`) into every page when:
- A new page is created with `browser.newPage()`
- Navigation occurs with `page.goto()`

The injection script provides:
- **Helper utilities** accessible via `window.__webdriverHelpers`
- **Element information** retrieval
- **Visibility detection** for elements
- **Wait utilities** for selectors
- **Page metrics** collection

You can access these helpers in your `page.evaluate()` calls:
```javascript
const isVisible = await page.evaluate(
  "window.__webdriverHelpers.isVisible(document.querySelector('button'))"
);
```

### Advanced Selector Support

This extension supports multiple selector strategies beyond basic CSS selectors:

#### Native WebDriver Selectors (Fast)
```javascript
// CSS Selector (default)
await page.click("button.submit");

// XPath
await page.click("xpath=//button[@type='submit']");
await page.click("//div[@class='container']//button");

// ID
await page.click("id=submitBtn");

// Class Name
await page.click("class=submit-button");

// Tag Name
await page.click("tag=button");

// Link Text (exact match)
await page.click("link=Click Here");

// Partial Link Text
await page.click("partial-link=Click");
```

#### Custom JavaScript Selectors (Powerful)
```javascript
// Text Content (exact match)
await page.click("text=Submit Form");

// Visible Text (only visible elements)
await page.click("visible-text=Submit");

// Data Test ID
await page.click("data-testid=submit-button");

// ARIA Label
await page.click("aria-label=Close dialog");

// ARIA Role
await page.click("role=button");
```

The extension automatically detects the selector type and uses the optimal strategy. See `examples/selectors.js` for more examples.

## Usage

```javascript file=script.js
import { browser } from "k6/x/browser_safari";

export default async function () {
  // Create a new page
  const page = await browser.newPage();
  
  // Navigate to a website and wait for network to be idle
  await page.goto("https://example.com", { waitUntil: 'networkidle' });
  
  // Get page information
  console.log("Current URL:", page.url());
  const title = await page.title();
  console.log("Page title:", title);
  
  // Execute JavaScript
  const result = await page.evaluate("document.body.innerHTML.length");
  console.log("Page content length:", result);
  
  // Interact with elements
  await page.click("a[href]");
  await page.fill("input[name='search']", "test query");
  
  // Take a screenshot
  await page.screenshot({ path: "my-screenshot.png" });
  
  // Close the page
  await page.close();
  
  // Close the browser
  await browser.close();
}
```

## Locator API

The Locator API provides a Playwright-style way to find and interact with elements. Locators are created synchronously but resolve elements lazily when actions are performed.

### Basic Usage

```javascript
// Create a locator (doesn't find the element yet)
const button = page.locator('button.submit');

// Perform action (finds element at this moment, then clicks)
await button.click();
```

### Locator Methods

#### `page.locator(selector)`
Creates a locator for finding element(s) on the page.

**Parameters:**
- `selector` (string): Selector for the element(s) (supports all selector strategies)

**Returns:** `Locator`

**Example:**
```javascript
const button = page.locator('button.primary');
const items = page.locator('div.item');
```

#### `locator.click()`
Clicks on the element matched by the locator.

**Returns:** `Promise<void>`

**Example:**
```javascript
await page.locator('button.submit').click();
```

#### `locator.count()`
Returns the number of elements matching the locator.

**Returns:** `Promise<number>`

**Example:**
```javascript
const count = await page.locator('div.item').count();
console.log('Found', count, 'items');
```

#### `locator.all()`
Returns all elements matching the locator as an array of Locators (each representing a specific element).

**Returns:** `Promise<Locator[]>`

**Example:**
```javascript
const buttons = await page.locator('button').all();
for (let i = 0; i < buttons.length; i++) {
  await buttons[i].click();
}
```

#### `locator.waitFor(options?)`
Waits for the element to reach a specific state (fixed 30 second timeout).

**Parameters:**
- `options` (object, optional):
  - `state` (string): State to wait for - `'attached'`, `'detached'`, `'visible'` (default), or `'hidden'`

**Returns:** `Promise<void>`

**Example:**
```javascript
// Wait for button to be visible (default)
await page.locator('button.submit').waitFor();

// Wait for button to be visible (explicit)
await page.locator('button.submit').waitFor({ state: 'visible' });

// Wait for loading spinner to disappear
await page.locator('div.loading').waitFor({ state: 'hidden' });

// Wait for element to be attached to DOM
await page.locator('div.new-content').waitFor({ state: 'attached' });

// Wait for element to be removed from DOM
await page.locator('div.old-content').waitFor({ state: 'detached' });
```

#### `locator.textContent()`
Returns the text content of the element.

**Returns:** `Promise<string>` - A promise that resolves to the text content

**Example:**
```javascript
// Get text from a heading
const heading = page.locator('h1');
const text = await heading.textContent();
console.log('Heading:', text);

// Get text from multiple elements
const items = await page.locator('li.item').all();
for (const item of items) {
  const text = await item.textContent();
  console.log('Item:', text);
}

// Use with custom selectors
const buttonText = await page.locator('visible-text=Click me').textContent();
console.log('Button says:', buttonText);
```

#### `locator.type(text, options?)`
Types text into the element character by character. Similar to `page.fill()` but uses the WebDriver SendKeys command.

**Parameters:**
- `text` (string): Text to type into the element
- `options` (object, optional): Typing options
  - `delay` (number): Delay in milliseconds between keystrokes (note: acknowledged but not fully implemented due to WebDriver limitations)

**Returns:** `Promise<void>` - A promise that resolves when typing is complete

**Example:**
```javascript
// Type into an input field
await page.locator('input[name="email"]').type('user@example.com');

// Type into a textarea
await page.locator('textarea#message').type('Hello, world!');

// Type with delay option (for future use)
await page.locator('input[name="search"]').type('search query', { delay: 100 });

// Type after finding the element
const searchInput = page.locator('input[type="search"]');
await searchInput.type('k6 testing');

// Use with custom selectors
await page.locator('data-testid=username-input').type('testuser');
```

**Note:** This method uses WebDriver's SendKeys command. The `delay` option is accepted but not currently implemented due to WebDriver's native limitations.

### Why Use Locators?

1. **Auto-waiting**: Locators find elements at action time, making tests more reliable
2. **Reusable**: Create once, use multiple times
3. **Composable**: Works with all selector strategies
4. **Playwright-compatible**: Familiar API for Playwright users

### Example

```javascript
import { browser } from "k6/x/browser_safari";

export default async function () {
  const page = await browser.newPage();
  
  await page.goto("https://example.com");
  
  // Create locators
  const items = page.locator('div.item');
  
  // Check count
  const count = await items.count();
  console.log('Found', count, 'items');
  
  // Click all items
  const allItems = await items.all();
  for (let i = 0; i < allItems.length; i++) {
    await allItems[i].click();
  }
  
  await page.close();
  await browser.close();
}
```

## API Reference

### Browser

The `Browser` interface provides methods to control a Safari browser instance.

#### `browser.newContext(options?)`
Creates a new browser context with optional configuration (conceptual layer for grouping pages).

**Parameters:**
- `options` (object, optional):
  - `viewport` (object): Viewport dimensions
    - `width` (number): Viewport width in pixels (default: 1280)
    - `height` (number): Viewport height in pixels (default: 720)

**Returns:** `BrowserContext`

**Example:**
```javascript
const context = browser.newContext();
const context = browser.newContext({ viewport: { width: 1920, height: 1080 } });
const page = await context.newPage();
```

**Note:** Since WebDriver doesn't have a native context concept, this is a logical grouping. All pages still share the same WebDriver session state.

#### `browser.newPage(options?)`
Creates a new page (tab) in the browser with optional viewport configuration.

**Parameters:**
- `options` (object, optional):
  - `viewport` (object): Viewport dimensions
    - `width` (number): Viewport width in pixels (default: 1280)
    - `height` (number): Viewport height in pixels (default: 720)

**Returns:** `Promise<Page>`

**Example:**
```javascript
// Default viewport (1280x720)
const page = await browser.newPage();

// Custom viewport
const page = await browser.newPage({ 
  viewport: { width: 1920, height: 1080 } 
});

// Mobile viewport
const page = await browser.newPage({ 
  viewport: { width: 375, height: 667 } 
});
```

#### `browser.close()`
Closes the browser and all its pages.

**Returns:** `Promise<void>` - A promise that resolves when the browser is closed

### BrowserContext

The `BrowserContext` interface provides methods to manage an isolated browser context.

#### `context.newPage()`
Creates a new page in this browser context. Uses the viewport settings from the context.

**Returns:** `Promise<Page>`

**Example:**
```javascript
const context = browser.newContext({ viewport: { width: 1920, height: 1080 } });
const page = await context.newPage();
```

#### `context.cookies()`
Returns all cookies for this browser context from the active WebDriver session.

**Returns:** `Promise<Cookie[]>` - A promise that resolves to an array of cookies

**Example:**
```javascript
const context = browser.newContext();
const page = await context.newPage();
await page.goto("https://example.com");

// Get cookies after navigation
const cookies = await context.cookies();
console.log("Number of cookies:", cookies.length);
cookies.forEach(cookie => {
  console.log(`${cookie.name}: ${cookie.value}`);
});
```

**Note:** Requires at least one page to be created in the context. If no session is active, this will return an error.

### Page

The `Page` interface provides methods to interact with a web page.

#### `page.goto(url, options?)`
Navigates to the specified URL with optional wait conditions.

**Parameters:**
- `url` (string): The URL to navigate to
- `options` (object, optional): Navigation options
  - `waitUntil` (string, optional): When to consider navigation succeeded. One of:
    - `'load'` - Wait for the load event (default)
    - `'domcontentloaded'` - Wait for DOMContentLoaded event  
    - `'networkidle'` - Wait for document.readyState === 'complete' + 500ms (simplified network idle detection)

**Returns:** `Promise<void>` - A promise that resolves when navigation is complete

**Example:**
```javascript
// Wait for load event (default)
await page.goto("https://example.com");

// Wait for DOM to be ready
await page.goto("https://example.com", { waitUntil: 'domcontentloaded' });

// Wait for network to be idle
await page.goto("https://example.com", { waitUntil: 'networkidle' });
```

#### `page.url()`
Gets the current page URL.

**Returns:** `string` - The current URL

#### `page.title()`
Gets the current page title.

**Returns:** `Promise<string>` - A promise that resolves to the page title

#### `page.evaluate(script)`
Executes JavaScript in the page context.

**Parameters:**
- `script` (string): The JavaScript code to execute

**Returns:** `Promise<any>` - A promise that resolves to the result of the script execution

#### `page.click(selector)`
Clicks an element by CSS selector.

**Parameters:**
- `selector` (string): CSS selector for the element to click

**Returns:** `Promise<void>` - A promise that resolves when the click is complete

#### `page.fill(selector, text)`
Fills an input field with text.

**Parameters:**
- `selector` (string): CSS selector for the input field
- `text` (string): Text to fill in the field

**Returns:** `Promise<void>` - A promise that resolves when the text is filled

#### `page.screenshot(options?)`
Takes a screenshot of the current page and returns the image data as a buffer.

**Parameters:**
- `options` (object, optional): Screenshot options
  - `path` (string, optional): Path where to save the screenshot file

**Returns:** `Promise<ArrayBuffer>` - A promise that resolves to a buffer containing the PNG screenshot data

**Example:**
```javascript
// Save to file and get the buffer
const buffer = await page.screenshot({ path: "example.png" });
console.log(`Screenshot saved, size: ${buffer.length} bytes`);

// Get buffer without saving to file
const buffer = await page.screenshot();
// You can then process the buffer as needed

// Multiple screenshots
const screenshot1 = await page.screenshot({ path: "screenshot-1.png" });
const screenshot2 = await page.screenshot({ path: "screenshot-2.png" });
```

**Note:** Like Playwright, this method always returns the screenshot buffer regardless of whether a path is provided. This allows you to both save the screenshot and process the image data.

#### `page.waitForTimeout(milliseconds)`
Waits for the specified number of milliseconds. Useful for adding delays in test scripts.

**Parameters:**
- `milliseconds` (number): Number of milliseconds to wait

**Returns:** `Promise<void>` - A promise that resolves after the timeout

**Example:**
```javascript
await page.waitForTimeout(1000); // Wait for 1 second
await page.waitForTimeout(500);  // Wait for 500ms
```

**Note:** For waiting for elements or navigation, prefer using `locator.waitFor()` or checking for specific elements rather than fixed timeouts.

#### `page.close()`
Closes the page.

**Returns:** `Promise<void>` - A promise that resolves when the page is closed

## Quick start

1. **Build the extension**:
    ```shell
   xk6 build --with xk6-browser-safari=.
    ```

2. **Run the test script**:
    ```shell
   ./k6 run script.js
    ```

## Development environment

While using a GitHub codespace in the browser is a good starting point, you can also set up a local development environment for a better developer experience.

To create a local development environment, you need an IDE that supports [Development Containers](https://containers.dev/). [Visual Studio Code](https://code.visualstudio.com/) supports Development Containers after installing the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers).

1. First, clone the `xk6-quickstart` repository to your machine and open it in Visual Studio Code. Make sure to replace `USER` with your GitHub username:

   ```shell
   git clone https://github.com/USER/xk6-quickstart.git
   code xk6-quickstart
   ```

2. Visual Studio Code will detect the [development container](https://containers.dev/) configuration and show a pop-up to open the project in a dev container. Accept the prompt and the project opens in the dev container, and the container image is rebuilt if necessary.

3. Run the test script. The repository's root directory includes a `script.js` file. When developing k6 extensions, use the `xk6 run` command instead of `k6 run` to execute your scripts.

    ```shell
    xk6 run script.js
    ```

## Download

Building a custom k6 binary with the `xk6-browser-safari` extension is necessary for its use. You can download pre-built k6 binaries from the [Releases page](https://xk6-browser-safari/releases/).

## Build

Use the [xk6](https://github.com/grafana/xk6) tool to build a custom k6 binary with the `xk6-browser-safari` extension. Refer to the [xk6 documentation](https://github.com/grafana/xk6) for more information.

## Contribute

If you wish to contribute to this project, please start by reading the [Contributing Guidelines](CONTRIBUTING.md).
