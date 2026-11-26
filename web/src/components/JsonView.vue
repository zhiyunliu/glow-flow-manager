<template>
  <pre class="json-view" :style="{ maxHeight }">{{ text }}</pre>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { prettyJson } from '@/utils'

const props = withDefaults(defineProps<{ value: any; maxHeight?: string; empty?: string }>(), { maxHeight: '420px', empty: '(空)' })

const text = computed(() => {
  if (props.value === undefined || props.value === null || props.value === '') return props.empty
  return prettyJson(props.value)
})
</script>
