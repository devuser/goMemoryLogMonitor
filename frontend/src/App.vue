<template>
  <div class="app-container">
    <div class="card">
      <div class="header">
        <div>
          <div class="title">MemoryLogMonitor</div>
          <div class="subtitle">实时查看内存中最近接收到的日志</div>
        </div>
        <div class="header-buttons">
          <button class="button" @click="reset" :disabled="loading">重置</button>
          <button class="button primary" @click="refresh" :disabled="loading">刷新</button>
        </div>
      </div>

      <div class="filters">
        <input class="input" type="date" v-model="startDate" placeholder="开始日期" />
        <input class="input" type="date" v-model="endDate" placeholder="结束日期" />
        <input
          class="input"
          type="text"
          v-model="query"
          placeholder="日志内容关键字"
          @keyup.enter="goFirstPage"
        />
        <input
          class="input"
          type="number"
          min="0"
          v-model.number="topN"
          placeholder="Top N 最新日志（0 表示关闭）"
        />
      </div>

      <div class="stats">
        <div>缓存日志条数：{{ cacheCount }}</div>
        <div>缓存大小：{{ (cacheSizeBytes / (1024 * 1024)).toFixed(1) }} MB</div>
        <div>当前查询结果：{{ total }} 条</div>
      </div>

      <div class="table-container">
        <table>
          <thead>
            <tr>
              <th style="width: 80px">序号</th>
              <th
                style="width: 200px; cursor: pointer; user-select: none"
                @click="toggleSort('time')"
                class="sortable"
              >
                时间
                <span class="sort-icon">
                  <span v-if="sortBy === 'time' && sortOrder === 'desc'">▼</span>
                  <span v-else-if="sortBy === 'time' && sortOrder === 'asc'">▲</span>
                  <span v-else>⇅</span>
                </span>
              </th>
              <th
                style="cursor: pointer; user-select: none"
                @click="toggleSort('content')"
                class="sortable"
              >
                内容
                <span class="sort-icon">
                  <span v-if="sortBy === 'content' && sortOrder === 'desc'">▼</span>
                  <span v-else-if="sortBy === 'content' && sortOrder === 'asc'">▲</span>
                  <span v-else>⇅</span>
                </span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="!loading && logs.length === 0">
              <td colspan="3">暂无日志</td>
            </tr>
            <tr v-for="(log, idx) in logs" :key="idx">
              <td>{{ getRowNumber(idx) }}</td>
              <td>{{ formatTime(log.time) }}</td>
              <td class="log-content">{{ log.content }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="pagination">
        <div class="pagination-info">
          <div>第 {{ page }} / {{ totalPages }} 页 （每页 {{ pageSize }} 条）</div>
          <div class="page-size-selector">
            <label for="page-size">每页显示：</label>
            <select
              id="page-size"
              class="select"
              v-model.number="pageSize"
              @change="onPageSizeChange"
              :disabled="loading"
            >
              <option :value="10">10</option>
              <option :value="20">20</option>
              <option :value="30">30</option>
              <option :value="40">40</option>
              <option :value="50">50</option>
              <option :value="100">100</option>
              <option :value="200">200</option>
            </select>
          </div>
        </div>
        <div class="pagination-buttons">
          <div class="page-jump">
            <span>跳转到</span>
            <input
              type="number"
              class="page-input"
              :min="1"
              :max="totalPages"
              v-model.number="jumpPage"
              @keyup.enter="goToPage"
              :disabled="loading"
            />
            <span>页</span>
            <button class="button" @click="goToPage" :disabled="loading">跳转</button>
          </div>
          <button class="button" @click="prevPage" :disabled="page <= 1 || loading">上一页</button>

          <button class="button" @click="nextPage" :disabled="page >= totalPages || loading">
            下一页
          </button>
        </div>
      </div>
      <div class="footer">
        <!-- 注意：MemoryLogMonitor 仅在内存中缓存最近的日志，程序重启后日志将丢失。-->
        <div class="table-divider"></div>

        <div class="table-bottom-buttons">
          <button class="button" @click="clear" :disabled="loading">清空</button>
          <button class="button" @click="showStatus" :disabled="loading">状态</button>
        </div>
      </div>
    </div>

    <!-- Status Modal -->
    <div v-if="showStatusModal" class="modal-overlay" @click="closeStatus">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>服务状态</h3>
          <button class="modal-close" @click="closeStatus">×</button>
        </div>
        <div class="modal-body">
          <table class="status-table">
            <tbody>
              <tr>
                <td class="status-label">程序占用内存</td>
                <td class="status-value">
                  {{ statusData.memoryMB ? statusData.memoryMB.toFixed(2) : "-" }}
                  MB
                </td>
              </tr>
              <tr>
                <td class="status-label">日志条数</td>
                <td class="status-value">{{ statusData.logCount ?? "-" }}</td>
              </tr>
              <tr>
                <td class="status-label">HTTP端口</td>
                <td class="status-value">{{ statusData.httpPort ?? "-" }}</td>
              </tr>
              <tr>
                <td class="status-label">TCP端口</td>
                <td class="status-value">{{ statusData.tcpPort ?? "-" }}</td>
              </tr>
            </tbody>
          </table>
          <div class="modal-footnote">
            注意：MemoryLogMonitor 仅在内存中缓存最近的日志，程序重启后日志将丢失。
          </div>
        </div>
        <div class="modal-footer">
          <button class="button primary" @click="closeStatus">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, computed, watch } from "vue";
