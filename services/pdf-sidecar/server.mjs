import { createServer } from "node:http";
import { chromium } from "playwright";

const port = Number(process.env.PORT || 3001);
const browser = await chromium.launch({ args: ["--no-sandbox"] });

function sendJSON(res, status, payload) {
  res.writeHead(status, { "content-type": "application/json" });
  res.end(JSON.stringify(payload));
}

async function readJSON(req) {
  const chunks = [];
  for await (const chunk of req) chunks.push(chunk);
  if (!chunks.length) return {};
  return JSON.parse(Buffer.concat(chunks).toString("utf8"));
}

function withTimeout(promise, ms = 30000) {
  return Promise.race([
    promise,
    new Promise((_, reject) =>
      setTimeout(() => reject(new Error(`timed out after ${ms}ms`)), ms)
    ),
  ]);
}

async function withPage(run) {
  const ctx = await browser.newContext({
    userAgent:
      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36",
    viewport: { width: 1280, height: 800 },
  });
  const page = await ctx.newPage();
  page.setDefaultTimeout(30000);
  try {
    return await withTimeout(run(page), 30000);
  } finally {
    await ctx.close();
  }
}

const server = createServer(async (req, res) => {
  const url = new URL(req.url, `http://localhost:${port}`);

  if (req.method === "GET" && url.pathname === "/health") {
    return sendJSON(res, 200, { status: "ok", browser: "ready" });
  }

  // POST /generate-pdf  { html: string }  → { pdf: base64 }
  if (req.method === "POST" && url.pathname === "/generate-pdf") {
    try {
      const { html } = await readJSON(req);
      const pdf = await withPage(async (page) => {
        await page.setContent(html, { waitUntil: "networkidle" });
        return page.pdf({ format: "A4", printBackground: true });
      });
      return sendJSON(res, 200, { pdf: pdf.toString("base64") });
    } catch (err) {
      return sendJSON(res, 500, { error: err.message });
    }
  }

  // POST /scrape  { url: string }  → { title, text }
  if (req.method === "POST" && url.pathname === "/scrape") {
    try {
      const { url: target } = await readJSON(req);
      const result = await withPage(async (page) => {
        await page.goto(target, { waitUntil: "domcontentloaded" });
        const title = await page.title();
        const text = await page.evaluate(() => document.body.innerText);
        return { title, text: text.slice(0, 10000) };
      });
      return sendJSON(res, 200, result);
    } catch (err) {
      return sendJSON(res, 500, { error: err.message });
    }
  }

  sendJSON(res, 404, { error: "not found" });
});

server.listen(port, () => console.log(`sidecar listening on :${port}`));
