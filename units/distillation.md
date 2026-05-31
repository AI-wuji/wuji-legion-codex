# 蒸馏师团 - 官方源核验 + 能力蒸馏 + 瘦身闸门

## 核心定位

蒸馏师团隶属进化部，执行入口是 `进化主帅`。

目标不是装更多 skill，而是把外部有效机制拆出来，蒸馏进现有师团主帅的内置模式里。

---

## 蒸馏流程

```text
source scan
-> latest check
-> source/code read
-> license check
-> failure-mode mapping
-> owner mapping
-> absorb/defer/reject
-> validation
-> changelog
```

## 裁决标准

| 裁决 | 条件 |
|---|---|
| absorb | 明确解决失败模式，能落入现有主帅/模式，不显著增加 token 噪音 |
| defer | 有价值但缺验证、缺许可证、缺适用场景或当前不急 |
| reject | 只增加口号/重复入口/外部编制/上下文污染 |

## 当前已核验来源

| 来源 | 最新检查 | commit | 吸收点 |
|---|---|---|---|
| `openai/skills` | 2026-05-31 | `a8924c2` | skill 结构、按需加载、资源分层 |
| `anthropics/skills` | 2026-05-31 | `da20c92` | skill 规范、轻上下文、工具资源分离 |
| `addyosmani/agent-skills` | 2026-05-31 | `6ce0298` | 阶段识别、上下文工程、薄切片、五轴review |
| `github/awesome-copilot` | 2026-05-31 | `9b74459` | 大规模技能索引、Rust/QA/安全细分规则、bundled assets |
| `marketingskills` | 2026-05-31 | `7f4af1e` | 产品/受众/定位先行，营销技能互相关联但有主线 |
| `humanizer` | 2026-05-31 | `a2ace14` | AI写作痕迹识别、作者声音匹配 |
| `powerpoint-skill` | 2026-05-31 | `a39cd8c` | 视觉优先、密度边界、重叠检查、预览验证 |
| `ppt-master` | 2026-05-31 | `232415d` | 真PPTX结构先行、可编辑交付；clone偏慢，保留为已知参考源 |
| `AMAP-ML/SkillClaw` | 2026-05-31 | `1f96ec8` | 任务后进化、候选验证、版本轨迹 |
| `cft0808/edict` | 2026-05-31 | `14a2075` | 前置封驳、状态审计、可观测协作 |
| `vercel-labs/agent-skills` | 2026-05-31 | `1801156` | skill 组织和轻量安装参考 |
| `DannyMac180/skills` | 2026-06-01 | `5695fa1` | 动态工作流工件、packet 切片、结果收集、验证脚手架 |

## 本轮蒸馏结论

- `content`：多个写作专家压缩为 `内容主帅`，用内置模式覆盖小说、剧本、分镜、教程、计划书、营销方案、短内容和人味改稿。
- `visual`：PPT、HTML、UI、图表、信息图压缩为 `视觉主帅`，以交付物类型选择模式。
- `dev`：软件、Rust/Tauri、前端、小程序、ComfyUI插件、AI工程和自动化压缩为 `开发主帅`。
- `security`、`qa`：不合并进执行者，保持安全、合规、白帽、质检、性能基准独立。
- `staff`、`intel`、`evolve`、`expedition`、`archive`、`prompt`：均压缩为师团主帅 + 内置模式。
- `dynamic workflow`：不新增入口，蒸馏为复杂 LEGION_TASK 的最小可审计轨迹；参谋本部定 contract，交付主帅管 packets/results，质检主帅按 verification 放行。

## 退回红线

- 只看文章介绍，不看官方源或源码。
- 没有版本/commit/许可证记录。
- 不能说明吸收后解决哪个失败模式。
- 复制外部大段文本或组织编制。
- 把多个 skill 名字叠成路由噪音。
- 改一个文件不改全局，形成补丁冲突。
