/** Shared E2E data for the acceptance-test suites. */
export const PASSWORD = 'Test1234!'

export const API = process.env.E2E_API_BASE ?? 'http://127.0.0.1:8080/api/v1'

export const USERS = {
  owner: { nickname: 'e2e-d-owner', password: PASSWORD },
  viewer: { nickname: 'e2e-d-viewer', password: PASSWORD },
  domainViewer: { nickname: 'e2e-domain-viewer', password: PASSWORD },
  domainCreator: { nickname: 'e2e-domain-creator', password: PASSWORD },
  domainOther: { nickname: 'e2e-domain-other', password: PASSWORD },
  /** 普通用户（角色可能被 E2E-TC13-02 修改，勿用于权限断言） */
  target: { nickname: 'tuser', password: PASSWORD },
  /** 审核员 */
  moderator: { nickname: 'tmod', password: PASSWORD },
  /** 管理员 */
  admin: { nickname: 'tadmin', password: PASSWORD },
  /** 专用普通用户，用于 UC13 越权断言。 */
  plain: { nickname: 'e2eplain', password: PASSWORD },
} as const