import axios from "axios";

interface LogEntry {
  time: string;
  content: string;
}

// Session storage keys
const STORAGE_KEY_PAGE = "nemo-monitor-page";
const STORAGE_KEY_PAGE_SIZE = "nemo-monitor-page-size";

// Load from sessionStorage on init
function loadFromSessionStorage() {
  const savedPage = sessionStorage.getItem(STORAGE_KEY_PAGE);
  const savedPageSize = sessionStorage.getItem(STORAGE_KEY_PAGE_SIZE);

  if (savedPage) {
    const pageNum = parseInt(savedPage, 10);
    if (pageNum > 0) {
      return {
        page: pageNum,
        pageSize: savedPageSize ? parseInt(savedPageSize, 10) : 20,
      };
    }
  }
  if (savedPageSize) {
    const pageSizeNum = parseInt(savedPageSize, 10);
    if (pageSizeNum > 0) {
      return { page: 1, pageSize: pageSizeNum };
    }
  }
  return { page: 1, pageSize: 20 };
}

// Save to sessionStorage
function saveToSessionStorage() {
  sessionStorage.setItem(STORAGE_KEY_PAGE, page.value.toString());
  sessionStorage.setItem(STORAGE_KEY_PAGE_SIZE, pageSize.value.toString());
}

const { page: initialPage, pageSize: initialPageSize } = loadFromSessionStorage();

const logs = ref<LogEntry[]>([]);
const page = ref(initialPage);
const pageSize = ref(initialPageSize);
const total = ref(0);
const cacheCount = ref(0);
const cacheSizeBytes = ref(0);
const jumpPage = ref<number | null>(null);

const startDate = ref<string>("");
const endDate = ref<string>("");
const query = ref("");
const topN = ref<number | null>(null);
const loading = ref(false);

// Status modal
const showStatusModal = ref(false);
const statusData = ref<{
  memoryMB?: number;
  logCount?: number;
  httpPort?: number;
  tcpPort?: number;
  cacheSizeMB?: number;
}>({});

// Sorting: default to time descending (newest first)
const sortBy = ref<string>("time");
const sortOrder = ref<string>("desc");

const totalPages = computed(() => {
  if (total.value === 0) return 1;
  return Math.max(1, Math.ceil(total.value / pageSize.value));
});

