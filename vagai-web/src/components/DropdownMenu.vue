<template>
  <div class="relative" ref="containerRef">
    <slot name="trigger" :open="isOpen" :toggle="toggle" />
    <transition name="dropdown">
      <div v-if="isOpen" :class="panelClasses">
        <slot :close="close" />
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'

const props = defineProps({
  position: {
    type: String,
    default: 'bottom-left',
    validator: (v) => ['bottom-left', 'bottom-right', 'top-left', 'top-right'].includes(v)
  },
  width: {
    type: String,
    default: 'w-56'
  },
  panelClass: {
    type: String,
    default: ''
  }
})

const isOpen = defineModel('open', { default: false })
const containerRef = ref(null)

const positionClasses = {
  'bottom-left': 'left-0 mt-2',
  'bottom-right': 'right-0 mt-2',
  'top-left': 'left-0 mb-2 bottom-full',
  'top-right': 'right-0 mb-2 bottom-full'
}

const panelClasses = computed(() => {
  const pos = positionClasses[props.position]
  const w = props.width || ''
  return `absolute ${pos} ${w} glass-card p-2 z-50 overflow-hidden ${props.panelClass}`.trim()
})

function toggle() {
  isOpen.value = !isOpen.value
}

function close() {
  isOpen.value = false
}

function onClickOutside(e) {
  if (containerRef.value && !containerRef.value.contains(e.target)) {
    isOpen.value = false
  }
}

onMounted(() => document.addEventListener('click', onClickOutside))
onUnmounted(() => document.removeEventListener('click', onClickOutside))
</script>

<style scoped>
.dropdown-enter-active, .dropdown-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.dropdown-enter-from, .dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
