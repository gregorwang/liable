# 监控告警系统 - 快速部署指南

## 5分钟快速部署

本指南帮助你快速部署基础监控告警功能。

---

## 步骤 1: 执行数据库迁移（2分钟）

在 Supabase Dashboard 的 SQL Editor 中依次执行以下 SQL：

### 1.1 创建告警表

```sql
-- 告警配置表
CREATE TABLE IF NOT EXISTS alert_config (
    id SERIAL PRIMARY KEY,
    alert_type VARCHAR(50) NOT NULL UNIQUE,
    alert_name VARCHAR(100) NOT NULL,
    description TEXT,
    conditions JSONB NOT NULL DEFAULT '{}',
    severity VARCHAR(20) NOT NULL DEFAULT 'medium'
        CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    notification_config JSONB DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    cooldown_minutes INTEGER DEFAULT 60,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 告警历史表
CREATE TABLE IF NOT EXISTS alert_history (
    id SERIAL PRIMARY KEY,
    alert_config_id INTEGER REFERENCES alert_config(id),
    alert_type VARCHAR(50) NOT NULL,
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL,
    trigger_data JSONB DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'resolved', 'acknowledged')),
    notification_sent BOOLEAN DEFAULT FALSE,
    notification_sent_at TIMESTAMP,
    resolved_at TIMESTAMP,
    resolved_by INTEGER REFERENCES users(id),
    resolution_notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 监控指标表
CREATE TABLE IF NOT EXISTS monitoring_metrics (
    id SERIAL PRIMARY KEY,
    metric_type VARCHAR(50) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value NUMERIC NOT NULL,
    metric_unit VARCHAR(20),
    queue_name VARCHAR(50),
    pool VARCHAR(10),
    details JSONB DEFAULT '{}',
    recorded_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 创建索引
CREATE INDEX idx_alert_config_type ON alert_config(alert_type);
CREATE INDEX idx_alert_history_created ON alert_history(created_at DESC);
CREATE INDEX idx_monitoring_metrics_recorded ON monitoring_metrics(recorded_at DESC);
```

### 1.2 插入基础告警配置

```sql
INSERT INTO alert_config (alert_type, alert_name, description, conditions, severity) VALUES
('queue_backlog', '队列积压告警', '待处理任务超过100个',
    '{"threshold": 100}', 'high'),
('task_timeout', '任务超时告警', '任务超过2小时未完成',
    '{"timeout_minutes": 120}', 'high'),
('review_rate_low', '审核速率过低', '每小时少于10个任务',
    '{"min_rate_per_hour": 10, "time_window_minutes": 60}', 'medium')
ON CONFLICT (alert_type) DO NOTHING;
```

---

## 步骤 2: 创建监控函数（1分钟）

```sql
-- 队列积压检测
CREATE OR REPLACE FUNCTION check_queue_backlog()
RETURNS VOID AS $$
DECLARE
    config RECORD;
    queue RECORD;
BEGIN
    SELECT * INTO config FROM alert_config
    WHERE alert_type = 'queue_backlog' AND is_active = TRUE LIMIT 1;

    IF NOT FOUND THEN RETURN; END IF;

    FOR queue IN
        SELECT queue_name, pending_tasks
        FROM unified_queue_stats
        WHERE pending_tasks > (config.conditions->>'threshold')::INTEGER
    LOOP
        -- 检查是否在冷却期内
        IF NOT EXISTS (
            SELECT 1 FROM alert_history
            WHERE alert_type = 'queue_backlog'
                AND trigger_data->>'queue_name' = queue.queue_name
                AND status = 'active'
                AND created_at > NOW() - (config.cooldown_minutes || ' minutes')::INTERVAL
        ) THEN
            INSERT INTO alert_history (
                alert_config_id, alert_type, title, message, severity, trigger_data
            ) VALUES (
                config.id,
                'queue_backlog',
                '队列积压: ' || queue.queue_name,
                format('%s 待处理任务 %s 个，超过阈值 %s',
                    queue.queue_name, queue.pending_tasks, config.conditions->>'threshold'),
                config.severity,
                jsonb_build_object(
                    'queue_name', queue.queue_name,
                    'pending_count', queue.pending_tasks
                )
            );
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 任务超时检测
CREATE OR REPLACE FUNCTION check_task_timeout()
RETURNS VOID AS $$
DECLARE
    config RECORD;
    timeout_count INTEGER;
BEGIN
    SELECT * INTO config FROM alert_config
    WHERE alert_type = 'task_timeout' AND is_active = TRUE LIMIT 1;

    IF NOT FOUND THEN RETURN; END IF;

    -- 检查评论一审超时
    SELECT COUNT(*)::INTEGER INTO timeout_count
    FROM review_tasks
    WHERE status = 'in_progress'
        AND claimed_at < NOW() - (config.conditions->>'timeout_minutes' || ' minutes')::INTERVAL;

    IF timeout_count > 0 THEN
        INSERT INTO alert_history (
            alert_config_id, alert_type, title, message, severity, trigger_data
        ) VALUES (
            config.id,
            'task_timeout',
            '任务超时: 评论一审',
            format('有 %s 个任务超过 %s 分钟未完成',
                timeout_count, config.conditions->>'timeout_minutes'),
            config.severity,
            jsonb_build_object('timeout_count', timeout_count, 'queue_name', 'comment_first_review')
        );
    END IF;
END;
$$ LANGUAGE plpgsql;

-- 主监控函数
CREATE OR REPLACE FUNCTION run_monitoring_checks()
RETURNS JSONB AS $$
BEGIN
    PERFORM check_queue_backlog();
    PERFORM check_task_timeout();

    RETURN jsonb_build_object(
        'timestamp', NOW(),
        'status', 'completed'
    );
END;
$$ LANGUAGE plpgsql;
```

