# 北师大版2019高中英语词库

覆盖七册 **Vocabulary in Each Unit**，包含表内单词、短语、缩写及专有名词，保留 Topic Talk、Lesson 1/2/3、Writing Workshop、Reading Club。字母索引用于交叉检查。课文全部词形、不规则动词附录和独立 Names and Places 表不混作新增单元词条。

运行时 JSON：`services/api/internal/textbook/data/`。本目录 CSV 是同一数据的导出；CSV 的 `word` 同时保留 `*` 前缀，JSON 将星号单独保存在 `starred`。

| 字段 | 含义 |
| --- | --- |
| `id` | 册次内出现位置编号 |
| `word` / `phonetic` / `pos` / `chinese` | 英文、音标、词性、中文；原表未单列的音标或词性留空，不编造 |
| `unit` / `unit_title` / `lesson` | 原册 Unit 编号、标题和课节 |
| `starred` | 教材原星号，表示必修及选择性必修课标以外词汇 |
| `source_url` / `source_page` | 出处及 PDF 物理页，便于回查；不在学习界面显示页码 |

主要来源为北京师范大学出版社教材扫描件：[ChinaTextbook 北师大英语目录](https://github.com/TapXWorld/ChinaTextbook/tree/master/高中/英语/北师大版-北京师范大学出版社)。按单元词表41页核对；公开 PDF 分片链接保留在条目中，项目不提交教材全文扫描件。

参考索引：柯帕斯 [必修1](https://www.cpsenglish.com/article/733)、[必修2](https://www.cpsenglish.com/article/734)、[必修3](https://www.cpsenglish.com/article/735)、[选择性必修1](https://www.cpsenglish.com/article/736)、[选择性必修2](https://www.cpsenglish.com/article/737)、[选择性必修3](https://www.cpsenglish.com/article/739)、[选择性必修4](https://www.cpsenglish.com/article/740)。不包含该站教学文章或练习解析。

已修正 `radium → stadium`、`passionate` 音标、`occasionally`/`comfortably` 等词性、课节误归及漏星号；补入原教材中的 `saying`、`humbly`、`independence`、`reflect on`，并补入原版字母表中的 `intend`（选择性必修四）。修正见 `corrections.json`。

`source_page=0` 的3条只见于网络索引：必修二 `instant`、`instant messenger`，选择性必修一 `so far`。保留真实网络出处，可能涉及印次差异；不伪造原教材定位页。不同印次个别词条、顺序和释义可能不同，校对时应结合自己的教材。

`manifest.json` 保存各册/Unit/课节计数、星号计数和 JSON SHA-256。执行 `python3 scripts/check_textbooks.py` 检查 JSON/CSV 一致性、校验和和归属边界。
