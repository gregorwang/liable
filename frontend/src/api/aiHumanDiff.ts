import { createTaskApi, AIHumanDiffApiConfig } from './taskApiFactory'
import type {
  AIHumanDiffTask,
  AIHumanDiffTasksResponse,
  SubmitAIHumanDiffRequest
} from '../types'

const baseApi = createTaskApi<AIHumanDiffTask, SubmitAIHumanDiffRequest>(AIHumanDiffApiConfig)

export function claimAIHumanDiffTasks(count: number) {
  return baseApi.claimTasks(count) as Promise<AIHumanDiffTasksResponse>
}

export function getMyAIHumanDiffTasks() {
  return baseApi.getMyTasks() as Promise<AIHumanDiffTasksResponse>
}

export function submitAIHumanDiffReview(review: SubmitAIHumanDiffRequest) {
  return baseApi.submitReview(review)
}

export function submitBatchAIHumanDiffReviews(reviews: SubmitAIHumanDiffRequest[]) {
  return baseApi.submitBatchReviews(reviews)
}

export function returnAIHumanDiffTasks(taskIds: number[]) {
  return baseApi.returnTasks(taskIds)
}
