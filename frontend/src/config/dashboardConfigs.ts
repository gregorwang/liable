import type { DashboardConfig, StatConfig } from '@/components/GenericReviewDashboard.vue'

/**
 * 质检工作台配置
 */
export const qualityCheckDashboardConfig: DashboardConfig = {
    title: '质检工作台',
    showSearch: true,
    showBatchSubmit: true,
    claimButtonText: '领取质检任务',
    emptyText: '暂无待质检任务，点击「领取质检任务」开始工作',
    stats: [
        { key: 'pending_tasks', label: '待质检任务' },
        { key: 'today_completed', label: '今日已完成' },
        { key: 'total_completed', label: '累计完成' },
        {
            key: 'pass_rate',
            label: '质检通过率',
            format: (value: number) => `${(value || 0).toFixed(1)}%`
        }
    ]
}

/**
 * 二审工作台配置
 */
export const secondReviewDashboardConfig: DashboardConfig = {
    title: '二审工作台',
    showSearch: true,
    showBatchSubmit: true,
    claimButtonText: '领取二审任务',
    emptyText: '暂无待二审任务，点击「领取二审任务」开始工作',
    stats: [
        { key: 'pending_tasks', label: '待二审任务' },
        { key: 'today_completed', label: '今日已完成' },
        { key: 'total_completed', label: '累计完成' }
    ]
}

/**
 * 视频一审工作台配置
 */
export const videoFirstReviewDashboardConfig: DashboardConfig = {
    title: '抖音短视频一审工作台',
    showSearch: true,
    showBatchSubmit: true,
    claimButtonText: '领取新任务',
    emptyText: '暂无待审核任务，点击「领取新任务」开始工作',
    stats: [
        { key: 'pending_tasks', label: '待审核任务' },
        { key: 'today_completed', label: '今日已完成' }
    ]
}

/**
 * 视频二审工作台配置
 */
export const videoSecondReviewDashboardConfig: DashboardConfig = {
    title: '抖音短视频二审工作台',
    showSearch: true,
    showBatchSubmit: true,
    claimButtonText: '领取二审任务',
    emptyText: '暂无待二审任务，点击「领取二审任务」开始工作',
    stats: [
        { key: 'pending_tasks', label: '待二审任务' },
        { key: 'today_completed', label: '今日已完成' }
    ]
}

/**
 * 创建自定义 Dashboard 配置
 */
export function createDashboardConfig(
    title: string,
    options: Partial<DashboardConfig> = {}
): DashboardConfig {
    return {
        title,
        showSearch: true,
        showBatchSubmit: true,
        claimButtonText: '领取任务',
        emptyText: '暂无待审核任务，点击「领取任务」开始工作',
        stats: [],
        ...options
    }
}

/**
 * 创建统计项配置
 */
export function createStatConfig(
    key: string,
    label: string,
    format?: (value: any) => string
): StatConfig {
    return { key, label, format }
}

/**
 * 常用的统计项格式化函数
 */
export const statFormatters = {
    percentage: (value: number) => `${(value || 0).toFixed(1)}%`,
    count: (value: number) => String(value || 0),
    duration: (value: number) => {
        const minutes = Math.floor(value / 60)
        const seconds = value % 60
        return `${minutes}:${seconds.toString().padStart(2, '0')}`
    }
}
