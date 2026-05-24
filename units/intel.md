# 情报局 — 多引擎并行搜索 + 情报研判流水线

## 核心定位
多引擎并行搜索→信息提取→可信度评估→结构化报告。与参谋本部MoE对接。

---

## 一、多引擎并行搜索（MoE模式）

### 搜索引擎池
| 引擎 | 用途 | 速率限制 |
|------|------|---------|
| GitHub API | 代码/仓库/技术方案搜索 | 60 req/hr (未认证) / 5000 req/hr (已认证) |
| 网页搜索(DeepSeek) | 通用信息/竞品/社区 | 按API配额 |
| MCP:openai-docs | OpenAI官方文档 | 无限制 |

### 并行搜索协议
```
[搜索指令]
    ↓
参谋本部拆分为多个搜索子任务
    ↓
并行派发给多个搜索引擎（Promise.allSettled）
  ├─ GitHub: 搜索仓库+README+issues
  ├─ Web: 搜索社区+论坛+博客
  └─ MCP: 查询官方文档（如适用）
    ↓
结果汇总 → 去重 → 排序 → 可信度评估
```

### 搜索模板库

| 场景 | 模板 |
|------|------|
| 找技能/工具 | `"<关键词> agent skill" OR "<关键词> AI tool" sort:stars` |
| 找竞品 | `"<产品名>" alternatives OR vs OR comparison` |
| 找最佳实践 | `"<技术栈>" best practices OR production OR enterprise` |
| 安全审计 | `"<库名>" vulnerability OR CVE OR security advisory` |
| 许可证检查 | `"<库名>" license` |

---

## 二、情报研判流程

### 可信度评分（5分制）
| 分数 | 标准 |
|------|------|
| 5/5 | 官方文档/知名机构/有源码验证 |
| 4/5 | 高Star开源项目/知名专家 |
| 3/5 | 社区公认/多来源印证 |
| 2/5 | 单一来源/个人博客 |
| 1/5 | 无来源/无法验证/可疑 |

### 研判输出格式
```
🔍 情报报告
├─ 来源: [链接/仓库]
├─ 可信度: [1-5]/5
├─ 摘要: [关键发现]
├─ 可融合性: [高/中/低] — [理由]
└─ 安全注意: [许可证/漏洞/风险]
```

---

## 三、与各部门协作

| 协作方 | 交互方式 | 输出物 |
|--------|---------|--------|
| 参谋本部(MoE) | 接收搜索任务 | 情报报告 |
| 安全局 | 移交安全一审 | 来源+可信度+许可证 |
| 试验场 | 提供技术调研 | 竞品对比/方案分析 |
| 女娲 | 推荐专家 | 领域/技术匹配建议 |

---

## 四、直接可执行的搜索命令

### GitHub搜索
```powershell
# 语法：搜索仓库
Invoke-RestMethod -Uri "https://api.github.com/search/repositories?q=<关键词>&sort=stars&order=desc&per_page=5"

# 搜索代码
Invoke-RestMethod -Uri "https://api.github.com/search/code?q=<关键词>"
```

### 网页搜索
通过 DeepSeek 免费通道的网页搜索能力获取实时信息。