---

## 步骤 3: 启用定时任务（1分钟）

### 3.1 启用 pg_cron 扩展

在 Supabase Dashboard:
1. 进入 **Database** → **Extensions**
2. 搜索 `pg_cron`
3. 点击 **Enable**

### 3.2 配置定时监控

```sql
-- 每5分钟执行一次监控
SELECT cron.schedule(
    'monitoring-checks',
    '*/5 * * * *',
    $$SELECT run_monitoring_checks()$$
);

-- 验证定时任务已创建
SELECT * FROM cron.job WHERE jobname = 'monitoring-checks';
```

---

## 步骤 4: 创建通知触发器（1分钟）

```sql
-- 自动发送系统通知
CREATE OR REPLACE FUNCTION send_alert_notification()
RETURNS TRIGGER AS $$
BEGIN
    -- 插入通知给所有管理员
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

    -- 更新通知发送状态
    UPDATE alert_history
    SET notification_sent = TRUE,
        notification_sent_at = NOW()
    WHERE id = NEW.id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器
DROP TRIGGER IF EXISTS trigger_send_alert ON alert_history;
CREATE TRIGGER trigger_send_alert
    AFTER INSERT ON alert_history
    FOR EACH ROW
    WHEN (NEW.status = 'active')
    EXECUTE FUNCTION send_alert_notification();
```

---

## 步骤 5: 测试验证（1分钟）

### 5.1 手动触发检测

```sql
-- 执行监控检查
SELECT run_monitoring_checks();

-- 查看是否有告警
SELECT * FROM alert_history ORDER BY created_at DESC LIMIT 5;

-- 查看是否发送了通知
SELECT * FROM notifications
WHERE type = 'alert'
ORDER BY created_at DESC
LIMIT 5;
```

### 5.2 模拟告警

```sql
-- 临时插入大量待处理任务来触发积压告警
-- 注意：这会创建实际的任务记录，仅用于测试
INSERT INTO review_tasks (comment_id, status)
SELECT id, 'pending'
FROM comments
LIMIT 120;

-- 执行检查
SELECT check_queue_backlog();

-- 查看告警
SELECT * FROM alert_history WHERE alert_type = 'queue_backlog';
```

---

## 创建简单监控视图（可选）

```sql
-- 活跃告警
CREATE OR REPLACE VIEW active_alerts AS
SELECT
    id,
    alert_type,
    title,
    message,
    severity,
    trigger_data,
    created_at
FROM alert_history
WHERE status = 'active'
ORDER BY
    CASE severity
        WHEN 'critical' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        ELSE 4
    END,
    created_at DESC;

-- 查看活跃告警
SELECT * FROM active_alerts;
```

---

## 前端快速集成

在你的管理后台添加告警展示：

