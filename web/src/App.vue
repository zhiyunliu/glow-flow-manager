<template>
  <router-view v-if="fullscreen" />
  <el-container v-else class="app-layout">
    <el-aside :width="collapsed ? '64px' : '208px'" class="app-aside">
      <div class="app-logo">
        <el-icon :size="22"><Share /></el-icon>
        <span v-show="!collapsed">流程编排管理</span>
      </div>
      <el-menu :default-active="activeMenu" :collapse="collapsed" router class="app-menu">
        <el-menu-item index="/chains">
          <el-icon><Files /></el-icon>
          <template #title>流程定义</template>
        </el-menu-item>
        <el-menu-item index="/node-types">
          <el-icon><Grid /></el-icon>
          <template #title>节点注册</template>
        </el-menu-item>
        <el-menu-item index="/history">
          <el-icon><Tickets /></el-icon>
          <template #title>执行历史</template>
        </el-menu-item>
        <el-menu-item index="/audit">
          <el-icon><Document /></el-icon>
          <template #title>操作审计</template>
        </el-menu-item>
      </el-menu>
      <div class="app-collapse" @click="collapsed = !collapsed">
        <el-icon><component :is="collapsed ? 'Expand' : 'Fold'" /></el-icon>
      </div>
    </el-aside>
    <el-main class="app-main">
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const collapsed = ref(false)
const fullscreen = computed(() => route.meta.fullscreen === true)
const activeMenu = computed(() => '/' + (route.path.split('/')[1] ?? ''))
</script>
