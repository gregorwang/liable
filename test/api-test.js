/**
 * API 端到端测试脚本
 * 用于测试所有API接口的可用性
 */

const axios = require('axios');

// 配置
const API_BASE_URL = 'http://127.0.0.1:8080/api';
const API_BASE_URL_NO_PREFIX = 'http://127.0.0.1:8080';

// 存储测试状态
let testResults = {
  total: 0,
  passed: 0,
  failed: 0,
  errors: []
};

// 存储认证信息
let authData = {
  token: null,
  userId: null,
  username: null
};

// 存储测试数据
let testData = {
  taskId: null,
  notificationId: null,
  tagId: null,
  videoTagId: null,
  moderationRuleId: null,
  taskQueueId: null
};

// 颜色输出
const colors = {
  reset: '\x1b[0m',
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  cyan: '\x1b[36m',
  magenta: '\x1b[35m'
};

// 测试统计
function logStats() {
  console.log(`\n${colors.cyan}═══════════════════════════════════════════════════════════${colors.reset}`);
  console.log(`${colors.cyan}测试统计${colors.reset}`);
  console.log(`${colors.cyan}═══════════════════════════════════════════════════════════${colors.reset}`);
  console.log(`总计: ${testResults.total}`);
  console.log(`${colors.green}通过: ${testResults.passed}${colors.reset}`);
  console.log(`${colors.red}失败: ${testResults.failed}${colors.reset}`);
  console.log(`${colors.cyan}═══════════════════════════════════════════════════════════${colors.reset}\n`);

  if (testResults.errors.length > 0) {
    console.log(`${colors.red}失败的测试详情:${colors.reset}\n`);
    testResults.errors.forEach((error, index) => {
      console.log(`${colors.red}${index + 1}. ${error.name}${colors.reset}`);
      console.log(`   ${colors.yellow}路径: ${error.method} ${error.path}${colors.reset}`);
      console.log(`   ${colors.red}错误: ${error.message}${colors.reset}\n`);
    });
  }
}

// 测试辅助函数
function testApi(name, method, path, data = null, headers = {}, expectedStatus = 200) {
  testResults.total++;
  const fullUrl = path.startsWith('http') ? path : API_BASE_URL + path;
  const axiosConfig = {
    method,
    url: fullUrl,
    headers,
    data,
    timeout: 5000,
    validateStatus: () => true // 不自动抛出错误
  };

  return axios(axiosConfig)
    .then(response => {
      const status = response.status;
      const success = status === expectedStatus;

      if (success) {
        testResults.passed++;
        console.log(`${colors.green}✓${colors.reset} ${name} (${method} ${path}) - ${colors.green}200 OK${colors.reset}`);
        return { success: true, data: response.data, status };
      } else {
        testResults.failed++;
        const errorMsg = `Expected ${expectedStatus}, got ${status}`;
        testResults.errors.push({
          name,
          method,
          path,
          message: errorMsg,
          response: response.data
        });
        console.log(`${colors.red}✗${colors.reset} ${name} (${method} ${path}) - ${colors.red}${status} ${errorMsg}${colors.reset}`);
        if (response.data && response.data.error) {
          console.log(`   ${colors.yellow}错误信息: ${response.data.error}${colors.reset}`);
        }
        return { success: false, data: response.data, status };
      }
    })
    .catch(error => {
      testResults.failed++;
      const errorMsg = error.message || error.code || 'Network error';
      testResults.errors.push({
        name,
        method,
        path,
        message: errorMsg
      });
      console.log(`${colors.red}✗${colors.reset} ${name} (${method} ${path}) - ${colors.red}${errorMsg}${colors.reset}`);
      return { success: false, error: errorMsg };
    });
}

// 获取认证头
function getAuthHeaders() {
  return authData.token ? { Authorization: `Bearer ${authData.token}` } : {};
}

// ============================================================================
// 测试套件
// ============================================================================

console.log(`${colors.cyan}═══════════════════════════════════════════════════════════${colors.reset}`);
console.log(`${colors.cyan}API 端到端测试${colors.reset}`);
console.log(`${colors.cyan}═══════════════════════════════════════════════════════════${colors.reset}`);
console.log(`测试服务器: ${API_BASE_URL_NO_PREFIX}`);
console.log(`${colors.cyan}═══════════════════════════════════════════════════════════${colors.reset}\n`);

