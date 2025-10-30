import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '../types'
import { setToken, setUser, getUser, removeToken } from '../utils/auth'
import { login as loginApi, getProfile, loginWithCode as loginWithCodeApi } from '../api/auth'

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(getUser())
  const token = ref<string | null>(null)

  async function login(username: string, password: string) {
    const res = await loginApi(username, password)
    token.value = res.token
    user.value = res.user
    setToken(res.token)
    setUser(res.user)
    return res
  }

  async function loadProfile() {
    try {
      const res = await getProfile()
      user.value = res.user
      setUser(res.user)
    } catch (error) {
      console.error('Failed to load profile:', error)
      logout()
      throw error
    }
  }

  async function loginWithCode(email: string, code: string) {
    const res = await loginWithCodeApi(email, code)
    token.value = res.token
    user.value = res.user
    setToken(res.token)
    setUser(res.user)
    return res
  }

  function logout() {
    user.value = null
    token.value = null
    removeToken()
  }

  function isAdmin() {
    return user.value?.role === 'admin'
  }

  function isReviewer() {
    return user.value?.role === 'reviewer'
  }

  return {
    user,
    token,
    login,
    loginWithCode,
    logout,
    loadProfile,
    isAdmin,
    isReviewer,
  }
})

