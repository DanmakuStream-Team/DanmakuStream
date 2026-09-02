export const PASSWORD = 'Test1234!'

export const GATEWAY_URL = process.env.MICRO_E2E_GATEWAY_URL ?? 'http://127.0.0.1:18888'
export const API = process.env.E2E_API_BASE ?? `${GATEWAY_URL}/api/v1`

export const ENGAGEMENT_VIDEO_TITLE = 'E2E-UC05-互动测试视频'

export const USERS = {
  owner:            { nickname: 'e2e-d-owner',        password: PASSWORD },
  viewer:           { nickname: 'e2e-d-viewer',       password: PASSWORD },
  target:           { nickname: 'tuser',              password: PASSWORD },
  moderator:        { nickname: 'tmod',               password: PASSWORD },
  admin:            { nickname: 'tadmin',             password: PASSWORD },
  plain:            { nickname: 'e2eplain',           password: PASSWORD },
  memberCCreator:   { nickname: 'e2e-mc-creator',     password: PASSWORD },
  memberCModerator: { nickname: 'e2e-mc-moderator',   password: PASSWORD },
  memberCPlain:     { nickname: 'e2e-mc-plain',       password: PASSWORD },
} as const
