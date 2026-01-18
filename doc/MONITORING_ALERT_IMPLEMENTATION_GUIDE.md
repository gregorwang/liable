# 监控告警系统实现指南

## 概述

本文档详细说明如何基于 Supabase 实现评论审核平台的监控告警系统，包括：
1. **队列积压告警** - 监控各审核队列任务堆积情况
2. **审核速率监控** - 追踪审核员工作效率和系统吞吐量
3. **流转异常检测** - 识别任务流转中的异常情况

所有功能基于 Supabase 的原生能力实现，无需修改现有业务代码。

---

## 一、系统架构

### 1.1 技术栈

- **PostgreSQL** - 核心数据存储和计算引擎
- **pg_cron** - PostgreSQL 定时任务扩展
- **Supabase Edge Functions** - 告警通知发送
- **Supabase Realtime** - 实时数据推送（可选）
- **通知表** - 已有的 `notifications` 表用于存储告警记录

### 1.2 监控数据流

```
定时任务(pg_cron)
    ↓
监控函数执行
    ↓
检测异常条件
    ↓
记录告警日志 + 发送通知
    ↓
Edge Function / Webhook
```

---

## 二、数据库设计

### 2.1 告警配置表

创建告警配置表用于管理各类监控规则：

```sql
-- Migration: 007_monitoring_alert_system.sql

-- 1. 告警配置表
CREATE TABLE IF NOT EXISTS alert_config (
    id SERIAL PRIMARY KEY,
    alert_type VARCHAR(50) NOT NULL UNIQUE,
    alert_name VARCHAR(100) NOT NULL,
    description TEXT,

    -- 告警条件配置（JSONB存储灵活配置）
    conditions JSONB NOT NULL DEFAULT '{}',
    -- 示例: {
    --   "threshold": 100,           -- 阈值
    --   "time_window_minutes": 30,  -- 时间窗口
    --   "severity": "high"          -- 严重级别
    -- }

    -- 告警级别
    severity VARCHAR(20) NOT NULL DEFAULT 'medium'
        CHECK (severity IN ('low', 'medium', 'high', 'critical')),

    -- 通知配置
    notification_config JSONB DEFAULT '{}',
    -- 示例: {
    --   "channels": ["email", "webhook"],
    --   "recipients": ["admin@example.com"],
    --   "webhook_url": "https://..."
    -- }

    -- 启用状态
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    -- 冷却时间（分钟）- 避免重复告警
    cooldown_minutes INTEGER DEFAULT 60,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2. 告警历史表
CREATE TABLE IF NOT EXISTS alert_history (
    id SERIAL PRIMARY KEY,
    alert_config_id INTEGER REFERENCES alert_config(id),
    alert_type VARCHAR(50) NOT NULL,

    -- 告警详情
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL,

    -- 触发数据（快照）
    trigger_data JSONB DEFAULT '{}',

    -- 状态
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'resolved', 'acknowledged')),

    -- 通知状态
    notification_sent BOOLEAN DEFAULT FALSE,
    notification_sent_at TIMESTAMP,
    notification_error TEXT,

    -- 解决信息
    resolved_at TIMESTAMP,
    resolved_by INTEGER REFERENCES users(id),
    resolution_notes TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 3. 监控指标表（用于存储历史指标数据）
CREATE TABLE IF NOT EXISTS monitoring_metrics (
    id SERIAL PRIMARY KEY,
    metric_type VARCHAR(50) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,

    -- 指标值
    metric_value NUMERIC NOT NULL,
    metric_unit VARCHAR(20), -- 'count', 'percentage', 'minutes', etc.

    -- 关联信息
    queue_name VARCHAR(50),
    pool VARCHAR(10),

    -- 详细数据
    details JSONB DEFAULT '{}',

    recorded_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 创建索引
CREATE INDEX idx_alert_config_type ON alert_config(alert_type);
CREATE INDEX idx_alert_config_active ON alert_config(is_active);
CREATE INDEX idx_alert_history_type ON alert_history(alert_type);
CREATE INDEX idx_alert_history_status ON alert_history(status);
CREATE INDEX idx_alert_history_created ON alert_history(created_at DESC);
CREATE INDEX idx_monitoring_metrics_type ON monitoring_metrics(metric_type);
CREATE INDEX idx_monitoring_metrics_recorded ON monitoring_metrics(recorded_at DESC);
CREATE INDEX idx_monitoring_metrics_queue ON monitoring_metrics(queue_name);
```

### 2.2 初始化告警配置

