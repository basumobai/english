import assert from "node:assert/strict";
import { chromium } from "playwright";
import { installMockAPI } from "./mock-api.mjs";
const browser = await chromium.launch({
  headless: true,
  executablePath: process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH || undefined,
  args: ["--no-sandbox"],
});
const base = process.env.UI_BASE || "http://127.0.0.1:3007";
let checks = 0;
try {
  const context = await browser.newContext({
    viewport: { width: 390, height: 844 },
  });
  const mock = await installMockAPI(context),
    page = await context.newPage(),
    errors = [];
  page.on("pageerror", (e) => errors.push(e.message));
  await page.goto(base + "/vocabulary");
  await Promise.all([
    page.waitForResponse(
      (r) =>
        r.url().includes("vocabulary?") &&
        r.url().includes("entry_kind=phrase"),
    ),
    page.getByLabel("词条类型", { exact: true }).selectOption("phrase"),
  ]);
  checks++;
  await page.goto(base + "/english");
  await page.getByLabel("词条类型", { exact: true }).selectOption("phrase");
  // The formal study API, not just the browser list, carries the phrase constraint.
  await page.getByRole("button", { name: "开始学习", exact: true }).click();
  await page.waitForFunction(() => !!document.querySelector(".question"));
  assert.equal(
    mock.requests.findLast((r) => r.path === "/study/sessions" && r.body)?.body
      .entry_kind,
    "phrase",
  );
  checks++;
  await page.goto(base + "/challenge");
  await page.getByLabel("作答计时", { exact: true }).selectOption("2");
  await page.getByLabel("词条类型", { exact: true }).selectOption("phrase");
  await page.getByRole("button", { name: "开始挑战", exact: true }).click();
  await page.locator(".challenge-options button").last().waitFor();
  assert.ok(
    Number(await page.locator(".challenge-clock strong").innerText()) <= 2,
  );
  checks++;
  const start = mock.requests.findLast(
    (r) => r.path === "/learning/words/next",
  );
  assert.equal(start.body.seconds, 2);
  assert.equal(start.body.entry_kind, "phrase");
  checks++;
  await page.locator(".challenge-options button").nth(1).click();
  await page.getByRole("button", { name: "下一题", exact: true }).waitFor();
  const count = mock.requests.filter(
    (r) => r.path === "/learning/words/next",
  ).length;
  await page.waitForTimeout(1400);
  assert.equal(
    mock.requests.filter((r) => r.path === "/learning/words/next").length,
    count,
    "wrong answer skipped feedback",
  );
  checks++;
  assert.equal(
    mock.requests.filter((r) => r.path.endsWith("/messages")).length,
    0,
  );
  checks++;
  await page.getByRole("button", { name: "请 AI 讲解用法与易混点" }).click();
  await page
    .getByText("adapt to 表示适应；这是 AI 原创例句。", { exact: true })
    .waitFor();
  checks++;
  await page.screenshot({
    path: "../../docs/previews/english-phrase-challenge.png",
    fullPage: true,
  });
  await page.getByRole("link", { name: /继续追问或提交小练习答案/ }).click();
  await page.getByText("adapt to 表示适应。", { exact: true }).waitFor();
  checks++;

  const sessions = new Map(),
    calls = [];
  const analysis = {
    translation: "这本书帮助了我。",
    main_clause: "The book helped me.",
    chunks: [{ text: "The book", role: "主语", explanation: "书是主语" }],
    grammar: ["主谓宾结构"],
    paraphrase: "This book was helpful.",
    check: {
      question: "什么帮助了我？",
      answer: "The book",
      explanation: "书是主语",
    },
  };
  await context.route("**/api/v1/learning/english/**", async (route) => {
    const p = new URL(route.request().url()).pathname.slice("/api/v1".length),
      body = route.request().postDataJSON();
    calls.push({ p, body });
    let data = {};
    if (p === "/learning/english/syntax") {
      if (route.request().method() === "GET")
        data = {
          sessions: [...sessions.values()].map((s) => ({
            id: s.id,
            title: s.text,
          })),
        };
      else {
        data = { id: "syntax1", text: body.text, analysis: null, attempts: [] };
        sessions.set(data.id, data);
      }
    } else {
      const id = p.split("/")[4],
        session = sessions.get(id);
      if (p.endsWith("/attempt")) {
        session.attempts.push({
          marks: body.marks,
          translation: body.translation,
          skipped: body.skip,
        });
        data = { saved: true };
      } else if (p.endsWith("/reveal")) {
        session.analysis = analysis;
        data = session;
      } else if (p.endsWith("/transfer")) {
        data = { id: "syntax2" };
        sessions.set(data.id, {
          id: data.id,
          text: "The film inspired her.",
          parent_id: id,
          ai_generated: true,
          analysis: null,
          attempts: [],
        });
      } else data = session;
    }
    await route.fulfill({ json: data });
  });
  await page.goto(base + "/assistant#sentence");
  await page.getByLabel("独立拆句原句").fill("The book helped me.");
  await page.getByRole("button", { name: "开始独立拆句", exact: true }).click();
  await page.getByRole("button", { name: "选择词 1 The", exact: true }).click();
  await page
    .getByRole("button", { name: "选择词 2 book", exact: true })
    .click();
  await page.getByLabel("我的翻译", { exact: true }).fill("这本书帮助了我。");
  assert.equal(calls.filter((c) => c.p.endsWith("/reveal")).length, 0);
  checks++;
  await page.getByRole("button", { name: "提交尝试并看反馈" }).click();
  await page.getByRole("heading", { name: "对照参考分析" }).waitFor();
  assert.deepEqual(calls.find((c) => c.p.endsWith("/attempt")).body.marks, [
    { start: 0, end: 8, role: "subject" },
  ]);
  checks++;
  await page.getByRole("button", { name: "换新句验证", exact: true }).click();
  await page
    .getByRole("button", { name: "选择词 2 film", exact: true })
    .waitFor();
  assert.equal(
    await page.getByRole("heading", { name: "对照参考分析" }).count(),
    0,
  );
  checks++;
  await page.getByRole("button", { name: "跳过尝试，直接看讲解" }).click();
  await page.getByRole("heading", { name: "对照参考分析" }).waitFor();
  assert.equal(
    calls.filter((c) => c.p.endsWith("/attempt")).at(-1).body.skip,
    true,
  );
  checks++;
  await page.reload();
  await page.getByText("恢复独立拆句记录（最近30条）", { exact: true }).click();
  await page
    .getByRole("button", { name: "The book helped me.", exact: true })
    .click();
  await page.getByRole("heading", { name: "对照参考分析" }).waitFor();
  checks++;
  assert.ok(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth + 1,
    ),
  );
  checks++;
  await page.screenshot({
    path: "../../docs/previews/english-independent-syntax.png",
    fullPage: true,
  });
  // Saving is useful without paying for ASR; transcription is a separate action.
  let uploaded = false,
    asrCalls = 0,
    transcript = null;
  const spoken = {
    id: "asr-lesson",
    kind: "listening",
    format: "guangdong_a",
    title: "朗读",
    scenario: "朗读原句",
    keywords: [],
    reading_text: "I will meet you on Friday.",
    speaking_tasks: [{ index: 0, kind: "read", cue: "朗读" }],
  };
  await context.route("**/api/v1/ai/models", (route) =>
    route.fulfill({
      json: {
        models: [
          { id: 1, name: "讲解模型", kind: "chat", voices: [] },
          { id: 8, name: "ASR模型", kind: "asr", voices: [] },
        ],
      },
    }),
  );
  await context.route("**/api/v1/ai/training", async (route) => {
    if (route.request().method() !== "POST") return route.fallback();
    await route.fulfill({ json: spoken });
  });
  await context.route("**/api/v1/learning/recordings?*", (route) =>
    route.fulfill({
      json: { recordings: uploaded ? [{ id: "saved-audio", transcript }] : [] },
    }),
  );
  await context.route("**/api/v1/learning/recordings", async (route) => {
    const body = route.request().postData();
    assert.ok(body.includes("asr-lesson"));
    uploaded = true;
    await route.fulfill({ status: 201, json: { id: "saved-audio" } });
  });
  await context.route(
    "**/api/v1/learning/recordings/saved-audio/transcribe",
    async (route) => {
      asrCalls++;
      assert.equal(route.request().postDataJSON().model_id, 8);
      transcript = "I will meet you on Friday.";
      await route.fulfill({
        json: { text: transcript, note: "自动转写，请核对" },
      });
    },
  );
  await page.goto(base + "/assistant#listening");
  await page
    .getByRole("button", { name: "生成新的听力练习", exact: true })
    .click();
  await page.getByText("保存录音并自动转写（可选）", { exact: true }).click();
  await page.getByLabel("或选择音频（小于20MB）").setInputFiles({
    name: "sample.wav",
    mimeType: "audio/wav",
    buffer: Buffer.from("RIFF1234WAVEfixture"),
  });
  await page.getByRole("button", { name: "仅保存录音", exact: true }).click();
  await page.getByText("录音已保存到本人账户。", { exact: true }).waitFor();
  assert.equal(asrCalls, 0);
  checks++;
  await page
    .getByRole("button", { name: "转写这段已保存录音", exact: true })
    .click();
  await page.getByText("自动转写，请核对", { exact: true }).waitFor();
  assert.equal(
    await page
      .getByLabel("我的英文作答／录音转写", { exact: true })
      .inputValue(),
    transcript,
  );
  assert.equal(asrCalls, 1);
  checks++;
  await page.reload();
  await page
    .getByRole("button", { name: "生成新的听力练习", exact: true })
    .click();
  await page.getByText("保存录音并自动转写（可选）", { exact: true }).click();
  await page
    .getByRole("button", { name: "使用已保存转写", exact: true })
    .click();
  assert.equal(
    await page
      .getByLabel("我的英文作答／录音转写", { exact: true })
      .inputValue(),
    transcript,
  );
  assert.equal(asrCalls, 1);
  checks++;
  await page.screenshot({
    path: "../../docs/previews/english-recording-transcription.png",
    fullPage: true,
  });
  assert.deepEqual(errors, []);
  checks++;
  console.log(`English teaching UI checks passed: ${checks}`);
} catch (error) {
  const p = browser.contexts()[0]?.pages()[0];
  if (p) {
    console.log((await p.locator("main").innerText()).slice(-7000));
    await p.screenshot({
      path: "../../docs/previews/english-debug.png",
      fullPage: true,
    });
  }
  throw error;
} finally {
  await browser.close();
}
