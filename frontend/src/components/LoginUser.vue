<script setup lang="ts">
import InputGroup from 'primevue/inputgroup'
import InputGroupAddon from 'primevue/inputgroupaddon'
import InputText from 'primevue/inputtext'
import FloatLabel from 'primevue/floatlabel'
import Button from 'primevue/button'

import { computed, ref } from 'vue'
import axios from 'axios'
import type { User } from '@/types/user'
import store from '@/store'
import { useToast } from 'primevue'

interface LoginResponse {
  user: User
  refresh_token: string
  access_token: string
}

const username = ref<string>('')
const password = ref<string>('')
const isDisabled = computed(() => !username.value || !password.value)

const errorMessages = ref<string>('')
const toast = useToast()

const handleLogin = async () => {
  try {
    const response = await axios.post<LoginResponse>('http://localhost:8080/v1/login_user', {
      username: username.value,
      password: password.value,
    })

    store.setUser(response.data.user, response.data.access_token, response.data.refresh_token)
    toast.add({
      severity: 'success',
      summary: `Hello ${response.data.user.full_name}`,
      detail: 'You have successfully logged in!',
      life: 3000,
    })
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (error: any) {
    if (error.response && error.response.status === 404) {
      errorMessages.value = error.response.data.message
    } else {
      errorMessages.value = 'Something went wrong. Please try again later.'
    }
    toast.add({
      severity: 'error',
      summary: 'Login failed',
      detail: errorMessages.value,
      life: 3000,
    })
  }
}
</script>

<template>
  <div class="flex flex-column row-gap-5 justify-center items-center">
    <InputGroup>
      <InputGroupAddon>
        <i class="pi pi-user"></i>
      </InputGroupAddon>
      <FloatLabel variant="on">
        <InputText id="username" type="text" v-model="username" />
        <label for="username">Username</label>
      </FloatLabel>
    </InputGroup>
    <InputGroup>
      <InputGroupAddon>
        <i class="pi pi-lock"></i>
      </InputGroupAddon>
      <FloatLabel variant="on">
        <InputText type="password" id="password" v-model="password" />
        <label for="password">Password</label>
      </FloatLabel>
    </InputGroup>

    <Button label="Login" :disabled="isDisabled" @click="handleLogin" />
  </div>
</template>
