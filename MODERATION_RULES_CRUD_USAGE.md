# 审核规则CRUD操作使用指南

## 概述

审核规则库现在支持完整的CRUD（创建、读取、更新、删除）操作。所有审核规则都会显示在规则列表中，搜索和筛选功能用于快速定位特定规则。

## 前端功能

### 1. 查看所有规则

- 进入"审核规则"页面时，会加载并显示所有审核规则
- 规则会通过分页显示（每页20条，可调整）
- 支持按"规则编号"、"描述"、"分类"或"风险等级"搜索和筛选

### 2. 新增规则

**操作步骤：**
1. 点击页面右上角的"+ 新增规则"按钮
2. 在弹出的对话框中填写以下必填字段：
   - **规则编号**：如 A1, B2, C3（不可重复）
   - **分类**：如"人身安全与暴力"
   - **二级标签**：如"真实威胁/伤害"
   - **描述**：规则的简要描述
   - **风险等级**：L(低) / M(中) / H(高) / C(极高)

3. 可选字段：
   - **判定要点**：详细的判定标准
   - **处置动作**：应采取的处置动作
   - **边界说明**：规则的边界和限制条件
   - **示例**：具体的应用示例
   - **快捷标签**：可选的快捷标签

4. 点击"创建"按钮提交表单
5. 创建成功后，新规则会立即出现在规则列表中

### 3. 编辑规则

**操作步骤：**
1. 在规则列表中找到要编辑的规则
2. 点击该行右侧的"编辑"按钮
3. 在弹出的对话框中修改规则内容（规则编号不可修改）
4. 点击"更新"按钮保存更改
5. 更新成功后，列表会自动刷新显示新的内容

### 4. 删除规则

**操作步骤：**
1. 在规则列表中找到要删除的规则
2. 点击该行右侧的"删除"按钮
3. 在确认对话框中点击"确认"
4. 删除成功后，该规则会从列表中移除

### 5. 搜索和筛选

- **搜索框**：按规则编号或描述搜索（不区分大小写）
- **分类筛选**：按规则分类筛选（动态从已有规则加载）
- **风险等级筛选**：按风险等级（L/M/H/C）筛选

这些筛选条件可以组合使用，结果会实时更新。

## 后端API

所有操作都需要管理员权限（需要登录且角色为admin）。

### 1. 获取规则列表

```
GET /api/moderation-rules
```

**Query参数：**
- `page` (可选)：页码，默认1
- `page_size` (可选)：每页数量，默认20，最大100
- `category` (可选)：按分类筛选
- `risk_level` (可选)：按风险等级筛选（L/M/H/C）
- `search` (可选)：按规则编号或描述搜索

**响应示例：**
```json
{
  "data": [
    {
      "id": 1,
      "rule_code": "A1",
      "category": "人身安全与暴力",
      "subcategory": "真实威胁/伤害",
      "description": "对个人/群体发出信息援助...",
      "judgment_criteria": "判定要点内容",
      "risk_level": "C",
      "action": "处置动作",
      "boundary": "边界说明",
      "examples": "示例内容",
      "quick_tag": "标签",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 29,
  "page": 1,
  "page_size": 20,
  "total_pages": 2
}
```

### 2. 创建规则

```
POST /api/admin/moderation-rules
```

**需要管理员权限**

**请求体：**
```json
{
  "rule_code": "A1",
  "category": "人身安全与暴力",
  "subcategory": "真实威胁/伤害",
  "description": "对个人/群体发出信息援助...",
  "judgment_criteria": "判定要点内容",
  "risk_level": "C",
  "action": "处置动作",
  "boundary": "边界说明",
  "examples": "示例内容",
  "quick_tag": "标签"
}
```

**必填字段：**
- `rule_code`
- `category`
- `subcategory`
- `description`
- `risk_level`