// ============================================================================
// 1. 公共接口（无需认证）
// ============================================================================
console.log(`${colors.cyan}[1/10] 公共接口测试${colors.reset}\n`);

(async () => {
  // 健康检查 (必须使用完整URL，因为health在/api前缀之外)
  await testApi('健康检查', 'GET', 'http://127.0.0.1:8080/health', null, {}, 200);

  // 检查邮箱
  await testApi('检查邮箱（未注册）', 'GET', '/auth/check-email?email=nonexistent@test.com', null, {}, 200);

  // Moderation Rules (公开)
  await testApi('获取moderation rules列表', 'GET', '/moderation-rules');
  await testApi('获取所有moderation rules', 'GET', '/moderation-rules/all');
  await testApi('获取rule分类', 'GET', '/moderation-rules/categories');
  await testApi('获取风险等级', 'GET', '/moderation-rules/risk-levels');

  // 公共任务队列
  await testApi('获取公共任务队列列表', 'GET', '/queues');
  await testApi('获取单个公共任务队列', 'GET', '/queues/1', null, {}, 404); // 可能不存在

  // ============================================================================
  // 2. 认证接口测试
  // ============================================================================
  console.log(`\n${colors.cyan}[2/10] 认证接口测试${colors.reset}\n`);

  // 尝试注册（可能因重复而失败）
  const registerResult = await testApi('用户注册', 'POST', '/auth/register', {
    username: `testuser${Date.now()}`,
    password: 'Test123456!'
  }, {}, 200);

  if (!registerResult.success) {
    console.log(`   ${colors.yellow}注: 注册可能失败因为用户已存在或限流${colors.reset}`);
  }

  // 登录
  const loginResult = await testApi('用户登录', 'POST', '/auth/login', {
    username: 'admin',
    password: 'admin123'
  }, {}, 200);

  if (loginResult.success && loginResult.data && loginResult.data.token) {
    authData.token = loginResult.data.token;
    authData.userId = loginResult.data.user_id;
    authData.username = loginResult.data.username;
    console.log(`   ${colors.green}✓${colors.reset} 认证成功，已保存token`);
  } else {
    console.log(`   ${colors.red}✗${colors.reset} 登录失败，后续需要认证的测试将跳过`);
  }

  // 登录后获取个人信息
  if (authData.token) {
    await testApi('获取用户信息', 'GET', '/auth/profile', null, getAuthHeaders(), 200);
  }

  // ============================================================================
  // 3. 任务接口测试（第一次审核）
  // ============================================================================
  console.log(`\n${colors.cyan}[3/10] 任务接口测试（第一次审核）${colors.reset}\n`);

  if (authData.token) {
    await testApi('获取我的任务', 'GET', '/tasks/my', null, getAuthHeaders(), 200);

    // 领取任务
    const claimResult = await testApi('领取任务', 'POST', '/tasks/claim', {
      count: 1
    }, getAuthHeaders(), 200);

    if (claimResult.success && claimResult.data && claimResult.data.tasks) {
      const tasks = claimResult.data.tasks;
      if (tasks.length > 0) {
        testData.taskId = tasks[0].id;
        console.log(`   ${colors.green}✓${colors.reset} 领取到任务，ID: ${testData.taskId}`);
      }
    }

    // 提交任务
    if (testData.taskId) {
      await testApi('提交单个审核', 'POST', '/tasks/submit', {
        task_id: testData.taskId,
        review_decision: 'approved',
        reason: '测试通过'
      }, getAuthHeaders(), 200);
    }

    // 批量提交
    await testApi('批量提交审核', 'POST', '/tasks/submit-batch', {
      reviews: []
    }, getAuthHeaders(), 200);

    // 归还任务
    await testApi('归还任务', 'POST', '/tasks/return', {
      task_ids: []
    }, getAuthHeaders(), 200);
  }

  // ============================================================================
  // 4. 第二次审核接口测试
  // ============================================================================
  console.log(`\n${colors.cyan}[4/10] 第二次审核接口测试${colors.reset}\n`);

  if (authData.token) {
    await testApi('获取我的第二次审核任务', 'GET', '/tasks/second-review/my', null, getAuthHeaders(), 200);

    const secondClaimResult = await testApi('领取第二次审核任务', 'POST', '/tasks/second-review/claim', {
      count: 1
    }, getAuthHeaders(), 200);

    if (secondClaimResult.success && secondClaimResult.data && secondClaimResult.data.tasks) {
      const tasks = secondClaimResult.data.tasks;
      if (tasks.length > 0) {
        const taskId = tasks[0].id;
        await testApi('提交单个第二次审核', 'POST', '/tasks/second-review/submit', {
          task_id: taskId,
          review_decision: 'approved',
          reason: '测试通过'
        }, getAuthHeaders(), 200);
      }
    }

    await testApi('批量提交第二次审核', 'POST', '/tasks/second-review/submit-batch', {
      reviews: []
    }, getAuthHeaders(), 200);

    await testApi('归还第二次审核任务', 'POST', '/tasks/second-review/return', {
      task_ids: []
    }, getAuthHeaders(), 200);
  }

  // // ============================================================================
  // 5. 质量检查接口测试
  // ============================================================================
  console.log(`\n${colors.cyan}[5/10] 质量检查接口测试${colors.reset}\n`);

  if (authData.token) {
    await testApi('获取我的质量检查任务', 'GET', '/tasks/quality-check/my', null, getAuthHeaders(), 200);

    const qcClaimResult = await testApi('领取质量检查任务', 'POST', '/tasks/quality-check/claim', {
      count: 1
    }, getAuthHeaders(), 200);

    if (qcClaimResult.success && qcClaimResult.data && qcClaimResult.data.tasks) {
      const tasks = qcClaimResult.data.tasks;
      if (tasks.length > 0) {
        const taskId = tasks[0].id;
        await testApi('提交单个质量检查', 'POST', '/tasks/quality-check/submit', {
          task_id: taskId,
          review_decision: 'approved',
          reason: '测试通过'
        }, getAuthHeaders(), 200);
      }
    }

    await testApi('批量提交质量检查', 'POST', '/tasks/quality-check/submit-batch', {
      reviews: []
    }, getAuthHeaders(), 200);

    await testApi('归还质量检查任务', 'POST', '/tasks/quality-check/return', {
      task_ids: []
    }, getAuthHeaders(), 200);

    await testApi('获取质量检查统计', 'GET', '/tasks/quality-check/stats', null, getAuthHeaders(), 200);
  }

  // ============================================================================
  // 6. 视频审核接口测试
  // ============================================================================
  console.log(`\n${colors.cyan}[6/10] 视频审核接口测试${colors.reset}\n`);

  if (authData.token) {
    // 视频第一次审核
    await testApi('获取我的视频第一次审核任务', 'GET', '/tasks/video-first-review/my', null, getAuthHeaders(), 200);

    await testApi('领取视频第一次审核任务', 'POST', '/tasks/video-first-review/claim', {
      count: 1
    }, getAuthHeaders(), 200);

    await testApi('提交单个视频第一次审核', 'POST', '/tasks/video-first-review/submit', {
      task_id: 1,
      review_decision: 'approved',
      reason: '测试通过',
      tags: []
    }, getAuthHeaders(), 200);

    await testApi('批量提交视频第一次审核', 'POST', '/tasks/video-first-review/submit-batch', {
      reviews: []
    }, getAuthHeaders(), 200);

    await testApi('归还视频第一次审核任务', 'POST', '/tasks/video-first-review/return', {
      task_ids: []
    }, getAuthHeaders(), 200);

    // 视频第二次审核
    await testApi('获取我的视频第二次审核任务', 'GET', '/tasks/video-second-review/my', null, getAuthHeaders(), 200);

    await testApi('领取视频第二次审核任务', 'POST', '/tasks/video-second-review/claim', {
      count: 1
    }, getAuthHeaders(), 200);

    await testApi('提交单个视频第二次审核', 'POST', '/tasks/video-second-review/submit', {
      task_id: 1,
      review_decision: 'approved',
      reason: '测试通过',
      tags: []
    }, getAuthHeaders(), 200);

    await testApi('批量提交视频第二次审核', 'POST', '/tasks/video-second-review/submit-batch', {
      reviews: []
    }, getAuthHeaders(), 200);

    await testApi('归还视频第二次审核任务', 'POST', '/tasks/video-second-review/return', {
      task_ids: []
    }, getAuthHeaders(), 200);
  }

  // ============================================================================
  // 7. 视频队列池接口测试
  // ============================================================================
  console.log(`\n${colors.cyan}[7/10] 视频队列池接口测试${colors.reset}\n`);

  if (authData.token) {
    const pools = ['100k', '1m', '10m'];
    for (const pool of pools) {
      console.log(`   ${colors.blue}测试 ${pool} 队列池:${colors.reset}`);

      await testApi(`获取${pool}队列标签`, 'GET', `/video/${pool}/tags`, null, getAuthHeaders(), 200);
      await testApi(`领取${pool}队列任务`, 'POST', `/video/${pool}/tasks/claim`, {
        count: 1
      }, getAuthHeaders(), 200);
      await testApi(`获取我的${pool}队列任务`, 'GET', `/video/${pool}/tasks/my`, null, getAuthHeaders(), 200);
      await testApi(`提交${pool}队列审核`, 'POST', `/video/${pool}/tasks/submit`, {
        task_id: 1,
        review_decision: 'push_next_pool',
        reason: '测试通过',
        tags: []
      }, getAuthHeaders(), 200);
      await testApi(`批量提交${pool}队列审核`, 'POST', `/video/${pool}/tasks/submit-batch`, {
        reviews: []
      }, getAuthHeaders(), 200);
      await testApi(`归还${pool}队列任务`, 'POST', `/video/${pool}/tasks/return`, {
        task_ids: []
      }, getAuthHeaders(), 200);
    }

    // 获取视频质量标签
    await testApi('获取视频质量标签', 'GET', '/video-quality-tags', null, getAuthHeaders(), 200);
  }

  // ============================================================================
  // 8. 通知接口测试
  // ============================================================================
  console.log(`\n${colors.cyan}[8/10] 通知接口测试${colors.reset}\n`);

  if (authData.token) {
    await testApi('获取未读通知', 'GET', '/notifications/unread', null, getAuthHeaders(), 200);
    await testApi('获取未读通知数', 'GET', '/notifications/unread-count', null, getAuthHeaders(), 200);
    await testApi('获取最近通知', 'GET', '/notifications/recent', null, getAuthHeaders(), 200);

    // 标记通知为已读
    await testApi('标记通知为已读', 'PUT', '/notifications/1/read', null, getAuthHeaders(), 200);

    // 获取今日统计
    await testApi('获取今日审核统计', 'GET', '/stats/today', null, getAuthHeaders(), 200);
  }

  // ============================================================================
  // 9. 管理接口测试
  // ============================================================================
  console.log(`\n${colors.cyan}[9/10] 管理接口测试${colors.reset}\n`);

  if (authData.token) {
    // 用户管理
    await testApi('获取待审批用户', 'GET', '/admin/users', null, getAuthHeaders(), 200);
    await testApi('获取所有用户', 'GET', '/admin/users/all', null, getAuthHeaders(), 200);

    // 权限管理
    await testApi('列出权限', 'GET', '/admin/permissions', null, getAuthHeaders(), 200);
    await testApi('获取所有权限', 'GET', '/admin/permissions/all', null, getAuthHeaders(), 200);
    await testApi('获取用户权限', 'GET', '/admin/permissions/user', null, getAuthHeaders(), 200);

    // 统计数据
    await testApi('获取概览统计', 'GET', '/admin/stats/overview', null, getAuthHeaders(), 200);
    await testApi('获取今日统计（admin）', 'GET', '/admin/stats/today', null, getAuthHeaders(), 200);
    await testApi('获取小时统计', 'GET', '/admin/stats/hourly', null, getAuthHeaders(), 200);
    await testApi('获取标签统计', 'GET', '/admin/stats/tags', null, getAuthHeaders(), 200);
    await testApi('获取审核员绩效', 'GET', '/admin/stats/reviewers', null, getAuthHeaders(), 200);

    // 标签管理
    const tagsResult = await testApi('获取所有标签', 'GET', '/admin/tags', null, getAuthHeaders(), 200);
    if (tagsResult.success && tagsResult.data && tagsResult.data.tags) {
      const tags = tagsResult.data.tags;
      if (tags.length > 0) {
        testData.tagId = tags[0].id;
      }
    }

    // 创建标签
    const createTagResult = await testApi('创建标签', 'POST', '/admin/tags', {
      name: `test_tag_${Date.now()}`,
      description: '测试标签',
      category: 'content'
    }, getAuthHeaders(), 200);

    if (createTagResult.success && createTagResult.data && createTagResult.data.tag) {
      testData.tagId = createTagResult.data.tag.id;
    }

    // 视频标签管理
    const videoTagsResult = await testApi('获取所有视频标签', 'GET', '/admin/video-tags', null, getAuthHeaders(), 200);
    if (videoTagsResult.success && videoTagsResult.data && videoTagsResult.data.tags) {
      const tags = videoTagsResult.data.tags;
      if (tags.length > 0) {
        testData.videoTagId = tags[0].id;
      }
    }

    // 创建视频标签
    const createVideoTagResult = await testApi('创建视频标签', 'POST', '/admin/video-tags', {
      name: `test_video_tag_${Date.now()}`,
      description: '测试视频标签',
      category: 'content'
    }, getAuthHeaders(), 200);

    if (createVideoTagResult.success && createVideoTagResult.data && createVideoTagResult.data.tag) {
      testData.videoTagId = createVideoTagResult.data.tag.id;
    }

    // Task Queue 管理
    await testApi('列出任务队列', 'GET', '/admin/task-queues', null, { ...getAuthHeaders() }, 200);
    await testApi('获取所有任务队列', 'GET', '/admin/task-queues-all', null, { ...getAuthHeaders() }, 200);

    // 视频队列统计
    await testApi('获取100k队列统计', 'GET', '/admin/video-queue/100k/stats', null, getAuthHeaders(), 200);
    await testApi('获取1m队列统计', 'GET', '/admin/video-queue/1m/stats', null, getAuthHeaders(), 200);
    await testApi('获取10m队列统计', 'GET', '/admin/video-queue/10m/stats', null, getAuthHeaders(), 200);

    // 视频管理
    await testApi('列出视频', 'GET', '/admin/videos', null, getAuthHeaders(), 200);
    await testApi('获取单个视频', 'GET', '/admin/videos/1', null, getAuthHeaders(), 404); // 可能不存在

    // 生成视频URL
    await testApi('生成视频URL', 'POST', '/admin/videos/generate-url', {
      video_path: 'test/path/video.mp4'
    }, getAuthHeaders(), 200);

    // Moderation Rules 管理
    await testApi('创建moderation rule', 'POST', '/admin/moderation-rules', {
      code: `test_rule_${Date.now()}`,
      name: '测试规则',
      description: '测试用审核规则',
      category: 'test',
      risk_level: 'low',
      active: true
    }, getAuthHeaders(), 200);
  }

  // ============================================================================
  // 10. 测试接口和特殊场景
  // ============================================================================
  console.log(`\n${colors.cyan}[10/10] 特殊接口和边界测试${colors.reset}\n`);

  // 测试视频审核数据结构
  await testApi('测试视频审核数据结构', 'POST', '/test/video-review-structure', {
    task_id: 1,
    review_decision: 'approved',
    reason: '测试',
    tags: []
  }, {}, 200);

  // 无认证测试（应该返回401）
  await testApi('未认证访问受保护接口', 'GET', '/tasks/my', null, {}, 401);
  await testApi('未认证访问管理接口', 'GET', '/admin/users', null, {}, 401);

  // ============================================================================
  // 最终统计
  // ============================================================================
  logStats();

  // 退出代码
  process.exit(testResults.failed > 0 ? 1 : 0);
})();
