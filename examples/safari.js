import { browser } from "k6/x/browser_safari";

export default async function () {
  const page = await browser.newPage();

  await page.goto("https://quickpizza.grafana.com", { 
    waitUntil: 'networkidle' 
  });

  const title = await page.title();
  console.log("Page title:", title);

  const buffer1 = await page.screenshot({ path: "example-screenshot-1.png" });
  console.log(`Screenshot saved as example-screenshot-1.png (${buffer1.length} bytes)`);

  await page.locator("visible-text=Pizza, Please!").click();

  await page.locator("//h2[text()='Our recommendation:']").waitFor();

  const buffer2 = await page.screenshot({ path: "example-screenshot-2.png" });
  console.log(`Screenshot 2 saved (${buffer2.length} bytes)`);

  await page.close();
  await browser.close();
}
