/** Shared E2E data for UC01~UC13.
 *  单环境默认直接连 backend(8080)。当 E2E_MICROSERVICES=1 时改为走网关(默认 18888)。
 *  也可以通过 E2E_GATEWAY_URL / E2E_API_BASE 分别覆盖页面/API 基地址。
 */
export const PASSWORD = 'Test1234!'

const MICRO = process.env.E2E_MICROSERVICES === '1'
export const GATEWAY_URL = process.env.E2E_GATEWAY_URL ?? (MICRO ? 'http://127.0.0.1:18888' : 'http://127.0.0.1:5173')
export const API = process.env.E2E_API_BASE ?? (MICRO ? `${GATEWAY_URL}/api/v1` : 'http://127.0.0.1:8080/api/v1')
export const ENGAGEMENT_VIDEO_TITLE = 'E2E-UC05-互动测试视频'

/** 微服务种子脚本写入的 5 条视频标题前缀，供断言兜底使用（乱码或前端标题变了仍能匹配）。 */
export const VIDEO_TITLE_FALLBACK = /E2E[_\-]MC[_\-]公|E2E-MC|公开视频|待审核|演示视频|分享视频|互动视频|MICRO|MEMBER|UC05|SEED/

export const USERS = {
  owner: { nickname: 'e2e-d-owner', password: PASSWORD },
  viewer: { nickname: 'e2e-d-viewer', password: PASSWORD },
  /** 普通用户（角色可能被 E2E-TC13-02 修改，勿用于权限断言） */
  target: { nickname: 'tuser', password: PASSWORD },
  /** 审核员 */
  moderator: { nickname: 'tmod', password: PASSWORD },
  /** 管理员 */
  admin: { nickname: 'tadmin', password: PASSWORD },
  /** 专用普通用户，用于 UC13 越权断言。 */
  plain: { nickname: 'e2eplain', password: PASSWORD },
  /** 成员 C 内容域专用创作者。 */
  memberCCreator: { nickname: 'e2e-mc-creator', password: PASSWORD },
  /** 成员 C 内容域专用审核员。 */
  memberCModerator: { nickname: 'e2e-mc-moderator', password: PASSWORD },
  /** 成员 C 内容域专用普通用户。 */
  memberCPlain: { nickname: 'e2e-mc-plain', password: PASSWORD },
} as const