```typescript
// api/alerts.ts
import { supabase } from '@/config/supabase'

export async function getActiveAlerts() {
  const { data, error } = await supabase
    .from('active_alerts')
    .select('*')
    .limit(10)

  if (error) throw error
  return data
}

export async function resolveAlert(alertId: number, notes: string) {
  const { error } = await supabase
    .from('alert_history')
    .update({
      status: 'resolved',
      resolved_at: new Date().toISOString(),
      resolution_notes: notes,
    })
    .eq('id', alertId)

  if (error) throw error
}

// 实时订阅新告警
export function subscribeToAlerts(callback: (alert: any) => void) {
  return supabase
    .channel('alerts')
    .on(
      'postgres_changes',
      {
        event: 'INSERT',
        schema: 'public',
        table: 'alert_history',
      },
      (payload) => callback(payload.new)
    )
    .subscribe()
}
```

在 Vue 组件中使用：

```vue
<template>
  <div class="alerts-panel">
    <h3>活跃告警 ({{ alerts.length }})</h3>
    <div v-for="alert in alerts" :key="alert.id" :class="['alert', alert.severity]">
      <h4>{{ alert.title }}</h4>
      <p>{{ alert.message }}</p>
      <small>{{ formatTime(alert.created_at) }}</small>
      <button @click="resolve(alert.id)">解决</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { getActiveAlerts, resolveAlert, subscribeToAlerts } from '@/api/alerts'

const alerts = ref([])
let subscription: any

onMounted(async () => {
  alerts.value = await getActiveAlerts()

  // 订阅新告警
  subscription = subscribeToAlerts((newAlert) => {
    alerts.value.unshift(newAlert)
    // 可选：显示浏览器通知
    if (Notification.permission === 'granted') {
      new Notification(newAlert.title, { body: newAlert.message })
    }
  })
})

onUnmounted(() => {
  subscription?.unsubscribe()
})

async function resolve(alertId: number) {
  const notes = prompt('解决说明：')
  if (notes) {
    await resolveAlert(alertId, notes)
    alerts.value = alerts.value.filter(a => a.id !== alertId)
  }
}

function formatTime(timestamp: string) {
  return new Date(timestamp).toLocaleString('zh-CN')
}
</script>

<style scoped>
.alert {
  padding: 12px;
  margin: 8px 0;
  border-left: 4px solid;
  border-radius: 4px;
}
.alert.critical { border-color: #d32f2f; background: #ffebee; }
.alert.high { border-color: #f57c00; background: #fff3e0; }
.alert.medium { border-color: #fbc02d; background: #fffde7; }
.alert.low { border-color: #388e3c; background: #e8f5e9; }
</style>
```

---

## 验证清单

- [ ] 告警表创建成功
- [ ] 基础告警配置已插入
- [ ] 监控函数可以执行
- [ ] pg_cron 扩展已启用
- [ ] 定时任务已创建并运行
- [ ] 通知触发器已创建
- [ ] 手动测试可以生成告警
- [ ] 告警通知正常发送
- [ ] 前端可以查看告警

---

## 下一步优化

完成基础部署后，可以参考 `MONITORING_ALERT_IMPLEMENTATION_GUIDE.md` 进行以下增强：

1. **添加更多监控指标**
   - 审核速率监控
   - 审核员空闲检测
   - 异常拒绝率监控

2. **集成外部通知**
   - 邮件通知
   - 钉钉/企业微信 Webhook
   - 短信告警

3. **优化告警策略**
   - 根据实际情况调整阈值
   - 设置不同时间段的阈值
   - 添加告警升级机制

4. **监控仪表板**
   - 创建实时监控图表
   - 历史趋势分析
   - 告警统计报表

---

## 常见问题

### Q: 如何修改告警阈值？

```sql
UPDATE alert_config
SET conditions = jsonb_set(conditions, '{threshold}', '200'::jsonb)
WHERE alert_type = 'queue_backlog';
```

### Q: 如何暂停某个告警？

```sql
UPDATE alert_config
SET is_active = FALSE
WHERE alert_type = 'queue_backlog';
```

### Q: 如何查看定时任务执行日志？

```sql
SELECT *
FROM cron.job_run_details
WHERE jobid = (SELECT jobid FROM cron.job WHERE jobname = 'monitoring-checks')
ORDER BY start_time DESC
LIMIT 10;
```

### Q: 如何清理历史告警？

```sql
-- 删除30天前已解决的告警
DELETE FROM alert_history
WHERE status = 'resolved'
    AND created_at < NOW() - INTERVAL '30 days';
```

---

## 支持

- 完整文档: `doc/MONITORING_ALERT_IMPLEMENTATION_GUIDE.md`
- Supabase 文档: https://supabase.com/docs
- pg_cron 文档: https://github.com/citusdata/pg_cron

---

**部署完成！你的监控告警系统现在已经开始工作了。** 🎉
