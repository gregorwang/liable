# 审核规则库 API 文档

## 概述

审核规则库 API 提供了对内容审核规则的查询、筛选和搜索功能。所有规则都基于详细的审核操作手册，涵盖内容安全、用户保护和社区秩序等多个方面。

## 基础信息

- **基础 URL**: `http://localhost:8000/api`
- **认证**: 部分端点需要认证，可选
- **内容类型**: `application/json`

## API 端点

### 1. 获取规则列表

获取分页的审核规则列表，支持筛选和搜索。

```
GET /moderation-rules
```

#### 查询参数

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `category` | string | 否 | 按分类筛选（精确匹配） |
| `risk_level` | string | 否 | 按风险等级筛选：`L`/`M`/`H`/`C` |
| `search` | string | 否 | 搜索规则编号或描述（模糊匹配） |
| `page` | integer | 否 | 页码，默认为 1 |
| `page_size` | integer | 否 | 每页条数，默认为 20，最大 100 |

#### 示例请求

```bash
# 获取第一页的规则
curl "http://localhost:8000/api/moderation-rules?page=1&page_size=20"

# 按分类筛选
curl "http://localhost:8000/api/moderation-rules?category=人身安全与暴力"

# 按风险等级筛选
curl "http://localhost:8000/api/moderation-rules?risk_level=H"

# 搜索规则
curl "http://localhost:8000/api/moderation-rules?search=A1"

# 组合筛选和搜索
curl "http://localhost:8000/api/moderation-rules?category=人身安全与暴力&risk_level=C&search=威胁"
```

#### 响应示例

```json
{
  "data": [
    {
      "id": 1,
      "rule_code": "A1",
      "category": "人身安全与暴力",
      "subcategory": "真实威胁/伤害",
      "description": "对个人/群体发出可信威胁，具时间/地点/手段或历史前科",
      "judgment_criteria": "对个人/群体发出可信威胁，具时间/地点/手段或历史前科",
      "risk_level": "C",
      "action": "立即删除+永久封禁；证据保全；必要时上报",
      "boundary": "夸张口头禅但无对象与可行性→转A2或C1评估",
      "examples": "威胁对象、时间、地点、手段的具体信息",
      "quick_tag": null,
      "created_at": "2025-10-25T00:00:00Z",
      "updated_at": "2025-10-25T00:00:00Z"
    },
    {
      "id": 2,
      "rule_code": "A2",
      "category": "人身安全与暴力",
      "subcategory": "教唆/美化暴力",
      "description": "鼓动他人伤害、歌颂真实暴力事件、分享可操作做法",
      "judgment_criteria": "鼓动他人伤害、歌颂真实暴力事件、分享可操作做法",
      "risk_level": "H",
      "action": "删除+长期/永久封禁；必要时上报",
      "boundary": "新闻/纪实讨论允许；但不可夹带鼓动性语言",
      "examples": "教唆他人暴力的言论、暴力做法的具体分享",
      "quick_tag": "暴力",
      "created_at": "2025-10-25T00:00:00Z",
      "updated_at": "2025-10-25T00:00:00Z"
    }
  ],
  "total": 42,
  "page": 1,
  "page_size": 20,
  "total_pages": 3
}
```

#### 风险等级说明

| 等级 | 代码 | 说明 |
|------|------|------|
| 低风险 | L | 影响体验但可教育纠正 |
| 中风险 | M | 破坏秩序/潜在伤害，需删除或限制 |
| 高风险 | H | 明确伤害或重大风险，需强力处置 |
| 极高风险 | C | 严重违法或对人身安全构成威胁，需立刻封禁与上报 |

---

### 2. 获取单条规则

根据规则编号获取详细的规则信息。

```
GET /moderation-rules/:code
```

#### 路径参数

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `code` | string | 是 | 规则编号（如 A1, B2 等） |

#### 示例请求

```bash
curl "http://localhost:8000/api/moderation-rules/A1"
curl "http://localhost:8000/api/moderation-rules/B2"
```

#### 响应示例

```json
{
  "id": 1,
  "rule_code": "A1",
  "category": "人身安全与暴力",
  "subcategory": "真实威胁/伤害",
  "description": "对个人/群体发出可信威胁，具时间/地点/手段或历史前科",
  "judgment_criteria": "对个人/群体发出可信威胁，具时间/地点/手段或历史前科",
  "risk_level": "C",
  "action": "立即删除+永久封禁；证据保全；必要时上报",
  "boundary": "夸张口头禅但无对象与可行性→转A2或C1评估",
  "examples": "威胁对象、时间、地点、手段的具体信息",
  "quick_tag": null,
  "created_at": "2025-10-25T00:00:00Z",
  "updated_at": "2025-10-25T00:00:00Z"
}
```

#### 错误响应

- **404 Not Found**: 当规则不存在时

```json
{
  "error": "Rule not found"
}
```

---

### 3. 获取分类列表

获取所有可用的规则分类。

```
GET /moderation-rules/categories
```

#### 示例请求

```bash
curl "http://localhost:8000/api/moderation-rules/categories"
```

#### 响应示例

```json
{
  "categories": [
    "人身安全与暴力",
    "仇恨与歧视",
    "骚扰与霸凌",
    "未成年人与性相关",
    "非法与危险活动",
    "虚假信息与公共危害",
    "隐私与个人信息",
    "垃圾信息与平台安全",
    "知识产权",
    "社区秩序与质量"
  ]
}
```

---

### 4. 获取风险等级列表

获取所有可用的风险等级。

```
GET /moderation-rules/risk-levels
```

#### 示例请求

```bash
curl "http://localhost:8000/api/moderation-rules/risk-levels"
```

#### 响应示例

