<template>
  <div class="modal-overlay" @click="closeModal">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3 class="modal-title">Large File Selection Warning</h3>
        <button class="close-button" @click="closeModal">
          <XIcon class="icon" />
        </button>
      </div>
      
      <div class="modal-body">
        <div class="warning-icon">
          <AlertTriangleIcon class="icon" />
        </div>
        
        <p class="warning-message">
          You have selected {{ fileCount }} files, which exceeds the recommended limit of 500 files.
        </p>
        
        <p class="warning-details">
          Selecting too many files may cause performance issues or memory problems. 
          Consider selecting fewer files for better performance.
        </p>
        
        <div class="stats">
          <div class="stat-item">
            <span class="label">Selected Files:</span>
            <span class="value">{{ fileCount }}</span>
          </div>
          <div class="stat-item">
            <span class="label">Recommended Limit:</span>
            <span class="value">500</span>
          </div>
        </div>
      </div>
      
      <div class="modal-footer">
        <button class="btn secondary" @click="closeModal">
          Cancel
        </button>
        <button class="btn primary" @click="confirmSelection">
          Continue Anyway
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { XIcon, AlertTriangleIcon } from 'lucide-vue-next'

interface Props {
  fileCount: number
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'confirm'): void
}>()

const closeModal = () => {
  emit('close')
}

const confirmSelection = () => {
  emit('confirm')
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: #1e293b;
  border-radius: 8px;
  border: 1px solid #334155;
  width: 100%;
  max-width: 450px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #334155;
}

.modal-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #f1f5f9;
  margin: 0;
}

.close-button {
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-button:hover {
  background: #334155;
  color: #f1f5f9;
}

.icon {
  width: 20px;
  height: 20px;
}

.modal-body {
  padding: 20px;
  flex: 1;
  overflow-y: auto;
}

.warning-icon {
  display: flex;
  justify-content: center;
  margin-bottom: 16px;
}

.warning-icon .icon {
  width: 48px;
  height: 48px;
  color: #f59e0b;
}

.warning-message {
  text-align: center;
  font-size: 1rem;
  color: #f1f5f9;
  margin-bottom: 12px;
  font-weight: 500;
}

.warning-details {
  text-align: center;
  font-size: 0.875rem;
  color: #94a3b8;
  margin-bottom: 20px;
  line-height: 1.5;
}

.stats {
  background: #334155;
  border-radius: 6px;
  padding: 16px;
  margin-bottom: 20px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
}

.stat-item:last-child {
  margin-bottom: 0;
}

.label {
  color: #94a3b8;
}

.value {
  color: #f1f5f9;
  font-weight: 500;
}

.modal-footer {
  display: flex;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #334155;
  justify-content: flex-end;
}

.btn {
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s;
}

.btn.secondary {
  background: #334155;
  color: #cbd5e1;
  border-color: #475569;
}

.btn.secondary:hover {
  background: #475569;
  border-color: #64748b;
}

.btn.primary {
  background: #f59e0b;
  color: #1e293b;
  border-color: #d97706;
}

.btn.primary:hover {
  background: #d97706;
  border-color: #b45309;
}
</style>