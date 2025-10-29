# 视频审核API测试指南

## 修复总结

### 问题原因
前端提交的`QualityDimensions`结构与数据库JSONB字段不兼容，Go无法直接将结构体写入JSONB。

### 解决方案
1. 移除了Go模型中`QualityDimension.Notes`字段（前端没有）
2. 在repository层添加JSON序列化，写入数据库前将结构体转为JSON字节
3. 在读取时添加JSON反序列化，从JSONB字段解析回结构体

## 测试步骤

### 1. 已为你准备好的任务
- **任务ID**: 7
- **视频ID**: 7  
- **状态**: in_progress（已领取）
- **审核员**: 你的账号（ID: 2, 1985738212@qq.com）

### 2. 提交审核（前端）
在VideoReviewForm中提交任务7，数据结构示例：

```json
{
  "task_id": 7,
  "is_approved": true,
  "quality_dimensions": {
    "content_quality": {
      "score": 8,
      "tags": ["创意优秀", "内容有趣"]
    },
    "technical_quality": {
      "score": 7,
      "tags": ["画质清晰", "音质良好"]
    },
    "compliance": {
      "score": 9,
      "tags": ["内容合规"]
    },
    "engagement_potential": {
      "score": 8,
      "tags": ["传播性强", "互动性好"]
    }
  },
  "traffic_pool_result": "精品池",
  "reason": "视频质量优秀，内容创意独特，符合平台规范"
}
```

### 3. API端点
- **后端地址**: `http://localhost:8080`
- **提交端点**: `POST /api/tasks/video-first-review/submit`
- **需要认证**: Bearer Token

### 4. 完整工作流程

#### 方式1：使用前端（推荐）
1. 打开前端 `http://localhost:3000`
2. 登录你的账号
3. 进入视频一审页面
4. 任务7应该会显示在"我的任务"中
5. 填写评分和理由后提交

#### 方式2：直接API测试
```powershell
# 1. 登录获取token
$loginData = @{username="1985738212@qq.com"; password="你的密码"} | ConvertTo-Json
$loginResp = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -Body $loginData -ContentType "application/json"
$token = $loginResp.token

# 2. 提交审核
$reviewData = @{
    task_id = 7
    is_approved = $true
    quality_dimensions = @{
        content_quality = @{ score = 8; tags = @("创意优秀") }
        technical_quality = @{ score = 7; tags = @("画质清晰") }
        compliance = @{ score = 9; tags = @("内容合规") }
        engagement_potential = @{ score = 8; tags = @("传播性强") }
    }
    traffic_pool_result = "精品池"
    reason = "测试提交"
} | ConvertTo-Json -Depth 10

Invoke-RestMethod -Uri "http://localhost:8080/api/tasks/video-first-review/submit" `
    -Method POST `
    -Body $reviewData `
    -ContentType "application/json" `
    -Headers @{"Authorization" = "Bearer $token"}
```

## 关于Redis队列

是的，系统使用了Redis来管理任务队列：

### Redis键结构
- `video:first:claimed:{reviewer_id}` - 用户领取的任务集合
- `video:first:lock:{task_id}` - 任务锁（防止重复领取）
- `video:review:queue:second` - 二审任务队列
- `video:stats:*` - 统计数据

### 工作流程
1. **领取任务** → PostgreSQL更新任务状态 + Redis记录领取信息
2. **提交审核** → PostgreSQL写入审核结果 + Redis清除任务记录 + 如果不通过则推送到二审队列
3. **归还任务** → PostgreSQL重置任务状态 + Redis清除记录

### 超时机制
- 后台worker会定期检查超时任务（默认30分钟）
- 超时任务会自动归还到pending状态
- Redis记录会被清除

## 可用的待审任务

如果任务7已完成，还有这些pending任务可用：
- 任务8 (video_id: 8)
- 任务9 (video_id: 9)
- 任务10 (video_id: 10)
- 任务11 (video_id: 11)

需要领取新任务时，可以：
1. 前端点击"领取任务"按钮
2. 或使用API: `POST /api/tasks/video-first-review/claim` with `{"count": 1}`

## 验证成功

提交成功后应该看到：
```json
{
  "message": "Video first review submitted successfully"
}
```

然后可以查询数据库验证：
```sql
SELECT * FROM video_first_review_results WHERE task_id = 7;
```

应该能看到quality_dimensions字段包含完整的JSONB数据。

## 常见问题

### Q: 提示"task not found or already completed"
A: 任务不是in_progress状态或不属于你。需要先领取任务。

### Q: 提示"you still have X uncompleted tasks"
A: 有未完成的任务，需要先完成或归还。

### Q: 前端连不上后端
A: 检查：
   - 后端运行在 8080 端口
   - 前端的代理配置指向正确端口
   - 查看 `vite.config.ts` 的 proxy 设置