```sql
-- 插入默认告警配置
INSERT INTO alert_config (alert_type, alert_name, description, conditions, severity, notification_config) VALUES

-- 队列积压告警
('queue_backlog', '队列积压告警', '当待处理任务数超过阈值时触发',
    '{"threshold": 100, "time_window_minutes": 30, "check_queues": ["comment_first_review", "comment_second_review", "video_first_review", "video_second_review"]}',
    'high',
    '{"channels": ["notification", "webhook"], "recipients": ["admin"]}'),

('queue_stagnation', '队列停滞告警', '当队列在指定时间内没有任务完成时触发',
    '{"time_window_minutes": 60, "min_completion_expected": 5}',
    'medium',
    '{"channels": ["notification"]}'),

-- 审核速率监控
('review_rate_low', '审核速率过低', '当审核速率低于预期时触发',
    '{"min_rate_per_hour": 10, "time_window_minutes": 60}',
    'medium',
    '{"channels": ["notification"]}'),

('review_rate_declining', '审核速率下降', '当审核速率持续下降时触发',
    '{"decline_percentage": 30, "time_window_minutes": 120}',
    'low',
    '{"channels": ["notification"]}'),

('reviewer_idle', '审核员空闲告警', '当有待处理任务但审核员空闲时触发',
    '{"idle_minutes": 30, "min_pending_tasks": 10}',
    'medium',
    '{"channels": ["notification"]}'),

-- 流转异常检测
('task_timeout', '任务超时告警', '任务领取后超过预期时间未完成',
    '{"timeout_minutes": 120}',
    'high',
    '{"channels": ["notification", "webhook"]}'),

('task_claiming_failure', '任务领取失败', '任务反复被领取又归还',
    '{"max_claim_count": 3, "time_window_hours": 24}',
    'high',
    '{"channels": ["notification"]}'),

('abnormal_rejection_rate', '异常拒绝率', '审核拒绝率异常高',
    '{"rejection_rate_threshold": 0.8, "min_sample_size": 20, "time_window_hours": 24}',
    'medium',
    '{"channels": ["notification"]}'),

('status_transition_error', '状态流转错误', '检测到非法状态转换',
    '{"check_transitions": true}',
    'critical',
    '{"channels": ["notification", "webhook"]}')

ON CONFLICT (alert_type) DO NOTHING;
```

---

## 三、监控函数实现

### 3.1 队列积压监控

```sql
-- 检测队列积压
CREATE OR REPLACE FUNCTION check_queue_backlog()
RETURNS TABLE (
    alert_triggered BOOLEAN,
    queue_name TEXT,
    pending_count INTEGER,
    threshold INTEGER
) AS $$
DECLARE
    config RECORD;
    queue RECORD;
    last_alert TIMESTAMP;
BEGIN
    -- 获取配置
    SELECT * INTO config FROM alert_config
    WHERE alert_type = 'queue_backlog' AND is_active = TRUE
    LIMIT 1;

    IF NOT FOUND THEN
        RETURN;
    END IF;

    -- 检查每个队列
    FOR queue IN
        SELECT
            uqs.queue_name,
            uqs.pending_tasks
        FROM unified_queue_stats uqs
        WHERE uqs.queue_name = ANY((config.conditions->>'check_queues')::TEXT[])
    LOOP
        IF queue.pending_tasks > (config.conditions->>'threshold')::INTEGER THEN
            -- 检查冷却时间
            SELECT MAX(created_at) INTO last_alert
            FROM alert_history
            WHERE alert_type = 'queue_backlog'
                AND trigger_data->>'queue_name' = queue.queue_name
                AND status = 'active'
                AND created_at > NOW() - (config.cooldown_minutes || ' minutes')::INTERVAL;

            IF last_alert IS NULL THEN
                -- 记录告警
                INSERT INTO alert_history (
                    alert_config_id,
                    alert_type,
                    title,
                    message,
                    severity,
                    trigger_data
                ) VALUES (
                    config.id,
                    'queue_backlog',
                    '队列积压告警: ' || queue.queue_name,
                    format('队列 %s 待处理任务数 %s 超过阈值 %s',
                        queue.queue_name,
                        queue.pending_tasks,
                        config.conditions->>'threshold'),
                    config.severity,
                    jsonb_build_object(
                        'queue_name', queue.queue_name,
                        'pending_count', queue.pending_tasks,
                        'threshold', config.conditions->>'threshold'
                    )
                );

                RETURN QUERY SELECT
                    TRUE,
                    queue.queue_name,
                    queue.pending_tasks::INTEGER,
                    (config.conditions->>'threshold')::INTEGER;
            END IF;
        END IF;
    END LOOP;

    RETURN;
END;
$$ LANGUAGE plpgsql;

-- 检测队列停滞
CREATE OR REPLACE FUNCTION check_queue_stagnation()
RETURNS TABLE (
    alert_triggered BOOLEAN,
    queue_name TEXT,
    last_completion TIMESTAMP,
    stagnation_minutes INTEGER
) AS $$
DECLARE
    config RECORD;
    queue RECORD;
    last_completion_time TIMESTAMP;
    stagnation_duration INTEGER;
BEGIN
    SELECT * INTO config FROM alert_config
    WHERE alert_type = 'queue_stagnation' AND is_active = TRUE
    LIMIT 1;

    IF NOT FOUND THEN
        RETURN;
    END IF;

    -- 检查各队列的最后完成时间
    FOR queue IN
        SELECT queue_name FROM (
            SELECT 'comment_first_review' as queue_name
            UNION SELECT 'comment_second_review'
            UNION SELECT 'video_first_review'
            UNION SELECT 'video_second_review'
            UNION SELECT 'quality_check'
        ) q
    LOOP
        -- 获取最后完成时间（根据队列类型查询对应表）
        EXECUTE format('
            SELECT MAX(completed_at)
            FROM %I
            WHERE status = ''completed''
        ', queue.queue_name || '_tasks')
        INTO last_completion_time;

        IF last_completion_time IS NOT NULL THEN
            stagnation_duration := EXTRACT(EPOCH FROM (NOW() - last_completion_time)) / 60;

            IF stagnation_duration > (config.conditions->>'time_window_minutes')::INTEGER THEN
                -- 记录告警
                INSERT INTO alert_history (
                    alert_config_id,
                    alert_type,
                    title,
                    message,
                    severity,
                    trigger_data
                ) VALUES (
                    config.id,
                    'queue_stagnation',
                    '队列停滞告警: ' || queue.queue_name,
                    format('队列 %s 已经 %s 分钟没有完成任务', queue.queue_name, stagnation_duration),
                    config.severity,
                    jsonb_build_object(
                        'queue_name', queue.queue_name,
                        'last_completion', last_completion_time,
                        'stagnation_minutes', stagnation_duration
                    )
                );

                RETURN QUERY SELECT
                    TRUE,
                    queue.queue_name,
                    last_completion_time,
                    stagnation_duration;
            END IF;
        END IF;
    END LOOP;

    RETURN;
END;
$$ LANGUAGE plpgsql;
```

