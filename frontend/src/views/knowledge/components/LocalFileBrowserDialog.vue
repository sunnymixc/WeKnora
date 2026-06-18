<template>
  <t-dialog
    v-model:visible="dialogVisible"
    :header="t('localImport.title')"
    :confirm-btn="{ content: t('common.confirm'), theme: 'primary', disabled: !canConfirm }"
    :cancel-btn="{ content: t('common.cancel') }"
    width="620px"
    @confirm="handleConfirm"
    @cancel="handleCancel"
  >
    <div class="local-browser">
      <!-- 面包屑 -->
      <div class="lb-breadcrumb">
        <t-breadcrumb :max-item-width="'160'">
          <t-breadcrumb-item @click="navigateTo('')">
            <t-icon name="folder" /> {{ t('localImport.root') }}
          </t-breadcrumb-item>
          <t-breadcrumb-item
            v-for="(seg, idx) in pathSegments"
            :key="`${seg}-${idx}`"
            @click="navigateTo(segmentPath(idx))"
          >
            {{ seg }}
          </t-breadcrumb-item>
        </t-breadcrumb>
      </div>

      <!-- 目录列表 -->
      <t-loading :loading="loading" size="small" class="lb-list">
        <div v-if="errorMsg" class="lb-error">{{ errorMsg }}</div>
        <template v-else>
          <div
            v-if="currentPath"
            class="lb-item lb-dir"
            @click="navigateTo(parentPath)"
          >
            <t-icon name="rollback" class="lb-icon" />
            <span class="lb-name">..</span>
          </div>
          <div
            v-for="entry in entries"
            :key="entry.name"
            class="lb-item"
            :class="{ 'lb-dir': entry.is_dir, 'lb-selected': !entry.is_dir && selectedPath === joinPath(currentPath, entry.name) }"
            @click="onEntryClick(entry)"
          >
            <t-icon :name="entry.is_dir ? 'folder' : 'file'" class="lb-icon" />
            <span class="lb-name" :title="entry.name">{{ entry.name }}</span>
            <span v-if="!entry.is_dir" class="lb-size">{{ formatSize(entry.size) }}</span>
          </div>
          <div v-if="!loading && entries.length === 0 && !currentPath" class="lb-empty">
            {{ t('localImport.empty') }}
          </div>
        </template>
      </t-loading>

      <!-- 手动输入路径 -->
      <div class="lb-manual">
        <div class="lb-manual-label">{{ t('localImport.manualLabel') }}</div>
        <t-input
          v-model="manualPath"
          :placeholder="t('localImport.manualPlaceholder')"
          clearable
          @enter="handleConfirm"
        />
        <div class="lb-manual-tip">{{ t('localImport.manualTip') }}</div>
      </div>
    </div>
  </t-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { listLocalFiles } from '@/api/knowledge-base'

interface LocalEntry {
  name: string
  is_dir: boolean
  size: number
  mod_time: number
}

const props = defineProps<{
  visible: boolean
  kbId: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  select: [path: string]
}>()

const { t } = useI18n()

const dialogVisible = computed({
  get: () => props.visible,
  set: (v: boolean) => emit('update:visible', v),
})

const currentPath = ref('')
const entries = ref<LocalEntry[]>([])
const selectedPath = ref('')
const manualPath = ref('')
const loading = ref(false)
const errorMsg = ref('')

const pathSegments = computed(() => currentPath.value.split('/').filter(Boolean))
const parentPath = computed(() => {
  const segs = pathSegments.value
  return segs.slice(0, -1).join('/')
})
const canConfirm = computed(() => !!selectedPath.value || !!manualPath.value.trim())

const segmentPath = (idx: number) => pathSegments.value.slice(0, idx + 1).join('/')
const joinPath = (dir: string, name: string) => (dir ? `${dir}/${name}` : name)

const formatSize = (size: number) => {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}

const loadDir = async (path: string) => {
  if (!props.kbId) return
  loading.value = true
  errorMsg.value = ''
  try {
    const resp: any = await listLocalFiles(props.kbId, path)
    const data = resp?.data ?? resp
    currentPath.value = data?.path ?? path
    entries.value = Array.isArray(data?.entries) ? data.entries : []
  } catch (error: any) {
    errorMsg.value = error?.error?.message || error?.message || t('localImport.loadFailed')
    entries.value = []
  } finally {
    loading.value = false
  }
}

const navigateTo = (path: string) => {
  selectedPath.value = ''
  loadDir(path)
}

const onEntryClick = (entry: LocalEntry) => {
  if (entry.is_dir) {
    navigateTo(joinPath(currentPath.value, entry.name))
  } else {
    selectedPath.value = joinPath(currentPath.value, entry.name)
    manualPath.value = ''
  }
}

const handleConfirm = () => {
  const path = manualPath.value.trim() || selectedPath.value
  if (!path) {
    MessagePlugin.warning(t('localImport.selectRequired'))
    return
  }
  emit('select', path)
  dialogVisible.value = false
}

const handleCancel = () => {
  dialogVisible.value = false
}

// Reset and load root each time the dialog opens.
watch(
  () => props.visible,
  (visible) => {
    if (visible) {
      currentPath.value = ''
      selectedPath.value = ''
      manualPath.value = ''
      errorMsg.value = ''
      loadDir('')
    }
  },
)
</script>

<style lang="less" scoped>
.local-browser {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.lb-breadcrumb {
  :deep(.t-breadcrumb__item) {
    cursor: pointer;
  }
}

.lb-list {
  min-height: 240px;
  max-height: 320px;
  overflow-y: auto;
  border: 1px solid var(--td-component-border);
  border-radius: 6px;
  padding: 4px;
}

.lb-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 4px;
  cursor: pointer;

  &:hover {
    background: var(--td-bg-color-container-hover);
  }
}

.lb-dir .lb-name {
  color: var(--td-text-color-primary);
  font-weight: 500;
}

.lb-selected {
  background: var(--td-brand-color-light);
}

.lb-icon {
  flex-shrink: 0;
  color: var(--td-text-color-secondary);
}

.lb-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.lb-size {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--td-text-color-placeholder);
}

.lb-empty,
.lb-error {
  padding: 24px;
  text-align: center;
  color: var(--td-text-color-placeholder);
}

.lb-error {
  color: var(--td-error-color);
}

.lb-manual {
  .lb-manual-label {
    margin-bottom: 6px;
    font-size: 14px;
    font-weight: 500;
    color: var(--td-text-color-primary);
  }

  .lb-manual-tip {
    margin-top: 6px;
    font-size: 12px;
    color: var(--td-text-color-placeholder);
  }
}
</style>
