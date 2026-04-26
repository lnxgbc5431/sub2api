import { describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { nextTick } from 'vue'

import KeysView from '../KeysView.vue'

const { list, getPublicSettings, getAvailable, getUserGroupRates, batchApiKeysUsageStats } = vi.hoisted(() => ({
  list: vi.fn(),
  getPublicSettings: vi.fn(),
  getAvailable: vi.fn(),
  getUserGroupRates: vi.fn(),
  batchApiKeysUsageStats: vi.fn(),
}))

vi.mock('@/api', () => ({
  keysAPI: { list },
  authAPI: { getPublicSettings },
  usageAPI: { batchApiKeysUsageStats },
  userGroupsAPI: { getAvailable, getUserGroupRates },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
    showWarning: vi.fn(),
    showInfo: vi.fn(),
  }),
}))

vi.mock('@/stores/onboarding', () => ({
  useOnboardingStore: () => ({
    isCurrentStep: vi.fn().mockReturnValue(false),
    nextStep: vi.fn(),
  }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true),
  }),
}))

describe('KeysView', () => {
  it('passes hasCustomBaseURL flag to UseKeyModal', async () => {
    list.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 10, pages: 0 })
    getPublicSettings.mockResolvedValue({ api_base_url: 'https://service.example.com' })
    getAvailable.mockResolvedValue([])
    getUserGroupRates.mockResolvedValue({})
    batchApiKeysUsageStats.mockResolvedValue({})

    const wrapper = mount(KeysView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="actions" /><slot name="table" /></div>' },
          DataTable: { template: '<div />' },
          Pagination: true,
          BaseDialog: true,
          ConfirmDialog: true,
          EmptyState: true,
          Select: true,
          SearchInput: true,
          Icon: true,
          EndpointPopover: true,
          GroupBadge: true,
          GroupOptionItem: true,
          Teleport: true,
          UseKeyModal: {
            props: ['show', 'apiKey', 'baseUrl', 'platform', 'hasCustomBaseUrl'],
            template: '<div data-test="use-key-modal" :data-has-custom="String(hasCustomBaseUrl)" :data-base-url="baseUrl" />',
          },
        },
      },
    })

    await flushPromises()
    const setupState = (wrapper.vm as any).$?.setupState
    setupState.selectedKey = {
      key: 'sk-test',
      group: {
        platform: 'openai',
        has_custom_base_url: true,
        allow_messages_dispatch: false,
      },
    }
    setupState.showUseKeyModal = true
    await nextTick()

    const modal = wrapper.get('[data-test="use-key-modal"]')
    expect(modal.attributes('data-has-custom')).toBe('true')
    expect(modal.attributes('data-base-url')).toBe('https://service.example.com')
  })
})