### 3.2 审核速率监控

```sql
-- 计算审核速率
CREATE OR REPLACE FUNCTION calculate_review_rate(
    p_queue_name TEXT,
    p_time_window_minutes INTEGER
)
RETURNS TABLE (
    queue_name TEXT,
    completed_count INTEGER,
    rate_per_hour NUMERIC,
    avg_process_time_minutes NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    EXECUTE format('
        SELECT
            %L::TEXT as queue_name,
            COUNT(*)::INTEGER as completed_count,
            (COUNT(*) * 60.0 / %s)::NUMERIC as rate_per_hour,
            AVG(EXTRACT(EPOCH FROM (completed_at - claimed_at)) / 60.0)::NUMERIC as avg_process_time_minutes
        FROM %I
        WHERE status = ''completed''
            AND completed_at > NOW() - interval ''%s minutes''
    ', p_queue_name, p_time_window_minutes, p_queue_name || '_tasks', p_time_window_minutes);
END;
$$ LANGUAGE plpgsql;

-- 检测审核速率过低
CREATE OR REPLACE FUNCTION check_review_rate_low()
RETURNS TABLE (
    alert_triggered BOOLEAN,
    queue_name TEXT,
    current_rate NUMERIC,
    expected_rate NUMERIC
) AS $$
DECLARE
    config RECORD;
    rate_data RECORD;
BEGIN
    SELECT * INTO config FROM alert_config
    WHERE alert_type = 'review_rate_low' AND is_active = TRUE
    LIMIT 1;

    IF NOT FOUND THEN
        RETURN;
    END IF;

    -- 检查各队列的审核速率
    FOR rate_data IN
        SELECT * FROM calculate_review_rate(
            'comment_first_review',
            (config.conditions->>'time_window_minutes')::INTEGER
        )
        UNION ALL
        SELECT * FROM calculate_review_rate(
            'comment_second_review',
            (config.conditions->>'time_window_minutes')::INTEGER
        )
        UNION ALL
        SELECT * FROM calculate_review_rate(
            'video_first_review',
            (config.conditions->>'time_window_minutes')::INTEGER
        )
        UNION ALL
        SELECT * FROM calculate_review_rate(
            'video_second_review',
            (config.conditions->>'time_window_minutes')::INTEGER
        )
    LOOP
        IF rate_data.rate_per_hour < (config.conditions->>'min_rate_per_hour')::NUMERIC THEN
            -- 记录告警
            INSERT INTO alert_history (
                alert_config_id,
                alert_type,
                title,
                message,
                severity,
                trigger_data
            ) VALUES (
                config.id,
                'review_rate_low',
                '审核速率过低: ' || rate_data.queue_name,
                format('队列 %s 当前审核速率 %.2f 任务/小时，低于预期 %s 任务/小时',
                    rate_data.queue_name,
                    rate_data.rate_per_hour,
                    config.conditions->>'min_rate_per_hour'),
                config.severity,
                jsonb_build_object(
                    'queue_name', rate_data.queue_name,
                    'current_rate', rate_data.rate_per_hour,
                    'expected_rate', config.conditions->>'min_rate_per_hour',
                    'completed_count', rate_data.completed_count
                )
            );

            RETURN QUERY SELECT
                TRUE,
                rate_data.queue_name,
                rate_data.rate_per_hour,
                (config.conditions->>'min_rate_per_hour')::NUMERIC;
        END IF;

        -- 记录监控指标
        INSERT INTO monitoring_metrics (
            metric_type,
            metric_name,
            metric_value,
            metric_unit,
            queue_name,
            details
        ) VALUES (
            'review_rate',
            'tasks_per_hour',
            rate_data.rate_per_hour,
            'tasks/hour',
            rate_data.queue_name,
            jsonb_build_object(
                'completed_count', rate_data.completed_count,
                'avg_process_time', rate_data.avg_process_time_minutes
            )
        );
    END LOOP;

    RETURN;
END;
$$ LANGUAGE plpgsql;

-- 检测审核员空闲
CREATE OR REPLACE FUNCTION check_reviewer_idle()
RETURNS TABLE (
    alert_triggered BOOLEAN,
    idle_reviewer_count INTEGER,
    pending_task_count INTEGER
) AS $$
DECLARE
    config RECORD;
    idle_count INTEGER;
    pending_count INTEGER;
BEGIN
    SELECT * INTO config FROM alert_config
    WHERE alert_type = 'reviewer_idle' AND is_active = TRUE
    LIMIT 1;

    IF NOT FOUND THEN
        RETURN;
    END IF;

    -- 获取待处理任务总数
    SELECT SUM(pending_tasks)::INTEGER INTO pending_count
    FROM unified_queue_stats;

    IF pending_count < (config.conditions->>'min_pending_tasks')::INTEGER THEN
        RETURN;
    END IF;

    -- 统计空闲审核员（有权限但没有进行中任务的用户）
    WITH active_reviewers AS (
        SELECT DISTINCT reviewer_id
        FROM review_tasks
        WHERE status = 'in_progress'
        UNION
        SELECT DISTINCT reviewer_id
        FROM second_review_tasks
        WHERE status = 'in_progress'
        UNION
        SELECT DISTINCT reviewer_id
        FROM video_first_review_tasks
        WHERE status = 'in_progress'
        UNION
        SELECT DISTINCT reviewer_id
        FROM video_second_review_tasks
        WHERE status = 'in_progress'
    ),
    potential_reviewers AS (
        SELECT DISTINCT u.id
        FROM users u
        JOIN user_role_assignments ura ON u.id = ura.user_id
        JOIN roles r ON ura.role_id = r.id
        WHERE r.name IN ('reviewer', 'quality_checker', 'admin')
            AND u.is_active = TRUE
    )
    SELECT COUNT(*)::INTEGER INTO idle_count
    FROM potential_reviewers pr
    WHERE pr.id NOT IN (SELECT reviewer_id FROM active_reviewers WHERE reviewer_id IS NOT NULL);

    IF idle_count > 0 THEN
        INSERT INTO alert_history (
            alert_config_id,
            alert_type,
            title,
            message,
            severity,
            trigger_data
        ) VALUES (
            config.id,
            'reviewer_idle',
            '审核员空闲告警',
            format('当前有 %s 个审核员空闲，但有 %s 个待处理任务', idle_count, pending_count),
            config.severity,
            jsonb_build_object(
                'idle_reviewer_count', idle_count,
                'pending_task_count', pending_count
            )
        );

        RETURN QUERY SELECT TRUE, idle_count, pending_count;
    END IF;

    RETURN;
END;
$$ LANGUAGE plpgsql;
```

