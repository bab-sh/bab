<script setup lang="ts">
import { ref, onMounted } from 'vue'

const isVisible = ref(false)

const dismiss = () => {
  isVisible.value = false
}

onMounted(() => {
  setTimeout(() => {
    isVisible.value = true
  }, 500)
})
</script>

<template>
  <Transition name="toast">
    <div v-if="isVisible" class="toast-container">
      <div class="toast-notification">
        <button class="close-button" @click="dismiss" aria-label="Dismiss notification">
          ✕
        </button>
        <div class="toast-content">
          <span class="warning-icon">⚠️</span>
          <p class="toast-message">
            <strong>Pre-Alpha:</strong> Bab is under active development and not ready for production use.
          </p>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.toast-container {
  position: fixed;
  bottom: 2rem;
  left: 50%;
  transform: translateX(-50%);
  z-index: 9999;
  pointer-events: none;
}

.toast-notification {
  position: relative;
  max-width: 600px;
  background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%);
  border: 1px solid #fbbf24;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  padding: 1.5rem 3rem 1.5rem 1.5rem;
  pointer-events: auto;
}

.close-button {
  position: absolute;
  top: 0.75rem;
  right: 0.75rem;
  background: none;
  border: none;
  font-size: 1.25rem;
  color: #92400e;
  cursor: pointer;
  padding: 0.25rem;
  line-height: 1;
  transition: color 0.2s, transform 0.2s;
}

.close-button:hover {
  color: #78350f;
  transform: scale(1.1);
}

.toast-content {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.warning-icon {
  font-size: 1.5rem;
  flex-shrink: 0;
}

.toast-message {
  margin: 0;
  color: #854d0e;
  font-size: 0.95rem;
  line-height: 1.5;
}

.toast-enter-active {
  animation: slideUp 0.3s ease-out;
}

.toast-leave-active {
  animation: fadeOut 0.2s ease-in;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(100px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}

@keyframes fadeOut {
  from {
    opacity: 1;
  }
  to {
    opacity: 0;
  }
}

@media (max-width: 768px) {
  .toast-container {
    bottom: 1rem;
    left: 1rem;
    right: 1rem;
    transform: none;
  }

  .toast-notification {
    max-width: none;
    padding: 1.25rem 2.5rem 1.25rem 1.25rem;
  }

  .toast-message {
    font-size: 0.9rem;
  }

  .warning-icon {
    font-size: 1.25rem;
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(100px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
}
</style>
