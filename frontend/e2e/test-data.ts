/** UC13 E2E 固定测试数据。账号由 globalSetup 注册并固化角色，勿在用例中修改 tmod/tadmin。 */
export const PASSWORD = 'Test1234!'

export const USERS = {
  /** 普通用户（角色可能被 E2E-TC13-02 修改，勿用于权限断言） */
  target: { nickname: 'tuser', password: PASSWORD },
  /** 审核员 */
  moderator: { nickname: 'tmod', password: PASSWORD },
  /** 管理员 */
  admin: { nickname: 'tadmin', password: PASSWORD },
  /** 专用普通用户：任何用例都不会改它的角色，用于越权断言 */
  plain: { nickname: 'e2eplain', password: PASSWORD },
} as const

export const API = '/api/v1'
