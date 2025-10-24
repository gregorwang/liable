import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Task, Tag } from '../types'
import { getMyTasks, getTags } from '../api/task'

export const useTaskStore = defineStore('task', () => {
  const tasks = ref<Task[]>([])
  const tags = ref<Tag[]>([])
  const loading = ref(false)

  async function fetchMyTasks() {
    loading.value = true
    try {
      const res = await getMyTasks()
      tasks.value = res.tasks
      return res
    } finally {
      loading.value = false
    }
  }

  async function fetchTags() {
    try {
      const res = await getTags()
      tags.value = res.tags
      return res
    } catch (error) {
      console.error('Failed to fetch tags:', error)
      throw error
    }
  }

  function removeTask(taskId: number) {
    const index = tasks.value.findIndex((t) => t.id === taskId)
    if (index !== -1) {
      tasks.value.splice(index, 1)
    }
  }

  return {
    tasks,
    tags,
    loading,
    fetchMyTasks,
    fetchTags,
    removeTask,
  }
})

