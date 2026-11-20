import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { chromium } from "playwright";

const root = new URL("../", import.meta.url);
const browser = await chromium.launch({ executablePath: process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE });
try {
  const page = await browser.newPage();
  await page.route("http://font-test.local/**", async (route) => {
    const path = new URL(route.request().url()).pathname;
    if (path === "/") {
      await route.fulfill({ contentType: "text/html", body: `<!doctype html>
        <link rel="stylesheet" href="/src/styles/tokens.css">
        <link rel="stylesheet" href="/src/styles/studio.css">
        <div class="flow"><div class="call" data-k="say"><div class="c"><div class="out">
          <div class="txt"><p class="md" id="sample"></p></div>
        </div></div></div></div>` });
      return;
    }
    assert.match(path, /^\/(src\/styles\/(tokens|studio)\.css|fonts\/[\w.-]+)$/);
    await route.fulfill({
      contentType: path.endsWith(".css") ? "text/css" : "application/octet-stream",
      body: await readFile(new URL(path.startsWith("/fonts/") ? `public${path}` : path.slice(1), root)),
    });
  });
  for (const pass of ["initial load", "reload"]) {
    await page.goto("http://font-test.local/");
    const results = await page.evaluate(async () => {
      const sample = document.querySelector("#sample");
      const text = "I’ll check it. I’m here, and it won’t take long. ‘Yes,’ she said, “don’t worry.” 中文";
      sample.textContent = text;
      const results = [];
      for (const weight of [400, 500, 600]) {
        sample.style.fontWeight = String(weight);
        sample.style.fontSize = "32px";
        const style = getComputedStyle(sample);
        await document.fonts.load(`${weight} 32px ${style.fontFamily}`, text);
        await document.fonts.ready;
        const canvas = document.createElement("canvas").getContext("2d");
        canvas.font = `${weight} 32px ${style.fontFamily}`;
        const widths = Object.fromEntries(Array.from("‘’“”'中文").map((char) => [char, canvas.measureText(char).width / 32]));
        results.push({ weight, widths, text: sample.textContent, family: style.fontFamily });
      }
      return results;
    });
    for (const { weight, widths, text, family } of results) {
      assert.match(family, /^"Inter Studio"/);
      assert.match(text, /I’ll.*I’m.*won’t/);
      for (const char of "‘’“”'") {
        assert.ok(widths[char] > 0 && widths[char] < 0.6,
          `${pass}, weight ${weight}: ${char} occupies ${widths[char]}em; expected narrow Latin punctuation`);
      }
      for (const char of "中文") assert.ok(Math.abs(widths[char] - 1) < 0.05, `${char} must retain full-width CJK spacing`);
      console.log(`ok ${pass}, weight ${weight}: curly quotes are narrow; Chinese remains full-width`);
    }
  }
} finally {
  await browser.close();
}