### 3.3 流转异常检测

```sql
-- 检测任务超时
CREATE OR REPLACE FUNCTION check_task_timeout()
RETURNS TABLE (
    alert_triggered BOOLEAN,
    queue_name TEXT,
    timeout_task_count INTEGER
) AS $$
DECLARE
    config RECORD;
    timeout_count INTEGER;
    queue TEXT;
BEGIN
    SELECT * INTO config FROM alert_config
    WHERE alert_type = 'task_timeout' AND is_active = TRUE
    LIMIT 1;

    IF NOT FOUND THEN
        RETURN;
    END IF;

    -- 检查各队列的超时任务
    FOR queue IN
        SELECT unnest(ARRAY[
            'review_tasks',
            'second_review_tasks',
            'quality_check_tasks',
            'video_first_review_tasks',
            'video_second_review_tasks'
        ])
    LOOP
        EXECUTE format('
            SELECT COUNT(*)::INTEGER
            FROM %I
            WHERE status = ''in_progress''
                AND claimed_at IS NOT NULL
                AND claimed_at < NOW() - interval ''%s minutes''
        ', queue, config.conditions->>'timeout_minutes')
        INTO timeout_count;

        IF timeout_count > 0 THEN
            INSERT INTO alert_history (
                alert_config_id,
                alert_type,
                title,
                message,
                severity,
                trigger_data
            ) VALUES (
                config.id,
                'task_timeout',
                '任务超时告警: ' || queue,
                format('队列 %s 有 %s 个任务超过 %s 分钟未完成',
                    queue,
                    timeout_count,
                    config.conditions->>'timeout_minutes'),
                config.severity,
                jsonb_build_object(
                    'queue_name', queue,
                    'timeout_count', timeout_count,
                    'timeout_threshold_minutes', config.conditions->>'timeout_minutes'
                )
            );

            RETURN QUERY SELECT TRUE, queue, timeout_count;
        END IF;
    END LOOP;

    RETURN;
END;
$$ LANGUAGE plpgsql;

-- 检测异常拒绝率
CREATE OR REPLACE FUNCTION check_abnormal_rejection_rate()
RETURNS TABLE (
    alert_triggered BOOLEAN,
    queue_name TEXT,
    rejection_rate NUMERIC,
    sample_size INTEGER
) AS $$
DECLARE
    config RECORD;
    stats RECORD;
BEGIN
    SELECT * INTO config FROM alert_config
    WHERE alert_type = 'abnormal_rejection_rate' AND is_active = TRUE
    LIMIT 1;

    IF NOT FOUND THEN
        RETURN;
    END IF;

    -- 检查评论一审拒绝率
    SELECT
        'comment_first_review' as queue,
        COUNT(*) as total,
        COUNT(*) FILTER (WHERE is_approved = FALSE) as rejected,
        CASE
            WHEN COUNT(*) > 0
            THEN (COUNT(*) FILTER (WHERE is_approved = FALSE))::NUMERIC / COUNT(*)
            ELSE 0
        END as rejection_rate
    INTO stats
    FROM review_results
    WHERE created_at > NOW() - interval '24 hours';

    IF stats.total >= (config.conditions->>'min_sample_size')::INTEGER
        AND stats.rejection_rate > (config.conditions->>'rejection_rate_threshold')::NUMERIC THEN

        INSERT INTO alert_history (
            alert_config_id,
            alert_type,
            title,
            message,
            severity,
            trigger_data
        ) VALUES (
            config.id,
            'abnormal_rejection_rate',
            '异常拒绝率: ' || stats.queue,
            format('队列 %s 拒绝率 %.2f%% 超过阈值 %.2f%%（样本量: %s）',
                stats.queue,
                stats.rejection_rate * 100,
                (config.conditions->>'rejection_rate_threshold')::NUMERIC * 100,
                stats.total),
            config.severity,
            jsonb_build_object(
                'queue_name', stats.queue,
                'rejection_rate', stats.rejection_rate,
                'sample_size', stats.total,
                'rejected_count', stats.rejected
            )
        );

        RETURN QUERY SELECT TRUE, stats.queue, stats.rejection_rate, stats.total;
    END IF;

    RETURN;
END;
$$ LANGUAGE plpgsql;

-- 主监控调度函数
CREATE OR REPLACE FUNCTION run_monitoring_checks()
RETURNS JSONB AS $$
DECLARE
    result JSONB;
    alerts_triggered INTEGER := 0;
BEGIN
    result := jsonb_build_object('timestamp', NOW(), 'checks', jsonb_build_array());

    -- 执行各项检查
    PERFORM check_queue_backlog();
    PERFORM check_queue_stagnation();
    PERFORM check_review_rate_low();
    PERFORM check_reviewer_idle();
    PERFORM check_task_timeout();
    PERFORM check_abnormal_rejection_rate();

    -- 统计新触发的告警数
    SELECT COUNT(*) INTO alerts_triggered
    FROM alert_history
    WHERE created_at > NOW() - interval '5 minutes';

    result := result || jsonb_build_object(
        'alerts_triggered', alerts_triggered,
        'status', 'completed'
    );

    RETURN result;
END;
$$ LANGUAGE plpgsql;
```

