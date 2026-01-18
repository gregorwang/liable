/**
 * Dashboard 配置属性测试
 * 
 * **Property 8: Dashboard 组件配置正确性**
 * **Validates: Requirements 4.1, 4.2, 4.4, 4.5, 4.6**
 */
import { describe, it, expect } from 'vitest'
import fc from 'fast-check'
import {
    qualityCheckDashboardConfig,
    secondReviewDashboardConfig,
    videoFirstReviewDashboardConfig,
    videoSecondReviewDashboardConfig,
    createDashboardConfig,
    createStatConfig,
    statFormatters
} from '../dashboardConfigs'
import type { DashboardConfig, StatConfig } from '@/components/GenericReviewDashboard.vue'

// 所有预定义的 Dashboard 配置
const allConfigs: DashboardConfig[] = [
    qualityCheckDashboardConfig,
    secondReviewDashboardConfig,
    videoFirstReviewDashboardConfig,
    videoSecondReviewDashboardConfig
]

describe('Dashboard Configs - Property Tests', () => {
    /**
     * Property 8.1: 所有预定义配置必须包含必需字段
     * For any predefined dashboard config, it should have title, showSearch, showBatchSubmit, 
     * claimButtonText, emptyText, and stats fields
     */
    it('Property 8.1: all predefined configs should have required fields', () => {
        for (const config of allConfigs) {
            expect(config.title).toBeDefined()
            expect(typeof config.title).toBe('string')
            expect(config.title.length).toBeGreaterThan(0)

            expect(config.showSearch).toBeDefined()
            expect(typeof config.showSearch).toBe('boolean')

            expect(config.showBatchSubmit).toBeDefined()
            expect(typeof config.showBatchSubmit).toBe('boolean')

            expect(config.claimButtonText).toBeDefined()
            expect(typeof config.claimButtonText).toBe('string')
            expect(config.claimButtonText!.length).toBeGreaterThan(0)

            expect(config.emptyText).toBeDefined()
            expect(typeof config.emptyText).toBe('string')
            expect(config.emptyText!.length).toBeGreaterThan(0)

            expect(config.stats).toBeDefined()
            expect(Array.isArray(config.stats)).toBe(true)
        }
    })

    /**
     * Property 8.2: 所有统计项配置必须有 key 和 label
     * For any stat config in any dashboard config, it should have key and label fields
     */
    it('Property 8.2: all stat configs should have key and label', () => {
        for (const config of allConfigs) {
            for (const stat of config.stats || []) {
                expect(stat.key).toBeDefined()
                expect(typeof stat.key).toBe('string')
                expect(stat.key.length).toBeGreaterThan(0)

                expect(stat.label).toBeDefined()
                expect(typeof stat.label).toBe('string')
                expect(stat.label.length).toBeGreaterThan(0)
            }
        }
    })

    /**
     * Property 8.3: createDashboardConfig 工厂函数应该返回有效配置
     * For any valid title string, createDashboardConfig should return a valid config
     */
    it('Property 8.3: createDashboardConfig should return valid config for any title', () => {
        fc.assert(
            fc.property(
                fc.string({ minLength: 1, maxLength: 100 }).filter(s => s.trim().length > 0),
                (title) => {
                    const config = createDashboardConfig(title)

                    // 验证返回的配置包含所有必需字段
                    expect(config.title).toBe(title)
                    expect(config.showSearch).toBe(true)
                    expect(config.showBatchSubmit).toBe(true)
                    expect(config.claimButtonText).toBe('领取任务')
                    expect(config.emptyText).toBeDefined()
                    expect(Array.isArray(config.stats)).toBe(true)

                    return true
                }
            ),
            { numRuns: 100 }
        )
    })

    /**
     * Property 8.4: createDashboardConfig 应该正确合并自定义选项
     * For any valid options, createDashboardConfig should merge them correctly
     */
    it('Property 8.4: createDashboardConfig should merge custom options correctly', () => {
        fc.assert(
            fc.property(
                fc.string({ minLength: 1, maxLength: 50 }).filter(s => s.trim().length > 0),
                fc.boolean(),
                fc.boolean(),
                fc.string({ minLength: 1, maxLength: 50 }).filter(s => s.trim().length > 0),
                (title, showSearch, showBatchSubmit, claimButtonText) => {
                    const config = createDashboardConfig(title, {
                        showSearch,
                        showBatchSubmit,
                        claimButtonText
                    })

                    expect(config.title).toBe(title)
                    expect(config.showSearch).toBe(showSearch)
                    expect(config.showBatchSubmit).toBe(showBatchSubmit)
                    expect(config.claimButtonText).toBe(claimButtonText)

                    return true
                }
            ),
            { numRuns: 100 }
        )
    })

    /**
     * Property 8.5: createStatConfig 应该返回有效的统计项配置
     * For any valid key and label, createStatConfig should return a valid stat config
     */
    it('Property 8.5: createStatConfig should return valid stat config', () => {
        fc.assert(
            fc.property(
                fc.string({ minLength: 1, maxLength: 50 }).filter(s => /^[a-z_]+$/.test(s)),
                fc.string({ minLength: 1, maxLength: 50 }).filter(s => s.trim().length > 0),
                (key, label) => {
                    const stat = createStatConfig(key, label)

                    expect(stat.key).toBe(key)
                    expect(stat.label).toBe(label)
                    expect(stat.format).toBeUndefined()

                    return true
                }
            ),
            { numRuns: 100 }
        )
    })

    /**
     * Property 8.6: createStatConfig 带格式化函数应该正确工作
     * For any valid stat config with format function, the format function should be callable
     */
    it('Property 8.6: createStatConfig with format function should work correctly', () => {
        fc.assert(
            fc.property(
                fc.string({ minLength: 1, maxLength: 50 }).filter(s => /^[a-z_]+$/.test(s)),
                fc.string({ minLength: 1, maxLength: 50 }).filter(s => s.trim().length > 0),
                fc.float({ min: 0, max: 100 }),
                (key, label, value) => {
                    const stat = createStatConfig(key, label, statFormatters.percentage)

                    expect(stat.key).toBe(key)
                    expect(stat.label).toBe(label)
                    expect(typeof stat.format).toBe('function')

                    // 验证格式化函数可以正确调用
                    const formatted = stat.format!(value)
                    expect(typeof formatted).toBe('string')
                    expect(formatted).toContain('%')

                    return true
                }
            ),
            { numRuns: 100 }
        )
    })

    /**
     * Property 8.7: statFormatters.percentage 应该正确格式化任何数字
     * For any number, percentage formatter should return a string ending with %
     */
    it('Property 8.7: percentage formatter should format any number correctly', () => {
        fc.assert(
            fc.property(
                fc.float({ min: -1000, max: 1000, noNaN: true }),
                (value) => {
                    const formatted = statFormatters.percentage(value)

                    expect(typeof formatted).toBe('string')
                    expect(formatted).toMatch(/^-?\d+\.\d%$/)

                    return true
                }
            ),
            { numRuns: 100 }
        )
    })

    /**
     * Property 8.8: statFormatters.count 应该正确格式化任何数字
     * For any number, count formatter should return a string representation
     */
    it('Property 8.8: count formatter should format any number correctly', () => {
        fc.assert(
            fc.property(
                fc.integer({ min: -10000, max: 10000 }),
                (value) => {
                    const formatted = statFormatters.count(value)

                    expect(typeof formatted).toBe('string')
                    expect(formatted).toBe(String(value))

                    return true
                }
            ),
            { numRuns: 100 }
        )
    })

    /**
     * Property 8.9: statFormatters.duration 应该正确格式化秒数
     * For any non-negative number of seconds, duration formatter should return mm:ss format
     */
    it('Property 8.9: duration formatter should format seconds correctly', () => {
        fc.assert(
            fc.property(
                fc.integer({ min: 0, max: 36000 }),
                (seconds) => {
                    const formatted = statFormatters.duration(seconds)

                    expect(typeof formatted).toBe('string')
                    expect(formatted).toMatch(/^\d+:\d{2}$/)

                    // 验证格式化正确
                    const [mins, secs] = formatted.split(':').map(Number)
                    expect(mins * 60 + secs).toBe(seconds)

                    return true
                }
            ),
            { numRuns: 100 }
        )
    })
})

