import request from '@/utils/request'
import type { PageResult, UserRole } from '@/types'

export interface InfrastructureMetrics {
  storage: {
    path: string
    usedBytes: number
    totalBytes: number
    freeBytes: number
    usagePercent: number
    warning: boolean
    critical: boolean
  }
  traffic: {
    todayDownBytes: number
    monthDownBytes: number
    source: string
  }
  cpu: {
    usagePercent: number
    warning: boolean
    critical: boolean
    source: string
  }
  online: {
    current: number
    highestConcurrent: number
    liveRoomCount: number
    liveViewerCount: number
    videoConnections: number
  }
}

export interface AdminUserItem {
  id: number
  username: string
  nickname: string
  avatar: string
  bio: string
  role: UserRole | 'moderator'
  followCount: number
  fanCount: number
  videoCount: number
  danmakuCount: number
  createdAt: string
}

export interface SiteBanner {
  id: number
  title: string
  imageUrl: string
  link: string
  enabled: boolean
  sort: number
  createdAt?: string
}

export interface SiteAnnouncement {
  id: number
  content: string
  enabled: boolean
  startedAt?: string
  endedAt?: string
  createdAt?: string
}

export const adminApi = {
  infrastructure() {
    return request.get<InfrastructureMetrics>('/admin/infrastructure')
  },
  users(params: { page: number; pageSize: number; keyword?: string }) {
    return request.get<PageResult<AdminUserItem>>('/admin/users', { params })
  },
  updateUserRole(id: number, role: AdminUserItem['role']) {
    return request.put<{ id: number; role: AdminUserItem['role'] }>(`/admin/users/${id}/role`, { role })
  },
  banners() {
    return request.get<SiteBanner[]>('/admin/banners')
  },
  createBanner(data: Omit<SiteBanner, 'id' | 'createdAt'>) {
    return request.post<SiteBanner>('/admin/banners', data)
  },
  updateBanner(id: number, data: Omit<SiteBanner, 'id' | 'createdAt'>) {
    return request.put<{ id: number }>(`/admin/banners/${id}`, data)
  },
  deleteBanner(id: number) {
    return request.delete<{ id: number }>(`/admin/banners/${id}`)
  },
  announcements() {
    return request.get<SiteAnnouncement[]>('/admin/announcements')
  },
  createAnnouncement(data: Omit<SiteAnnouncement, 'id' | 'createdAt'>) {
    return request.post<SiteAnnouncement>('/admin/announcements', data)
  },
  updateAnnouncement(id: number, data: Omit<SiteAnnouncement, 'id' | 'createdAt'>) {
    return request.put<{ id: number }>(`/admin/announcements/${id}`, data)
  },
  deleteAnnouncement(id: number) {
    return request.delete<{ id: number }>(`/admin/announcements/${id}`)
  },
}