---

## 四、定时任务配置

### 4.1 启用 pg_cron 扩展

```sql
-- 在 Supabase Dashboard 的 SQL Editor 中执行
CREATE EXTENSION IF NOT EXISTS pg_cron;
```

### 4.2 配置定时监控任务

```sql
-- 每5分钟执行一次监控检查
SELECT cron.schedule(
    'monitoring-checks',           -- 任务名称
    '*/5 * * * *',                 -- 每5分钟
    $$SELECT run_monitoring_checks()$$
);

-- 每小时记录一次系统指标
SELECT cron.schedule(
    'hourly-metrics-collection',
    '0 * * * *',                   -- 每小时整点
    $$
    INSERT INTO monitoring_metrics (metric_type, metric_name, metric_value, metric_unit, details)
    SELECT
        'system_health' as metric_type,
        'total_pending_tasks' as metric_name,
        SUM(pending_tasks) as metric_value,
        'count' as metric_unit,
        jsonb_build_object(
            'by_queue', jsonb_agg(
                jsonb_build_object(
                    'queue', queue_name,
                    'pending', pending_tasks
                )
            )
        ) as details
    FROM unified_queue_stats;
    $$
);

-- 每天清理30天前的告警历史（可选）
SELECT cron.schedule(
    'cleanup-old-alerts',
    '0 2 * * *',                   -- 每天凌晨2点
    $$
    DELETE FROM alert_history
    WHERE created_at < NOW() - interval '30 days'
        AND status = 'resolved';
    $$
);

-- 查看已配置的定时任务
SELECT * FROM cron.job;

-- 查看任务执行历史
SELECT * FROM cron.job_run_details ORDER BY start_time DESC LIMIT 10;
```

---

## 五、通知发送机制

### 5.1 创建 Edge Function 发送告警

创建 Supabase Edge Function 处理告警通知：

**文件路径**: `supabase/functions/send-alert-notification/index.ts`

