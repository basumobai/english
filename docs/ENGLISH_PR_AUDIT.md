# 英语功能与历史 PR 核对

核对日期：2026-10-24。基线：learning main `de6f5ad8325412027eca9b037113f4ebfdf32d76`，GitHub全部23个PR的标题、正文、合并/开放状态，并读取#13路线图、#20独立英语和#22录音源文件。已合并成果从main迁移；开放PR中的独立英语模块按当前架构移植，没有将未合并规划或其他五科依赖一并标成完成。

## 教学目的

#1最早目标是通过OpenAI兼容接口提供英语AI助学：独立账户、授权模型、FSRS回忆、句法拆解和听力。#3加入快选四选一；#4强调答后和主动求助才讲解；#13/#20进一步强调学生先独立尝试，AI给反馈，再用新内容验证。English保留供应商接口和权限底座，把这些教学动作连起来。

## 对照

| 上游PR | 核对状态 | 目的/英语相关内容 | English处理 |
| --- | --- | --- | --- |
| [#1](https://github.com/basumobai/learning/pull/1) | 已合并 | 账户/权限组、四维FSRS、导入、对话、拆句、听力、自然TTS | 已迁入，保留本人记录与模型服务端授权 |
| [#2](https://github.com/basumobai/learning/pull/2) | 已合并 | Apple风格界面、移动导航 | 已迁入；English首页单入口改为占满可用宽度 |
| [#3](https://github.com/basumobai/learning/pull/3) | 已合并 | 6秒/3秒四选一、PDF入库、动态讲解、个人接口 | 英语部分已迁入；本轮独立加入2秒、教材/短语范围和答错停留 |
| [#4](https://github.com/basumobai/learning/pull/4) | 已合并 | 答错/主动求助后讲解、身份组模型选择 | 已迁入；词汇新增按需用法讲解及继续追问 |
| [#5](https://github.com/basumobai/learning/pull/5) | 已合并 | 模型范围选择、批量授权、多端 | 已迁入；ASR追加独立用途及组授权 |
| [#6](https://github.com/basumobai/learning/pull/6) | 已合并 | 一键安装 | 已迁入，English独立Compose项目 |
| [#7](https://github.com/basumobai/learning/pull/7) | 已合并 | 手机/平板/桌面作答 | 已迁入；新增交互纳入浏览器回归 |
| [#8](https://github.com/basumobai/learning/pull/8) | 已合并 | 仓库地址 | English文档改为新仓库地址 |
| [#9](https://github.com/basumobai/learning/pull/9) | 已合并 | 端口与冲突检测 | 已迁入 |
| [#10](https://github.com/basumobai/learning/pull/10) | 已合并 | 安全部署/注册、教材确认、广东A/B/C | 已迁入；新增录音持久化/ASR，仍不冒充发音评分 |
| [#11](https://github.com/basumobai/learning/pull/11) | 已合并 | v2审计与设计文档 | 作为架构背景；没有把设计文档当作功能 |
| [#12](https://github.com/basumobai/learning/pull/12) | 已合并 | 模型目录/路由、A/B/C资料流水线/Worker | 英语处理依赖已迁入 |
| [#13](https://github.com/basumobai/learning/pull/13) | 开放 | v3.1教学目标与分阶段规划 | 参考独立尝试→反馈→新题验证；未把整份路线图标成已实现 |
| [#14](https://github.com/basumobai/learning/pull/14) | 已合并 | 管理导航、模型弹窗、防旧缓存 | 已迁入 |
| [#15](https://github.com/basumobai/learning/pull/15) | 开放 | 反向代理旧HTML修复 | 核对其目的；English保留当前响应/更新策略，不叠加未经合并的整套变更 |
| [#16](https://github.com/basumobai/learning/pull/16) | 开放 | 评分修订、不可变证据 | 通用v3依赖链；本轮未移植或启用这套新评分引擎 |
| [#17](https://github.com/basumobai/learning/pull/17) | 开放 | 知识映射与能力证据 | 依赖#16；本轮不由AI结果直接推导能力百分比 |
| [#18](https://github.com/basumobai/learning/pull/18) | 开放 | 错因候选/确认/更正 | 依赖#17；本轮词汇反馈不宣称具备整套错因引擎 |
| [#19](https://github.com/basumobai/learning/pull/19) | 开放 | 今日计划、自适应推荐、主动回忆卡 | 依赖#18；保留原词汇FSRS，不混用两套新引擎 |
| [#20](https://github.com/basumobai/learning/pull/20) | 开放 | 独立长难句标注/翻译、新句验证、听力反思 | 已单独移植英语模块，去除无关v3依赖；补恢复列表、对照本人尝试和取消请求清理 |
| [#21](https://github.com/basumobai/learning/pull/21) | 开放 | 审核资源检索、引导导师、模型教学评测 | 依赖评分/引用修订链；保留现有AI导师，不声称完成RAG/质量门禁 |
| [#22](https://github.com/basumobai/learning/pull/22) | 开放 | 整卷快照与录音/ASR | 单独移植录音/ASR及模型用途；整卷考试依赖前序v3链，本轮未移植 |
| [#23](https://github.com/basumobai/learning/pull/23) | 开放 | 弱网/账号、运维与备份恢复 | 参考请求幂等与隔离；保留现有复习恢复，不声称已做完整生产恢复 |

## 本轮可以直接使用的变化

- **短语**：七册167个短语/搭配出现位置已在原2,137条内，完整保留词形、空缺字段、中文、课节、星号与来源。新增单词/短语范围，贯穿浏览、正式复习、自由练习、快选的真实SQL。短语四字段CSV/TSV允许音标和词性留空；同时修复TSV连续空列被跳过。
- **快选**：默认保留普通6秒、困难/绝望3秒，计时可独立选2/3/6秒。服务端写入截止和请求设置，重试相同载荷恢复原题，不同载荷409；切页/停止不延长截止。答错或超时停留，答对自动下一题可关。范围无已审核内容时明确提示，管理员可从词库直接带词义编号进入生成/审核。
- **词汇AI**：词库与快选答后主动请求讲解，调用本人已授权聊天模型。给用法/搭配/易混点、原创例句与一道未揭答案的新练习；讲解保存到本人对话，可继续回答和追问。AI不重写教材或正式判分。
- **独立拆句**：先点起止词标主语/谓语/从句/修饰等片段，写翻译，或明确跳过；服务端验证Unicode范围。提交后AI结合原句及学生尝试分析，再换新句验证。原练习只有本人已交卷英语原文可进入。最近30条可恢复，首次尝试和跳过事实保留。
- **听力**：交卷后区分没听懂/听懂后忘记，保存关键词反思与不同训练步骤；可恢复。
- **听说录音**：本人录音可主动保存/回放，单独授权ASR用途，按OpenAI兼容Audio Transcriptions接口转写，核对后放入作答。无ASR仍可保存、回放、下载、手填；录音和转写均可恢复，不提供官方或专业发音评分。

## 来源与兼容

- #20源提交：`776b58fa083aec4e93926feb4b3be3ec1663648a`；#22源提交：`9badac2aa358cb992f71d267fac5248ea70ee9dd`。复制的是用户learning代码，未引入其他平台源码。
- 迁移0017–0020只新增派生类型、独立交互、请求快照、录音/转写表及ASR用途，不清空词义、复习、暂停、账号或旧答卷。保持现有Go/Next.js、Worker、供应商/身份组授权。
- 未合并通用v3依赖链（评分/知识追踪/自适应计划/RAG/整卷）仍需后续独立整合，不能只搬页面按钮并声称实现。

## 验证

语料脚本校验7册/21个Unit、2,137出现位置、236星号及167短语；Go单元/集成测试覆盖短语导入、SQL/Go分类一致、教材/短语学习范围、2秒截止、载荷冲突、答案隐藏、晚交/重放、独立尝试门槛、恢复及账号隔离、ASR调用去重。浏览器覆盖2秒/短语请求、错答停留、主动AI、对话衔接、独立标注/新句/恢复，另运行既有多端回归。

实际执行结果记录在 [本轮验收](ENGLISH_ENHANCEMENTS_VALIDATION.md)。AI/ASR自动测试使用模拟供应商，未宣称真实转写准确度、付费服务教学质量或生产部署验收。