describe('Dashboard Configs - Unit Tests', () => {
    it('qualityCheckDashboardConfig should have pass_rate stat with percentage format', () => {
        const passRateStat = qualityCheckDashboardConfig.stats?.find(s => s.key === 'pass_rate')

        expect(passRateStat).toBeDefined()
        expect(passRateStat?.format).toBeDefined()
        expect(passRateStat?.format!(85.5)).toBe('85.5%')
        expect(passRateStat?.format!(0)).toBe('0.0%')
        expect(passRateStat?.format!(100)).toBe('100.0%')
    })

    it('all configs should have pending_tasks and today_completed stats', () => {
        for (const config of allConfigs) {
            const hasPendingTasks = config.stats?.some(s => s.key === 'pending_tasks')
            const hasTodayCompleted = config.stats?.some(s => s.key === 'today_completed')

            expect(hasPendingTasks).toBe(true)
            expect(hasTodayCompleted).toBe(true)
        }
    })

    it('qualityCheckDashboardConfig should have additional stats', () => {
        const hasTotalCompleted = qualityCheckDashboardConfig.stats?.some(s => s.key === 'total_completed')
        const hasPassRate = qualityCheckDashboardConfig.stats?.some(s => s.key === 'pass_rate')

        expect(hasTotalCompleted).toBe(true)
        expect(hasPassRate).toBe(true)
    })
})