```typescript
import { serve } from 'https://deno.land/std@0.168.0/http/server.ts'
import { createClient } from 'https://esm.sh/@supabase/supabase-js@2'

interface AlertPayload {
  alert_id: number
  alert_type: string
  title: string
  message: string
  severity: string
  trigger_data: any
}

serve(async (req) => {
  try {
    const supabase = createClient(
      Deno.env.get('SUPABASE_URL') ?? '',
      Deno.env.get('SUPABASE_SERVICE_ROLE_KEY') ?? ''
    )

    const payload: AlertPayload = await req.json()

    // 1. 获取告警配置
    const { data: config } = await supabase
      .from('alert_config')
      .select('notification_config')
      .eq('alert_type', payload.alert_type)
      .single()

    if (!config) {
      return new Response(JSON.stringify({ error: 'Config not found' }), {
        status: 404,
      })
    }

    const notificationConfig = config.notification_config

    // 2. 发送系统内通知
    if (notificationConfig.channels?.includes('notification')) {
      const recipients = notificationConfig.recipients || ['admin']

      // 获取管理员用户ID
      const { data: adminUsers } = await supabase
        .from('users')
        .select('id')
        .eq('role', 'admin')
        .eq('is_active', true)

      const notifications = adminUsers?.map((user: any) => ({
        user_id: user.id,
        type: 'alert',
        title: payload.title,
        message: payload.message,
        priority: payload.severity === 'critical' ? 'high' : 'normal',
        metadata: {
          alert_id: payload.alert_id,
          alert_type: payload.alert_type,
          trigger_data: payload.trigger_data,
        },
      }))

      if (notifications && notifications.length > 0) {
        await supabase.from('notifications').insert(notifications)
      }
    }

    // 3. 发送 Webhook
    if (notificationConfig.channels?.includes('webhook') && notificationConfig.webhook_url) {
      await fetch(notificationConfig.webhook_url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          event: 'alert_triggered',
          timestamp: new Date().toISOString(),
          alert: payload,
        }),
      })
    }

    // 4. 发送邮件（如果配置了）
    if (notificationConfig.channels?.includes('email')) {
      // 集成你的邮件服务（如 SendGrid, AWS SES 等）
      // await sendEmail(payload)
    }

    // 5. 更新告警历史状态
    await supabase
      .from('alert_history')
      .update({
        notification_sent: true,
        notification_sent_at: new Date().toISOString(),
      })
      .eq('id', payload.alert_id)

    return new Response(
      JSON.stringify({ success: true, message: 'Notification sent' }),
      { status: 200 }
    )
  } catch (error) {
    console.error('Error sending notification:', error)
    return new Response(JSON.stringify({ error: error.message }), {
      status: 500,
    })
  }
})
```

### 5.2 创建数据库触发器自动调用 Edge Function

```sql
-- 创建触发器函数
CREATE OR REPLACE FUNCTION notify_alert_created()
RETURNS TRIGGER AS $$
DECLARE
    config RECORD;
BEGIN
    -- 获取告警配置
    SELECT * INTO config FROM alert_config
    WHERE alert_type = NEW.alert_type AND is_active = TRUE;

    IF FOUND THEN
        -- 调用 Edge Function（使用 pg_net 扩展）
        PERFORM
            net.http_post(
                url := current_setting('app.settings.edge_function_url') || '/send-alert-notification',
                headers := jsonb_build_object(
                    'Content-Type', 'application/json',
                    'Authorization', 'Bearer ' || current_setting('app.settings.service_role_key')
                ),
                body := jsonb_build_object(
                    'alert_id', NEW.id,
                    'alert_type', NEW.alert_type,
                    'title', NEW.title,
                    'message', NEW.message,
                    'severity', NEW.severity,
                    'trigger_data', NEW.trigger_data
                )
            );
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器
CREATE TRIGGER trigger_alert_notification
    AFTER INSERT ON alert_history
    FOR EACH ROW
    WHEN (NEW.status = 'active')
    EXECUTE FUNCTION notify_alert_created();
```

如果不使用 Edge Function，也可以直接在触发器中插入通知：

```sql
CREATE OR REPLACE FUNCTION notify_alert_created_simple()
RETURNS TRIGGER AS $$
BEGIN
    -- 直接插入系统通知
    INSERT INTO notifications (user_id, type, title, message, priority, metadata)
    SELECT
        u.id,
        'alert',
        NEW.title,
        NEW.message,
        CASE NEW.severity
            WHEN 'critical' THEN 'high'
            WHEN 'high' THEN 'high'
            ELSE 'normal'
        END,
        jsonb_build_object(
            'alert_id', NEW.id,
            'alert_type', NEW.alert_type,
            'trigger_data', NEW.trigger_data
        )
    FROM users u
    JOIN user_role_assignments ura ON u.id = ura.user_id
    JOIN roles r ON ura.role_id = r.id
    WHERE r.name = 'admin' AND u.is_active = TRUE;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_alert_notification_simple
    AFTER INSERT ON alert_history
    FOR EACH ROW
    WHEN (NEW.status = 'active')
    EXECUTE FUNCTION notify_alert_created_simple();
```

---

## 六、监控仪表板

### 6.1 告警查询 API

创建视图简化前端查询：

```sql
-- 活跃告警视图
CREATE OR REPLACE VIEW active_alerts AS
SELECT
    ah.id,
    ah.alert_type,
    ah.title,
    ah.message,
    ah.severity,
    ah.trigger_data,
    ah.status,
    ah.created_at,
    ac.alert_name,
    ac.description
FROM alert_history ah
LEFT JOIN alert_config ac ON ah.alert_config_id = ac.id
WHERE ah.status = 'active'
ORDER BY
    CASE ah.severity
        WHEN 'critical' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        WHEN 'low' THEN 4
    END,
    ah.created_at DESC;

-- 监控指标聚合视图
CREATE OR REPLACE VIEW monitoring_dashboard_metrics AS
SELECT
    metric_type,
    queue_name,
    AVG(metric_value) as avg_value,
    MAX(metric_value) as max_value,
    MIN(metric_value) as min_value,
    COUNT(*) as sample_count,
    MAX(recorded_at) as last_recorded
FROM monitoring_metrics
WHERE recorded_at > NOW() - interval '24 hours'
GROUP BY metric_type, queue_name;

-- 告警趋势统计
CREATE OR REPLACE VIEW alert_trend_stats AS
SELECT
    DATE_TRUNC('hour', created_at) as hour,
    alert_type,
    severity,
    COUNT(*) as alert_count
FROM alert_history
WHERE created_at > NOW() - interval '7 days'
GROUP BY DATE_TRUNC('hour', created_at), alert_type, severity
ORDER BY hour DESC;
```

