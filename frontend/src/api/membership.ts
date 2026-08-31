import request from '@/utils/request'
import type { UserInfo } from '@/types'

export interface MembershipPlan {
  creatorId: number
  priceCents: number
  benefits: string
  enabled: boolean
}

export interface CreatorSubscriptionInfo {
  id: number
  creator: UserInfo
  priceCents: number
  status: 'active' | 'canceled' | 'expired'
  autoRenew: boolean
  startedAt: string
  expiresAt: string
  daysRemaining: number
}

export interface SubscriptionOrderInfo {
  orderNo: string
  creator: UserInfo
  amountCents: number
  months: number
  status: 'pending' | 'paid' | 'canceled'
  paidAt: string | null
  createdAt: string
}

export const membershipApi = {
  plan(creatorId: number) {
    return request.get<MembershipPlan>(`/creators/${creatorId}/membership-plan`)
  },
  myPlan() {
    return request.get<MembershipPlan>('/creator/membership-plan')
  },
  updateMyPlan(data: Pick<MembershipPlan, 'priceCents' | 'benefits' | 'enabled'>) {
    return request.put<MembershipPlan>('/creator/membership-plan', data)
  },
  status(creatorId: number) {
    return request.get<{ active: boolean; subscription: CreatorSubscriptionInfo | null }>(`/subscriptions/creators/${creatorId}/status`)
  },
  createOrder(creatorId: number, months: number) {
    return request.post<SubscriptionOrderInfo>('/subscriptions/orders', { creatorId, months })
  },
  demoPay(orderNo: string) {
    return request.post<CreatorSubscriptionInfo>(`/subscriptions/orders/${orderNo}/demo-pay`)
  },
  mine() {
    return request.get<{ list: CreatorSubscriptionInfo[] }>('/subscriptions')
  },
  orders() {
    return request.get<{ list: SubscriptionOrderInfo[] }>('/subscriptions/orders')
  },
  cancelAutoRenew(creatorId: number) {
    return request.put(`/subscriptions/${creatorId}/auto-renew`, { enabled: false })
  },
}

export function formatMembershipPrice(cents: number) {
  return `¥${(cents / 100).toFixed(cents % 100 ? 2 : 0)}`
}
