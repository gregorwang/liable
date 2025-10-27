-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('info', 'warning', 'success', 'error', 'system', 'announcement', 'task_update')),
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    is_global BOOLEAN NOT NULL DEFAULT TRUE
);

-- Create user_notifications table for tracking read status
CREATE TABLE IF NOT EXISTS user_notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    notification_id INTEGER NOT NULL REFERENCES notifications(id),
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, notification_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_is_global ON notifications(is_global);
CREATE INDEX IF NOT EXISTS idx_user_notifications_user_id ON user_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_user_notifications_notification_id ON user_notifications(notification_id);
CREATE INDEX IF NOT EXISTS idx_user_notifications_is_read ON user_notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_user_notifications_user_read ON user_notifications(user_id, is_read);

-- Insert some sample notifications for testing
INSERT INTO notifications (title, content, type, created_by, is_global) VALUES
('系统维护通知', '系统将于本周六凌晨2:00-4:00进行例行维护，期间可能影响服务使用，请提前做好准备。', 'system', 1, true),
('审核规则更新', '审核规则库已更新，新增了3条违规标签规则，请审核员及时查看并按照新规则执行审核工作。', 'announcement', 1, true),
('功能优化通知', '系统界面已优化，新增了批量提交功能，提升了审核效率。如有问题请及时反馈。', 'info', 1, true)
ON CONFLICT DO NOTHING;
