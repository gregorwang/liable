# 安全性说明

## Supabase 安全建议

运行数据库安全检查后，发现以下建议：

### 1. RLS（行级安全）未启用 ⚠️

**状态**：我们创建的表（users, review_tasks, review_results, tag_config）没有启用 RLS。

**说明**：
- 我们的应用通过**后端 API** 访问数据库，不直接暴露 PostgREST API
- 所有数据访问都通过 JWT 认证和角色权限中间件控制
- **不需要启用 RLS**，因为前端不会直接访问数据库

**如果需要启用 RLS**（用于额外安全层）：

```sql
-- 启用 RLS
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE review_tasks ENABLE ROW LEVEL SECURITY;
ALTER TABLE review_results ENABLE ROW LEVEL SECURITY;
ALTER TABLE tag_config ENABLE ROW LEVEL SECURITY;

-- 创建策略（示例）
CREATE POLICY "Users can read own data" ON users
  FOR SELECT
  USING (auth.uid()::text = id::text);

-- 服务角色可以完全访问（用于后端）
CREATE POLICY "Service role can do anything" ON users
  FOR ALL
  TO service_role
  USING (true)
  WITH CHECK (true);
```

### 2. 其他安全建议

- **OTP 过期时间**：如果使用邮箱验证，建议设置 OTP 过期时间少于 1 小时
- **密码泄露保护**：建议启用 HaveIBeenPwned 检查
- **PostgreSQL 版本**：有新的安全补丁可用，建议升级

## 当前安全措施

### ✅ 已实施的安全措施

1. **密码加密**：使用 bcrypt (cost 10) 加密所有用户密码
2. **JWT 认证**：所有 API 需要有效的 JWT Token
3. **角色权限**：基于角色的访问控制（admin/reviewer）
4. **用户审批**：新注册用户需要管理员审批
5. **任务锁定**：Redis 分布式锁防止任务重复领取
6. **TLS 加密**：Redis 连接使用 TLS（Upstash）
7. **环境变量**：敏感信息通过环境变量配置，不硬编码

### 🔒 推荐的额外安全措施

1. **HTTPS**：生产环境使用 HTTPS
2. **速率限制**：添加 API 速率限制中间件
3. **日志审计**：记录敏感操作日志
4. **CORS 配置**：限制允许的来源域名
5. **SQL 注入防护**：使用参数化查询（已实施）
6. **XSS 防护**：前端实施输入清理

## 生产环境检查清单

- [ ] 修改默认管理员密码
- [ ] 设置强 JWT_SECRET
- [ ] 配置 CORS 允许的域名
- [ ] 启用 HTTPS
- [ ] 添加速率限制
- [ ] 配置日志系统
- [ ] 定期备份数据库
- [ ] 监控异常登录尝试
- [ ] 设置数据库连接池限制
- [ ] 定期更新依赖包

## 代码中的安全实践

### SQL 注入防护
```go
// ✅ 正确：使用参数化查询
db.QueryRow("SELECT * FROM users WHERE username = $1", username)

// ❌ 错误：字符串拼接
db.QueryRow("SELECT * FROM users WHERE username = '" + username + "'")
```

### 密码处理
```go
// ✅ 正确：使用 bcrypt
hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// ✅ 正确：密码字段不返回给前端
type User struct {
    Password string `json:"-"`  // 注意这里的 json:"-"
}
```

### JWT 验证
```go
// ✅ 所有保护的路由都使用认证中间件
router.Use(middleware.AuthMiddleware())
```

## 数据库访问控制

当前设置：
- 后端使用 `postgres` 用户连接（service_role 级别）
- 前端不直接访问数据库
- 所有数据操作通过后端 API

如果需要前端直接访问（使用 Supabase JS SDK）：
- 必须启用 RLS
- 使用 `anon` key 而不是 `service_role` key
- 创建适当的 RLS 策略

## 紧急响应

如果发现安全问题：

1. **立即修改密码**：
   - 数据库密码
   - Redis 密码
   - JWT_SECRET
   - 管理员账号密码

2. **撤销受影响的 Token**：
   - 实施 Token 黑名单机制
   - 强制所有用户重新登录

3. **审查日志**：
   - 检查异常访问
   - 确定影响范围

4. **更新代码**：
   - 修复安全漏洞
   - 重新部署

## 参考资源

- [Supabase 安全最佳实践](https://supabase.com/docs/guides/platform/going-into-prod#security)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go 安全编码实践](https://github.com/Checkmarx/Go-SCP)

---

**注意**：本项目主要用于学习和开发，生产环境部署前请进行完整的安全审计。

