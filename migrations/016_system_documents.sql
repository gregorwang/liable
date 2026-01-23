-- System documents and permissions
-- Migration: 016_system_documents
-- Description: Add editable system documents and permissions for editing docs and clearing AI review tasks

CREATE TABLE IF NOT EXISTS system_documents (
    key VARCHAR(50) PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by INTEGER REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_system_documents_updated_at ON system_documents(updated_at);

INSERT INTO system_documents (key, title, content)
VALUES
    (
        'system-usage',
        '系统使用说明',
        $$# 系统使用说明

## 角色与权限
- 管理员：管理用户、规则、标签、AI 审核批次与统计。
- 审核员：领取任务并提交审核结果。

## 登录与账号
- 支持用户名+密码登录，以及邮箱验证码登录。
- 新注册账号需要管理员审批后才可登录。

## 任务流程
1. 领取任务：从任务队列领取待审核内容。
2. 审核决策：选择通过/拒绝，并填写原因。
3. 违规标签：拒绝时请选择对应的违规标签；通过时无需填写标签。
4. 提交结果：系统记录审核结果并进入统计与质检流程。

## AI 审核批次
- 管理员可创建 AI 审核批次，对指定范围的评论进行自动审核。
- AI 结果会与人工一审结果对比，用于评估一致性与标签重叠率。
- AI 审核仅做辅助分析，不直接覆盖人工审核结果。

## 规则库
- 规则库用于描述违规类型、判定要点与处置动作。
- 规则条目以“规则编号 + 分类 + 风险等级”进行组织和检索。$$
    ),
    (
        'ai-confidence-scoring',
        '置信度打分说明',
        $$# 置信度打分说明

## 定义
置信度用于表达 AI 对审核结论的确定程度，范围为 0-100。

## 打分参考
- 50：非常不确定，证据不足或内容含糊。
- 60-70：偏不确定，需要人工重点关注。
- 80：较确定，判断依据相对充分。
- 90-95：高度确定，仅在明显违规或明显合规时使用。
- 96-100：极高确定度，极少出现。

## 使用原则
- 置信度不是“风险等级”，仅表示 AI 自信程度。
- 不建议默认输出固定数值（例如 85/95），应基于内容具体情况给分。
- 若 AI 判断为“通过”，通常应给出中等或偏低置信度，除非明显合规。

## 与标签的关系
- AI 判定为“拒绝”时，需要从系统配置的标签中选择 1-3 个。
- AI 判定为“通过”时，标签应为空。$$
    )
ON CONFLICT (key) DO NOTHING;

INSERT INTO permissions (permission_key, name, description, resource, action, category, is_active) VALUES
    ('docs:edit', '编辑系统文档', '允许编辑系统使用说明和置信度说明', 'system_docs', 'edit', 'configuration', true),
    ('ai-review:tasks:delete', '清空AI审核任务', '允许清空AI审核批次的任务与结果', 'ai_review', 'delete', 'ai_review', true)
ON CONFLICT (permission_key) DO NOTHING;

INSERT INTO user_permissions (user_id, permission_key, granted_by)
SELECT u.id, p.permission_key, u.id
FROM users u
CROSS JOIN (
    SELECT permission_key FROM permissions
    WHERE permission_key IN ('docs:edit', 'ai-review:tasks:delete')
) p
WHERE u.role = 'admin'
ON CONFLICT (user_id, permission_key) DO NOTHING;
