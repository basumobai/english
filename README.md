# English

专注高中英语的自部署学习平台。从 [learning](https://github.com/basumobai/learning) 的 `de6f5ad8325412027eca9b037113f4ebfdf32d76` 迁移英语功能及其账号、模型、资料处理依赖。

- 词汇识别、释义回忆、中译英拼写、听音拼写；FSRS、新词预算、自由练习、提示、撤销、暂停与恢复。
- 北师大版2019必修1–3、选择性必修1–4，按 **册次 → Unit → Lesson** 浏览和练习，保留星号。
- 普通6秒、困难/绝望3秒英中/中英四选一；阅读、七选五、完形、语法填空、应用文、读后续写、评分与动态讲解。
- AI 对话、长难句交互、听力、广东听说 A/B/C、录音、手动转写反馈和自然 TTS。
- 英语 PDF 上传/目录扫描、A/B/C 分类解析审核、任务控制、OpenList/WebDAV 导入。
- 服务商和模型目录、身份组、任务路由、个人 API；注册、密码、TOTP、数据导出；手机、平板与电脑适配。

## 启动

准备 Docker Engine、Docker Compose v2、Git、Python 3、OpenSSL、curl：

```bash
git clone https://github.com/basumobai/english.git
cd english
bash scripts/install.sh
```

安装器会询问空闲端口、生成 `.env` 和初始管理员密码、构建并启动服务。同机已有 learning 时请选择其他端口。English 默认使用独立的 Compose 项目名 `english`、数据库卷和登录 Cookie。

本机打开 `http://localhost:<WEB_PORT>`，使用 `admin` 和初始密码登录并改密。服务器默认监听 `127.0.0.1`；正式 HTTPS 反向代理需同步修改 `.env` 的 `APP_ORIGIN`。

七册词库随 API 编译，首次启动自动入库。重复启动不会重置复习记录、暂停状态或自建词库。普通词汇练习无需 AI；TTS、AI 分析和 PDF 识别需自行接入模型并授权身份组，费用由服务商收取。

## 教材词库

| 册次 | Unit | 条目 | 星号 |
| --- | --- | ---: | ---: |
| 必修第一册 | 1–3 | 359 | 13 |
| 必修第二册 | 4–6 | 310 | 21 |
| 必修第三册 | 7–9 | 273 | 22 |
| 选择性必修第一册 | 1–3 | 306 | 37 |
| 选择性必修第二册 | 4–6 | 343 | 34 |
| 选择性必修第三册 | 7–9 | 308 | 69 |
| 选择性必修第四册 | 10–12 | 238 | 40 |
| **合计** | **21个Unit** | **2,137** | **236** |

保存词、短语和缩写的音标、词性、中文、课节、星号及出处。跨册出现位置保留，学习义项复用记忆状态。原文未单列的音标/词性留空。3条仅见网络索引的记录单独标注。

- [CSV、统计与校正记录](data/vocabulary/) · [JSON](services/api/internal/textbook/data/)
- [数据来源与字段](data/vocabulary/README.md) · [迁移范围与验证](docs/ENGLISH_MIGRATION.md)

## 入口和更新

| 功能 | 路径 |
| --- | --- |
| 词汇练习 / 词库 | `/english` / `/vocabulary` |
| 限时挑战 / 英语专项 | `/challenge` / `/practice?subject=english` |
| AI、句子、听力与听说 | `/assistant` |
| 账户、模型、身份组、自建词库 | `/admin` |
| 资料工作台 / 任务 | `/admin/resources` / `/admin/resources/jobs` |
| 原双 PDF 导入与兼容配置 | `/admin/learning` |

```bash
git pull --ff-only
bash scripts/install.sh --update
```

更新重建前端并保留数据库和密钥。请备份数据库、`.env` 与 `APP_ENCRYPTION_KEY`；常规更新不要执行 `docker compose down -v`。English 不连接或修改原 learning 部署。

## 开发与验证

后端 Go 1.26+，前端 Node.js 24：

```bash
python3 scripts/check_textbooks.py
bash scripts/test-install.sh
cd services/api
go test -race -count=1 ./...
go vet ./...
go build ./cmd/api
cd ../../apps/web
npm ci
npm test
npm run build
```

`TEST_DATABASE_URL` 指向一次性 PostgreSQL 实例时执行数据库集成测试，未设置则明确跳过。GitHub Actions 提供 PostgreSQL 17、Chromium 多端测试与真实 Docker Compose 学习闭环。自动测试使用模拟 AI 服务，不消耗模型费用。执行结果以对应提交的 Actions 为准。

部署细节见 [DEPLOY](docs/DEPLOY.md) 与 [SECURITY](docs/SECURITY.md)。保留的上游设计文档用于共享引擎参考；当前产品范围以本 README 和迁移记录为准。
