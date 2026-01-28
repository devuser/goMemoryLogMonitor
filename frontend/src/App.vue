<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

interface LogEntry {
  timestamp: string
  message: string
}

interface LogResponse {
  logs: LogEntry[]
  total: number
  page: number
  pageSize: number
}

interface StatusResponse {
  status: string
  logCount: number
  timestamp: string
}

const logs = ref<LogEntry[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(50)
const keyword = ref('')
const startDate = ref('')
const endDate = ref('')
const sortOrder = ref('desc')
const status = ref<StatusResponse | null>(null)
const loading = ref(false)
let refreshInterval: number | null = null

const fetchLogs = async () => {
  loading.value = true
  try {
    const params = new URLSearchParams({
      page: currentPage.value.toString(),
      pageSize: pageSize.value.toString(),
      sort: sortOrder.value,
    })
    
    if (keyword.value) {
      params.append('keyword', keyword.value)
    }
    if (startDate.value) {
      params.append('start', new Date(startDate.value).toISOString())
    }
    if (endDate.value) {
      params.append('end', new Date(endDate.value).toISOString())
    }

    const response = await fetch(`/api/logs?${params}`)
    const data: LogResponse = await response.json()
    logs.value = data.logs || []
    total.value = data.total
  } catch (error) {
    console.error('Error fetching logs:', error)
  } finally {
    loading.value = false
  }
}

const fetchStatus = async () => {
  try {
    const response = await fetch('/api/status')
    status.value = await response.json()
  } catch (error) {
    console.error('Error fetching status:', error)
  }
}

const formatTimestamp = (timestamp: string) => {
  return new Date(timestamp).toLocaleString()
}

const goToPage = (page: number) => {
  currentPage.value = page
  fetchLogs()
}

const clearFilters = () => {
  keyword.value = ''
  startDate.value = ''
  endDate.value = ''
  currentPage.value = 1
  fetchLogs()
}

const totalPages = () => {
  return Math.ceil(total.value / pageSize.value)
}

onMounted(() => {
  fetchLogs()
  fetchStatus()
  refreshInterval = window.setInterval(() => {
    fetchLogs()
    fetchStatus()
  }, 5000) // Refresh every 5 seconds
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<template>
  <div class="app">
    <header>
      <h1>📊 MemoryLogMonitor</h1>
      <div v-if="status" class="status">
        <span>Status: {{ status.status }}</span>
        <span>Total Logs: {{ status.logCount }}</span>
      </div>
    </header>

    <div class="filters">
      <div class="filter-row">
        <input
          v-model="keyword"
          type="text"
          placeholder="Search by keyword..."
          @keyup.enter="fetchLogs"
        />
        <input
          v-model="startDate"
          type="datetime-local"
          placeholder="Start date"
        />
        <input
          v-model="endDate"
          type="datetime-local"
          placeholder="End date"
        />
        <select v-model="sortOrder" @change="fetchLogs">
          <option value="desc">Newest First</option>
          <option value="asc">Oldest First</option>
        </select>
      </div>
      <div class="filter-actions">
        <button @click="fetchLogs" class="btn-primary">Apply Filters</button>
        <button @click="clearFilters" class="btn-secondary">Clear</button>
      </div>
    </div>

    <div class="logs-container">
      <div v-if="loading" class="loading">Loading...</div>
      <div v-else-if="logs.length === 0" class="no-logs">
        No logs found. Send logs to TCP port 9090.
      </div>
      <table v-else class="logs-table">
        <thead>
          <tr>
            <th style="width: 200px">Timestamp</th>
            <th>Message</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(log, index) in logs" :key="index">
            <td class="timestamp">{{ formatTimestamp(log.timestamp) }}</td>
            <td class="message">{{ log.message }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="totalPages() > 1" class="pagination">
      <button
        @click="goToPage(currentPage - 1)"
        :disabled="currentPage === 1"
        class="btn-page"
      >
        Previous
      </button>
      <span class="page-info">
        Page {{ currentPage }} of {{ totalPages() }} ({{ total }} total)
      </span>
      <button
        @click="goToPage(currentPage + 1)"
        :disabled="currentPage >= totalPages()"
        class="btn-page"
      >
        Next
      </button>
    </div>
  </div>
</template>

<style scoped>
.app {
  max-width: 1400px;
  margin: 0 auto;
  padding: 20px;
}

header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 8px;
}

h1 {
  margin: 0;
  font-size: 28px;
}

.status {
  display: flex;
  gap: 20px;
  font-size: 14px;
}

.filters {
  background: #f5f5f5;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.filter-row {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.filter-row input,
.filter-row select {
  flex: 1;
  min-width: 150px;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.filter-actions {
  display: flex;
  gap: 10px;
}

button {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.3s;
}

.btn-primary {
  background: #667eea;
  color: white;
}

.btn-primary:hover {
  background: #5568d3;
}

.btn-secondary {
  background: #e0e0e0;
  color: #333;
}

.btn-secondary:hover {
  background: #d0d0d0;
}

.logs-container {
  min-height: 400px;
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.loading,
.no-logs {
  padding: 40px;
  text-align: center;
  color: #666;
  font-size: 16px;
}

.logs-table {
  width: 100%;
  border-collapse: collapse;
}

.logs-table thead {
  background: #f8f9fa;
}

.logs-table th {
  padding: 12px;
  text-align: left;
  font-weight: 600;
  border-bottom: 2px solid #dee2e6;
}

.logs-table td {
  padding: 12px;
  border-bottom: 1px solid #e9ecef;
}

.logs-table tbody tr:hover {
  background: #f8f9fa;
}

.timestamp {
  font-family: monospace;
  color: #666;
  white-space: nowrap;
}

.message {
  font-family: monospace;
  word-break: break-all;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 20px;
  margin-top: 20px;
  padding: 20px;
}

.btn-page {
  background: #667eea;
  color: white;
  padding: 8px 16px;
}

.btn-page:hover:not(:disabled) {
  background: #5568d3;
}

.btn-page:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.page-info {
  color: #666;
  font-size: 14px;
}
</style>
