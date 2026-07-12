const fs = require("node:fs");
const http = require("node:http");
const path = require("node:path");
const { pathToFileURL } = require("node:url");
const { chromium } = require("playwright");

const [input, output, mode = "static"] = process.argv.slice(2);
if (!input || !output) throw new Error("input and output are required");

(async () => {
  let server;
  let targetUrl = pathToFileURL(path.resolve(input)).href;
  if (mode === "slidev") {
    const root = path.dirname(path.resolve(input));
    const contentTypes = {
      ".css": "text/css; charset=utf-8",
      ".html": "text/html; charset=utf-8",
      ".js": "text/javascript; charset=utf-8",
      ".json": "application/json; charset=utf-8",
      ".png": "image/png",
      ".svg": "image/svg+xml",
      ".woff2": "font/woff2",
    };
    server = http.createServer((request, response) => {
      const pathname = decodeURIComponent(new URL(request.url, "http://127.0.0.1").pathname);
      const relative = pathname === "/" ? "index.html" : pathname.replace(/^\/+/, "");
      let candidate = path.resolve(root, relative);
      if (!candidate.startsWith(`${root}${path.sep}`) || !fs.existsSync(candidate) || fs.statSync(candidate).isDirectory()) {
        candidate = path.join(root, "index.html");
      }
      response.setHeader("Content-Type", contentTypes[path.extname(candidate).toLowerCase()] || "application/octet-stream");
      fs.createReadStream(candidate).pipe(response);
    });
    await new Promise((resolve, reject) => {
      server.once("error", reject);
      server.listen(0, "127.0.0.1", resolve);
    });
    targetUrl = `http://127.0.0.1:${server.address().port}/`;
  }
  const executablePath = process.env.CHROME_PATH && fs.existsSync(process.env.CHROME_PATH)
    ? process.env.CHROME_PATH
    : undefined;
  const browser = await chromium.launch({ headless: true, executablePath });
  try {
    const page = await browser.newPage({ viewport: { width: 1280, height: 720 }, deviceScaleFactor: 1 });
    const errors = [];
    page.on("pageerror", (error) => errors.push(error.message));
    await page.goto(targetUrl, { waitUntil: "load" });
    await page.waitForTimeout(900);
    const first = await page.screenshot({ path: output });
    const minimumBytes = mode === "pointer" ? 5_000 : 15_000;
    if (first.length < minimumBytes) throw new Error(`render is suspiciously small: ${first.length}`);
    if (mode === "pointer") {
      await page.mouse.move(240, 260);
      await page.mouse.move(900, 440, { steps: 72 });
      await page.waitForTimeout(900);
      const secondPath = output.replace(/\.png$/i, "-after.png");
      const second = await page.screenshot({ path: secondPath });
      if (Buffer.compare(first, second) === 0) throw new Error("pointer/animation probe produced identical frames");
    }
    if (mode === "presenter") {
      const popupPromise = page.waitForEvent("popup", { timeout: 5000 });
      await page.keyboard.press("s");
      const popup = await popupPromise;
      await popup.waitForLoadState("domcontentloaded");
      const dragHandles = await popup.locator("[data-drag]").count();
      const resizeHandles = await popup.locator(".resize-handle, [data-resize]").count();
      if (dragHandles < 4 || resizeHandles < 1) throw new Error(`presenter drag/resize contract failed: ${dragHandles}/${resizeHandles}`);
    }
    if (errors.length) throw new Error(`page errors: ${errors.join(" | ")}`);
    console.log(JSON.stringify({ input, output, bytes: first.length, mode }));
  } finally {
    await browser.close();
    if (server) await new Promise((resolve) => server.close(resolve));
  }
})().catch((error) => { console.error(error); process.exit(1); });
