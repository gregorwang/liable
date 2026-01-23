-- Update system usage document with shortcuts
-- Migration: 019_update_system_usage_doc
-- Description: Sync system-usage document content with latest shortcut guide

INSERT INTO system_documents (key, title, content, updated_at)
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

## 快捷键操作说明
说明：快捷键作用于当前高亮任务卡片（点击任务卡片可设为当前），输入框内不会触发快捷键。

通用（适用于评论审核/二审/差异/质检/视频审核）：
- Tab：切换到下一个任务
- Shift + Tab：切换到上一个任务
- Enter：提交当前任务
- Ctrl + Enter：批量提交
- Esc：清空当前任务表单
- R：刷新任务列表

评论一审（Reviewer Dashboard）：
- 1：通过
- 2：不通过
- Q：快速拒绝（预填常见拒绝原因）

评论二审（Second Review）：
- 1：通过
- 2：不通过

AI 人工差异（AI Human Diff）：
- 1：通过
- 2：不通过

质检（Quality Check）：
- 1：质检通过
- 2：质检不通过

视频审核（流量池）：
- 1：推送到下一流量池
- 2：自然流量池
- 3：违规下架

## AI 审核批次
- 管理员可创建 AI 审核批次，对指定范围的评论进行自动审核。
- AI 结果会与人工一审结果对比，用于评估一致性与标签重叠率。
- AI 审核仅做辅助分析，不直接覆盖人工审核结果。

## 规则库
- 规则库用于描述违规类型、判定要点与处置动作。
- 规则条目以“规则编号 + 分类 + 风险等级”进行组织和检索。$$,
        NOW()
    )
ON CONFLICT (key) DO UPDATE SET
    title = EXCLUDED.title,
    content = EXCLUDED.content,
    updated_at = NOW();