**响应示例：**
```json
{
  "id": 1,
  "rule_code": "A1",
  "category": "人身安全与暴力",
  "subcategory": "真实威胁/伤害",
  "description": "对个人/群体发出信息援助...",
  "judgment_criteria": "判定要点内容",
  "risk_level": "C",
  "action": "处置动作",
  "boundary": "边界说明",
  "examples": "示例内容",
  "quick_tag": "标签",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 3. 更新规则

```
PUT /api/admin/moderation-rules/:id
```

**需要管理员权限**

**路径参数：**
- `id`：规则ID

**请求体：** 同创建规则

**响应示例：** 同创建规则

### 4. 删除规则

```
DELETE /api/admin/moderation-rules/:id
```

**需要管理员权限**

**路径参数：**
- `id`：规则ID

**响应示例：**
```json
{
  "message": "Rule deleted successfully"
}
```

## 使用示例

### 使用curl测试

#### 创建规则
```bash
curl -X POST http://localhost:8080/api/admin/moderation-rules \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "rule_code": "A1",
    "category": "人身安全与暴力",
    "subcategory": "真实威胁/伤害",
    "description": "对个人/群体发出信息援助，具体时间/地点: 对个人/群体发出可信威胁",
    "judgment_criteria": "对个人/人群发出可信威胁...",
    "risk_level": "C",
    "action": "发布或禁言",
    "boundary": "具体的威胁陈述或鼓励行动表达",
    "examples": "示例...",
    "quick_tag": "真实威胁"
  }'
```

#### 获取规则列表
```bash
curl -X GET "http://localhost:8080/api/moderation-rules?page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 按分类筛选
```bash
curl -X GET "http://localhost:8080/api/moderation-rules?category=人身安全与暴力" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 按风险等级筛选
```bash
curl -X GET "http://localhost:8080/api/moderation-rules?risk_level=C" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 搜索
```bash
curl -X GET "http://localhost:8080/api/moderation-rules?search=A1" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 更新规则
```bash
curl -X PUT http://localhost:8080/api/admin/moderation-rules/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "rule_code": "A1",
    "category": "人身安全与暴力",
    "subcategory": "真实威胁/伤害",
    "description": "更新后的描述",
    "judgment_criteria": "更新后的判定要点",
    "risk_level": "H",
    "action": "更新后的处置动作",
    "boundary": "更新后的边界",
    "examples": "更新后的示例",
    "quick_tag": "新标签"
  }'
```

#### 删除规则
```bash
curl -X DELETE http://localhost:8080/api/admin/moderation-rules/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 权限要求

| 操作 | 权限要求 |
|------|--------|
| 查看规则列表 | 无（公开） |
| 搜索和筛选规则 | 无（公开） |
| 创建规则 | 管理员 |
| 编辑规则 | 管理员 |
| 删除规则 | 管理员 |

## 错误处理

### 常见错误码

| 错误码 | 原因 | 解决方案 |
|-------|------|--------|
| 400 | 缺少必填字段或数据格式错误 | 检查请求体中的必填字段 |
| 401 | 未授权（缺少token或token过期） | 重新登录获取新token |
| 403 | 禁止访问（权限不足） | 只有管理员可以执行CRUD操作 |
| 404 | 资源未找到 | 检查规则ID是否存在 |
| 409 | 规则编号已存在 | 使用不同的规则编号 |
| 500 | 服务器错误 | 查看服务器日志 |

## 最佳实践

1. **规则编号命名**：遵循现有的命名规范（如A1, B2, C3等）
2. **风险等级分配**：
   - L(低)：较低风险，一般性建议
   - M(中)：中等风险，需要注意
   - H(高)：高风险，严重内容
   - C(极高)：极高风险，违法有害内容

3. **描述简洁**：保持描述简明扼要，便于理解
4. **示例完善**：提供清晰的示例，帮助审核员理解

5. **定期更新**：根据需要定期更新规则内容以适应新情况

## 故障排查

### 问题：创建/编辑规则时出现"操作失败"错误

**解决方案：**
1. 检查是否已登录且为管理员
2. 检查所有必填字段是否已填
3. 检查规则编号是否重复（创建时）
4. 查看浏览器控制台的错误信息

### 问题：删除规则后仍然出现在列表中

**解决方案：**
1. 刷新页面
2. 检查浏览器缓存
3. 确认删除操作是否已完成

### 问题：搜索或筛选功能不工作

**解决方案：**
1. 清除所有筛选条件重新尝试
2. 刷新页面
3. 检查浏览器控制台是否有错误信息
