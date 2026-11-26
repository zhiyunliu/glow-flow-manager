<template>
  <el-drawer :model-value="modelValue" :title="chain ? `版本管理 - ${chain.name}` : '版本管理'" size="980px" @update:model-value="(v: boolean) => $emit('update:modelValue', v)" @open="load">
    <div class="toolbar">
      <div class="text-muted">
        引擎加载 version_id 最大的「已发布」版本；下线最新版本后自动回退到上一个已发布版本。
      </div>
      <el-button type="primary" size="small" :icon="'DocumentAdd'" @click="createDraft()">新建草稿</el-button>
    </div>
    <el-table v-loading="loading" :data="versions" border size="small">
      <el-table-column label="版本" width="120">
        <template #default="{ row }">
          <span class="nowrap">v{{ row.version_no }}</span>
          <el-tag v-if="row.is_effective" size="small" type="success" effect="dark" class="ml">生效中</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag size="small" :type="VERSION_STATES[row.publish_state]?.type">{{ VERSION_STATES[row.publish_state]?.label }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="节点/连接" width="90">
        <template #default="{ row }">{{ row.node_count }} / {{ row.conn_count }}</template>
      </el-table-column>
      <el-table-column prop="desc" label="说明" min-width="150" show-overflow-tooltip />
      <el-table-column label="发布" width="170">
        <template #default="{ row }">
          <div v-if="row.publish_time">{{ row.publish_time }}</div>
          <div class="text-muted">{{ row.publish_by }}</div>
        </template>
      </el-table-column>
      <el-table-column label="更新" width="170">
        <template #default="{ row }">
          <div>{{ row.update_time }}</div>
          <div class="text-muted">{{ row.operator }}</div>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="250" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="open(row)">{{ row.publish_state === 0 ? '设计' : '查看' }}</el-button>
          <el-button v-if="row.publish_state !== 0" link type="primary" size="small" @click="createDraft(row)">基于此版本新建草稿</el-button>
          <el-button v-if="row.publish_state === 1" link type="warning" size="small" @click="offline(row)">下线</el-button>
          <el-button v-if="row.publish_state === 2" link type="success" size="small" @click="online(row)">上线</el-button>
          <el-button v-if="row.publish_state === 0" link type="danger" size="small" @click="remove(row)">删除</el-button>
          <el-button link size="small" @click="exportVersion(row)">导出</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-drawer>
</template>

<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { chainApi, versionApi } from '@/api'
import { VERSION_STATES } from '@/constants'
import type { ChainInfo, VersionInfo } from '@/types'
import { downloadJson } from '@/utils'

const props = defineProps<{ modelValue: boolean; chain: ChainInfo | null }>()
const emit = defineEmits<{ (e: 'update:modelValue', value: boolean): void; (e: 'changed'): void }>()

const router = useRouter()
const loading = ref(false)
const versions = ref<VersionInfo[]>([])

async function load() {
  if (!props.chain) return
  loading.value = true
  try {
    versions.value = await versionApi.list(props.chain.chain_no)
  } finally {
    loading.value = false
  }
}

function open(row: VersionInfo) {
  router.push(`/designer/${row.version_id}`)
}

async function createDraft(base?: VersionInfo) {
  if (!props.chain) return
  const draft = await versionApi.createDraft(props.chain.chain_no, base?.version_id)
  if (base && draft.base_version_id !== base.version_id) {
    ElMessage.info(`流程已有草稿 v${draft.version_no}，已打开该草稿`)
  }
  router.push(`/designer/${draft.version_id}`)
}

async function offline(row: VersionInfo) {
  const effectiveOnly = row.is_effective && !versions.value.some((v) => v.version_id !== row.version_id && v.publish_state === 1)
  await ElMessageBox.confirm(
    effectiveOnly ? `v${row.version_no} 是唯一的已发布版本，下线后引擎将不再加载该流程。确定下线吗？` : `确定下线 v${row.version_no} 吗？`,
    '下线版本',
    { type: 'warning' },
  )
  await versionApi.offline(row.version_id)
  ElMessage.success('已下线')
  await load()
  emit('changed')
}

async function online(row: VersionInfo) {
  await ElMessageBox.confirm(`确定重新上线 v${row.version_no} 吗？若它是最新的已发布版本，引擎将加载它。`, '上线版本', { type: 'info' })
  await versionApi.online(row.version_id)
  ElMessage.success('已上线')
  await load()
  emit('changed')
}

async function remove(row: VersionInfo) {
  await ElMessageBox.confirm(`确定删除草稿 v${row.version_no} 吗？`, '删除草稿', { type: 'warning' })
  await versionApi.remove(row.version_id)
  ElMessage.success('已删除')
  await load()
  emit('changed')
}

async function exportVersion(row: VersionInfo) {
  const res = await chainApi.export(row.chain_no, row.version_id)
  downloadJson(res.file_name, res.content)
}
</script>

<style scoped>
.ml {
  margin-left: 6px;
}
</style>