function formatTime(t: string) {
  if (!t) return "";
  const d = new Date(t);
  if (isNaN(d.getTime())) return t;
  return d.toLocaleString();
}

function getRowNumber(idx: number): number {
  // 序号从1开始，基于当前页计算
  return (page.value - 1) * pageSize.value + idx + 1;
}

async function fetchLogs() {
  loading.value = true;
  try {
    const params: Record<string, any> = {
      page: page.value,
      pageSize: pageSize.value,
    };
    if (startDate.value) params.startDate = startDate.value;
    if (endDate.value) params.endDate = endDate.value;
    if (query.value) params.q = query.value;
    if (topN.value && topN.value > 0) params.topN = topN.value;
    if (sortBy.value) params.sortBy = sortBy.value;
    if (sortOrder.value) params.sortOrder = sortOrder.value;

    const res = await axios.get("/api/logs", { params });
    logs.value = res.data.items || [];
    total.value = res.data.total || 0;
    cacheCount.value = res.data.cacheCount || 0;
    cacheSizeBytes.value = res.data.cacheSizeBytes || 0;
  } catch (err) {
    console.error("fetch logs error", err);
  } finally {
    loading.value = false;
  }
}

function goFirstPage() {
  page.value = 1;
  saveToSessionStorage();
  fetchLogs();
}

function prevPage() {
  if (page.value > 1) {
    page.value -= 1;
    saveToSessionStorage();
    fetchLogs();
  }
}

function nextPage() {
  if (page.value < totalPages.value) {
    page.value += 1;
    saveToSessionStorage();
    fetchLogs();
  }
}

function goToPage() {
  if (jumpPage.value === null || jumpPage.value === undefined) {
    return;
  }

  const targetPage = Math.max(1, Math.min(jumpPage.value, totalPages.value));
  if (targetPage !== page.value) {
    page.value = targetPage;
    jumpPage.value = null;
    saveToSessionStorage();
    fetchLogs();
  }
}

function refresh() {
  fetchLogs();
}

function reset() {
  // Reset all filters and sorting to default values
  startDate.value = "";
  endDate.value = "";
  query.value = "";
  topN.value = null;
  sortBy.value = "time";
  sortOrder.value = "desc";
  page.value = 1;
  saveToSessionStorage();
  fetchLogs();
}

async function clear() {
  if (!confirm("确定要清空所有日志吗？此操作不可恢复。")) {
    return;
  }

  loading.value = true;
  try {
    await axios.delete("/api/logs");
    // Refresh the log list after clearing
    await fetchLogs();
  } catch (err) {
    console.error("clear logs error", err);
    alert("清空日志失败，请重试");
  } finally {
    loading.value = false;
  }
}

async function showStatus() {
  loading.value = true;
  try {
    const res = await axios.get("/api/status");
    statusData.value = res.data || {};
    showStatusModal.value = true;
  } catch (err) {
    console.error("fetch status error", err);
    alert("获取状态失败，请重试");
  } finally {
    loading.value = false;
  }
}

function closeStatus() {
  showStatusModal.value = false;
}

function toggleSort(field: string) {
  if (sortBy.value === field) {
    // Same field: toggle order (desc -> asc -> desc)
    sortOrder.value = sortOrder.value === "desc" ? "asc" : "desc";
  } else {
    // Different field: set new field, default to desc
    sortBy.value = field;
    sortOrder.value = "desc";
  }
  // Reset to first page when sorting changes
  page.value = 1;
  saveToSessionStorage();
  fetchLogs();
}

function onPageSizeChange() {
  // Reset to first page when page size changes
  page.value = 1;
  saveToSessionStorage();
  fetchLogs();
}

// Watch for page and pageSize changes and save to sessionStorage
watch(page, () => {
  saveToSessionStorage();
  // Reset jump page input when page changes
  jumpPage.value = null;
});

watch(pageSize, () => {
  saveToSessionStorage();
});

onMounted(() => {
  fetchLogs();
});
</script>