```json
{
  "levels": ["L", "M", "H", "C"]
}
```

---

## 规则分类体系

### A. 人身安全与暴力 (3 条规则)
- **A1**: 真实威胁/伤害 - 极高风险 (C)
- **A2**: 教唆/美化暴力 - 高风险 (H)
- **A3**: 自杀/自残相关 - 高风险 (H)

### B. 仇恨与歧视 (3 条规则)
- **B1**: 受保护群体仇恨 - 高风险 (H)
- **B2**: 非受保护群体恶意辱骂 - 中风险 (M)
- **B3**: 仇恨符号/隐语 - 高风险 (H)

### C. 骚扰与霸凌 (3 条规则)
- **C1**: 定向骚扰/跟踪暗示 - 高风险 (H)
- **C2**: 性骚扰/物化 - 中风险 (M)
- **C3**: 组织围攻/跨帖集火 - 高风险 (H)

### D. 未成年人与性相关 (3 条规则)
- **D1**: 未成年人性化/暗示 - 极高风险 (C)
- **D2**: 非自愿私密影像/威胁传播 - 极高风险 (C)
- **D3**: 成人露骨内容（非未成年人） - 中风险 (M)

### E. 非法与危险活动 (4 条规则)
- **E1**: 教唆/指导犯罪 - 高风险 (H)
- **E2**: 违禁品交易/危化品 - 极高风险 (C)
- **E3**: 诈骗/钓鱼/虚假中奖 - 高风险 (H)
- **E4**: 赌博与灰产推广 - 高风险 (H)

### F. 虚假信息与公共危害 (3 条规则)
- **F1**: 医疗健康错误信息 - 中风险 (M)
- **F2**: 公共事件/选举误导 - 高风险 (H)
- **F3**: 冒充身份/深度伪造 - 高风险 (H)

### G. 隐私与个人信息 (2 条规则)
- **G1**: 人肉搜索/信息泄露 - 高风险 (H)
- **G2**: 未经同意的影像/录音 - 高风险 (H)

### H. 垃圾信息与平台安全 (3 条规则)
- **H1**: 广告引流/模板化推广 - 中风险 (M)
- **H2**: 恶意链接/恶意软件 - 高风险 (H)
- **H3**: 刷屏/脚本灌水 - 中风险 (M)

### I. 知识产权 (2 条规则)
- **I1**: 未授权传播/盗版 - 中风险 (M)
- **I2**: 商标/肖像冒用 - 高风险 (H)

### J. 社区秩序与质量 (3 条规则)
- **J1**: 低质/题外/无信息量 - 低风险 (L)
- **J2**: 粗俗用语/不文明 - 低风险 (L)
- **J3**: 剧透/敏感未标注 - 低风险 (L)

---

## 快捷标签映射

以下快捷标签可用于快速审核操作：

| 快捷标签 | 映射规则 |
|---------|---------|
| 人身攻击 | B2, C1 |
| 垃圾 | J1, H3 |
| 广告 | H1 |
| 政治敏感 | F2 |
| 暴力 | A1, A2 |
| 色情 | D3, D1 |

---

## 处置动作参考

| 代码 | 动作 | 说明 |
|------|------|------|
| DEL | 删除 | 删除违规内容 |
| FOLD | 折叠 | 折叠内容仅本人可见 |
| HIDE | 仅自己可见 | 内容隐藏 |
| DERANK | 降权 | 降低内容权重 |
| NR | 推荐屏蔽 | 屏蔽推荐 |
| MUTE24 | 禁评24h | 禁止评论24小时 |
| MUTE7 | 禁评7d | 禁止评论7天 |
| BAN7 | 临封7d | 临时封禁7天 |
| BANP | 永封 | 永久封禁 |
| BLK | 拉黑 | 域名/IP拉黑 |
| REPORT | 上报 | 上报执法 |

---

## 错误响应

所有 API 错误都返回相应的 HTTP 状态码和 JSON 错误信息。

### 常见错误

| 状态码 | 错误 | 说明 |
|--------|------|------|
| 400 | Bad Request | 请求参数无效 |
| 404 | Not Found | 资源不存在 |
| 500 | Internal Server Error | 服务器错误 |

#### 错误响应示例

```json
{
  "error": "Failed to fetch rules"
}
```

---

## 使用示例

### Python

```python
import requests

# 获取高风险规则列表
url = "http://localhost:8000/api/moderation-rules"
params = {
    "risk_level": "H",
    "page": 1,
    "page_size": 50
}

response = requests.get(url, params=params)
rules = response.json()

for rule in rules['data']:
    print(f"{rule['rule_code']}: {rule['subcategory']}")
```

### JavaScript/TypeScript

```typescript
// 搜索规则
async function searchRules(searchTerm: string) {
  const response = await fetch(
    `/api/moderation-rules?search=${encodeURIComponent(searchTerm)}`
  );
  const data = await response.json();
  return data.data;
}

// 按分类获取规则
async function getRulesByCategory(category: string) {
  const response = await fetch(
    `/api/moderation-rules?category=${encodeURIComponent(category)}`
  );
  const data = await response.json();
  return data.data;
}

// 获取单条规则
async function getRule(code: string) {
  const response = await fetch(`/api/moderation-rules/${code}`);
  const rule = await response.json();
  return rule;
}
```

---

## 性能优化

- 使用 `page_size` 参数获取适当数量的记录
- 使用精确筛选参数减少需要搜索的数据量
- 考虑缓存常用的分类和风险等级列表

---

## 变更日志

### v1.0 (2025-10-25)

- 初版发布
- 实现完整的规则库 API
- 支持筛选、搜索和分页
- 包含 42 条审核规则

---

## 联系与支持

如有问题或建议，请联系内容安全团队。
