import request from '@/utils/request'

export type ImageUseType = 'common' | 'dynamic' | 'live' | 'cover'

export interface UploadedImage {
  url: string
  type: ImageUseType
  contentType: string
  size: number
  userId: number
}

export const mediaApi = {
  uploadImage(file: File, type: ImageUseType = 'common') {
    const formData = new FormData()
    formData.append('image', file)
    formData.append('type', type)
    return request.post<UploadedImage>('/images/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
}