### 6.2 前端集成示例

在你的 Vue 前端中添加监控页面：

```typescript
// api/monitoring.ts
import { supabase } from '@/config/supabase'

export interface Alert {
  id: number
  alert_type: string
  title: string
  message: string
  severity: string
  trigger_data: any
  status: string
  created_at: string
}

export interface Metric {
  metric_type: string
  queue_name: string
  avg_value: number
  max_value: number
  min_value: number
  sample_count: number
  last_recorded: string
}

// 获取活跃告警
export async function getActiveAlerts(): Promise<Alert[]> {
  const { data, error } = await supabase
    .from('active_alerts')
    .select('*')
    .order('created_at', { ascending: false })
    .limit(50)

  if (error) throw error
  return data || []
}

// 获取告警历史
export async function getAlertHistory(
  startDate: string,
  endDate: string,
  alertType?: string
): Promise<Alert[]> {
  let query = supabase
    .from('alert_history')
    .select('*')
    .gte('created_at', startDate)
    .lte('created_at', endDate)

  if (alertType) {
    query = query.eq('alert_type', alertType)
  }

  const { data, error } = await query.order('created_at', { ascending: false })

  if (error) throw error
  return data || []
}

// 确认告警
export async function acknowledgeAlert(alertId: number, userId: number) {
  const { error } = await supabase
    .from('alert_history')
    .update({ status: 'acknowledged' })
    .eq('id', alertId)

  if (error) throw error
}

// 解决告警
export async function resolveAlert(
  alertId: number,
  userId: number,
  notes: string
) {
  const { error } = await supabase
    .from('alert_history')
    .update({
      status: 'resolved',
      resolved_at: new Date().toISOString(),
      resolved_by: userId,
      resolution_notes: notes,
    })
    .eq('id', alertId)

  if (error) throw error
}

// 获取监控指标
export async function getMonitoringMetrics(): Promise<Metric[]> {
  const { data, error } = await supabase
    .from('monitoring_dashboard_metrics')
    .select('*')

  if (error) throw error
  return data || []
}

// 订阅实时告警
export function subscribeToAlerts(
  callback: (alert: Alert) => void
) {
  return supabase
    .channel('alert_updates')
    .on(
      'postgres_changes',
      {
        event: 'INSERT',
        schema: 'public',
        table: 'alert_history',
      },
      (payload) => {
        callback(payload.new as Alert)
      }
    )
    .subscribe()
}
```

---

## 七、部署清单

### 7.1 数据库迁移

```bash
# 1. 创建迁移文件
# migrations/007_monitoring_alert_system.sql

# 2. 在 Supabase Dashboard 或使用 CLI 执行迁移
supabase db push

# 或在 SQL Editor 中直接运行迁移文件内容
```

### 7.2 启用扩展

在 Supabase Dashboard 的 Database > Extensions 中启用：
- `pg_cron` - 定时任务
- `pg_net` - HTTP 请求（可选，用于调用 Edge Functions）

### 7.3 配置环境变量

如需使用触发器调用 Edge Functions，需配置：

```sql
-- 设置项目配置（需要超级用户权限）
ALTER DATABASE postgres SET app.settings.edge_function_url = 'https://your-project.supabase.co/functions/v1';
ALTER DATABASE postgres SET app.settings.service_role_key = 'your-service-role-key';
```

### 7.4 部署 Edge Function（可选）

```bash
# 1. 安装 Supabase CLI
npm install -g supabase

# 2. 登录
supabase login

# 3. 链接项目
supabase link --project-ref your-project-ref

# 4. 部署函数
supabase functions deploy send-alert-notification
```

### 7.5 配置定时任务

执行第四节中的 pg_cron 配置 SQL。

---

## 八、测试验证

### 8.1 手动触发监控检查

```sql
-- 执行所有监控检查
SELECT run_monitoring_checks();

-- 单独测试各项检查
SELECT * FROM check_queue_backlog();
SELECT * FROM check_queue_stagnation();
SELECT * FROM check_review_rate_low();
SELECT * FROM check_task_timeout();

-- 查看生成的告警
SELECT * FROM alert_history ORDER BY created_at DESC LIMIT 10;

-- 查看监控指标
SELECT * FROM monitoring_metrics ORDER BY recorded_at DESC LIMIT 20;
```

### 8.2 模拟告警场景

```sql
-- 模拟队列积压（插入大量待处理任务）
INSERT INTO review_tasks (comment_id, status)
SELECT id, 'pending'
FROM comments
LIMIT 150;

-- 执行检查应该触发告警
SELECT * FROM check_queue_backlog();

-- 查看告警是否生成
SELECT * FROM alert_history WHERE alert_type = 'queue_backlog' ORDER BY created_at DESC;
```

### 8.3 验证通知发送

```sql
-- 检查是否生成了通知记录
SELECT * FROM notifications
WHERE type = 'alert'
ORDER BY created_at DESC
LIMIT 10;

-- 检查 Edge Function 调用日志（如果使用）
SELECT * FROM net._http_response ORDER BY created_at DESC LIMIT 10;
```

---

## 九、运维管理

### 9.1 调整告警阈值

