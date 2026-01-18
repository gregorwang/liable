/**
 * Video queue API smoke + latency test.
 * Usage:
 *   TOKEN=... node test/video-queue-perf-test.js
 */

const axios = require('axios');

const API_BASE_URL = process.env.API_BASE_URL || 'http://127.0.0.1:8080/api';
const POOL = process.env.VIDEO_POOL || '100k';
const CLAIM_COUNT = Number(process.env.CLAIM_COUNT || '1');

function getAuthHeaders() {
  const token = process.env.TOKEN;
  if (!token) {
    throw new Error('Missing TOKEN env var for authenticated requests.');
  }
  return { Authorization: `Bearer ${token}` };
}

async function timedRequest(name, config) {
  const start = Date.now();
  try {
    const response = await axios({
      timeout: 20000,
      validateStatus: () => true,
      ...config,
    });
    const durationMs = Date.now() - start;
    const status = response.status;
    const ok = status >= 200 && status < 300;
    console.log(`${ok ? 'OK' : 'FAIL'} ${name} - ${status} (${durationMs}ms)`);
    if (!ok) {
      console.log('  Response:', JSON.stringify(response.data));
    }
    return { ok, status, durationMs };
  } catch (err) {
    const durationMs = Date.now() - start;
    console.log(`FAIL ${name} - error (${durationMs}ms)`);
    console.log('  Error:', err.message || err);
    return { ok: false, status: 0, durationMs };
  }
}

(async () => {
  console.log(`API base: ${API_BASE_URL}`);
  console.log(`Pool: ${POOL}`);

  await timedRequest('GET /queues', {
    method: 'GET',
    url: `${API_BASE_URL}/queues?page=1&page_size=20`,
  });

  const authHeaders = getAuthHeaders();

  await timedRequest(`GET /video/${POOL}/tasks/my`, {
    method: 'GET',
    url: `${API_BASE_URL}/video/${POOL}/tasks/my`,
    headers: authHeaders,
  });

  await timedRequest(`GET /video/${POOL}/tags`, {
    method: 'GET',
    url: `${API_BASE_URL}/video/${POOL}/tags`,
    headers: authHeaders,
  });

  await timedRequest(`POST /video/${POOL}/tasks/claim`, {
    method: 'POST',
    url: `${API_BASE_URL}/video/${POOL}/tasks/claim`,
    headers: authHeaders,
    data: { count: CLAIM_COUNT },
  });
})();