```sql
-- 修改队列积压阈值
UPDATE alert_config
SET conditions = jsonb_set(
    conditions,
    '{threshold}',
    '200'::jsonb
)
WHERE alert_type = 'queue_backlog';

-- 调整冷却时间
UPDATE alert_config
SET cooldown_minutes = 120
WHERE alert_type = 'queue_backlog';
```

### 9.2 查看监控状态

```sql
-- 查看活跃告警统计
SELECT
    severity,
    COUNT(*) as count
FROM alert_history
WHERE status = 'active'
GROUP BY severity;

-- 查看告警趋势
SELECT
    DATE(created_at) as date,
    alert_type,
    COUNT(*) as count
FROM alert_history
WHERE created_at > NOW() - interval '7 days'
GROUP BY DATE(created_at), alert_type
ORDER BY date DESC, count DESC;

-- 查看监控任务执行状态
SELECT
    jobname,
    last_run_status,
    last_successful_run,
    next_scheduled_run
FROM cron.job
WHERE jobname LIKE 'monitoring%';
```

### 9.3 告警管理最佳实践

1. **及时响应**: 设置告警后要确保有人监控和响应
2. **调整阈值**: 根据实际业务情况定期调整告警阈值
3. **避免告警疲劳**: 使用冷却时间防止重复告警
4. **定期复盘**: 分析告警历史，优化监控策略
5. **权限控制**: 只有管理员和相关人员才能看到告警

---

## 十、故障排查

### 10.1 定时任务未执行

```sql
-- 检查 pg_cron 是否启用
SELECT * FROM pg_extension WHERE extname = 'pg_cron';

-- 查看定时任务列表
SELECT * FROM cron.job;

-- 查看任务执行日志
SELECT * FROM cron.job_run_details
WHERE jobid = (SELECT jobid FROM cron.job WHERE jobname = 'monitoring-checks')
ORDER BY start_time DESC
LIMIT 10;

-- 检查任务是否有错误
SELECT * FROM cron.job_run_details
WHERE status = 'failed'
ORDER BY start_time DESC;
```

### 10.2 告警未发送

```sql
-- 检查告警是否生成
SELECT * FROM alert_history
WHERE created_at > NOW() - interval '1 hour'
ORDER BY created_at DESC;

-- 检查通知是否发送
SELECT
    ah.*,
    ah.notification_sent,
    ah.notification_error
FROM alert_history ah
WHERE ah.created_at > NOW() - interval '1 hour';

-- 检查触发器是否存在
SELECT * FROM pg_trigger WHERE tgname LIKE '%alert%';
```

### 10.3 性能问题

```sql
-- 如果监控函数执行缓慢，检查索引
SELECT * FROM pg_indexes WHERE tablename LIKE '%review%';

-- 分析查询性能
EXPLAIN ANALYZE SELECT * FROM unified_queue_stats;

-- 清理历史数据
DELETE FROM monitoring_metrics
WHERE recorded_at < NOW() - interval '90 days';

DELETE FROM alert_history
WHERE created_at < NOW() - interval '90 days'
    AND status = 'resolved';
```

---

## 十一、扩展功能

### 11.1 集成第三方告警平台

可以将告警发送到以下平台：

- **钉钉/企业微信**: 修改 Edge Function 调用 Webhook API
- **Slack**: 使用 Slack Incoming Webhooks
- **PagerDuty**: 用于紧急告警升级
- **Grafana**: 可视化监控数据
- **Prometheus**: 导出指标供 Prometheus 采集

### 11.2 自定义监控指标

```sql
-- 添加自定义监控函数
CREATE OR REPLACE FUNCTION check_custom_metric()
RETURNS VOID AS $$
BEGIN
    -- 你的自定义监控逻辑
    -- 例如：检测特定用户的审核质量

    INSERT INTO monitoring_metrics (metric_type, metric_name, metric_value, metric_unit, details)
    SELECT
        'reviewer_quality',
        'accuracy_rate',
        (COUNT(*) FILTER (WHERE is_accurate = TRUE))::NUMERIC / COUNT(*) * 100,
        'percentage',
        jsonb_build_object('reviewer_id', reviewer_id)
    FROM quality_check_results
    WHERE created_at > NOW() - interval '24 hours'
    GROUP BY reviewer_id;
END;
$$ LANGUAGE plpgsql;

-- 添加到定时任务
SELECT cron.schedule(
    'custom-metrics-check',
    '0 */4 * * *',  -- 每4小时
    $$SELECT check_custom_metric()$$
);
```

---

## 十二、总结

本监控告警系统具有以下特点：

### 优势
- 完全基于 Supabase，无需额外基础设施
- 利用 PostgreSQL 强大的数据处理能力
- 实时检测，及时响应
- 灵活配置，易于扩展
- 无需修改业务代码

### 关键组件
- **告警配置表** - 灵活的规则管理
- **监控函数** - 自动化检测逻辑
- **定时任务** - pg_cron 定期执行
- **通知机制** - 多渠道告警推送
- **历史记录** - 完整的审计日志

### 下一步
1. 执行数据库迁移
2. 启用 pg_cron 扩展
3. 配置定时任务
4. 测试告警流程
5. 根据业务需求调整阈值
6. （可选）部署 Edge Function 增强通知能力

有任何问题或需要进一步定制，请参考本文档或查询 Supabase 官方文档。
