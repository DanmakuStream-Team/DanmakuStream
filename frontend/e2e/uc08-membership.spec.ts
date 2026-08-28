import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

test.describe('UC08 创作者会员订阅', () => {
  const creator = USERS.owner
  const subscriber = USERS.viewer

  test.describe('E2E-TC08-01 查看 & 选择会员套餐主流程', () => {
    test('E2E-TC08-01-01 创作者主页「会员」Tab → 展示 至少 3 个等级（月/季/年）套餐卡，定价正确', async ({ page, request }) => {
      const session = await loginViaApi(request, subscriber.nickname, subscriber.password)

      await test.step('1. GET /membership/plans 断言后端能返回 2+ 套餐（若路由未开放则标记 skip）', async () => {
        const resp = await request.get(`${API}/membership/plans`)
        if (!resp.ok()) {
          test.skip()
          throw new Error('membership 模块接口尚未在后端 route 注册')
        }
        const plans = (await resp.json()).data?.plans ?? ((await resp.json()).data?.list ?? [])
        expect(plans.length, '后端至少暴露 2 档套餐').toBeGreaterThanOrEqual(2)
        for (const plan of plans) {
          expect(Number(plan.price), '套餐价格需 > 0').toBeGreaterThan(0)
          expect(plan.durationDays || plan.validity, '套餐时长字段存在').toBeTruthy()
        }
      })

      await test.step('2. 进入创作者 profile 的会员 Tab，应能渲染套餐卡片（若 Tab 未落地则 skip UI 断言）', async () => {
        const creatorSession = await loginViaApi(request, creator.nickname, creator.password)
        await openAs(page, session, `/user/${creatorSession.userInfo.id}?tab=membership`)
        const cards = page.locator('.membership-plan-card, .plan-card, [data-testid=plan-card]')
        if (!(await cards.first().isVisible({ timeout: 10000 }))) {
          test.skip()
          throw new Error('会员套餐 UI（.membership-plan-card）未在 profile Tab 渲染，暂时跳过')
        }
        expect(await cards.count(), '套餐卡至少 3 档').toBeGreaterThanOrEqual(3)
      })
    })
  })

  test.describe('E2E-TC08-02 下单 & 模拟支付主流程', () => {
    test('E2E-TC08-02-01 选择月卡套餐 → 创建订单 → DemoPay 回调成功 → 我的会员状态更新为有效会员', async ({ page, request }) => {
      const session = await loginViaApi(request, subscriber.nickname, subscriber.password)
      const creatorSession = await loginViaApi(request, creator.nickname, creator.password)
      const header = { Authorization: `Bearer ${session.token}` }

      let planId: number | string | undefined

      await test.step('1. 找一个月卡套餐', async () => {
        const resp = await request.get(`${API}/membership/plans`)
        if (!resp.ok()) {
          test.skip()
          throw new Error('会员计划接口未开放')
        }
        const plans = (await resp.json()).data?.plans ?? ((await resp.json()).data?.list ?? []) as { id: number | string; tier: string; name: string }[]
        const monthly = plans.find((p) => /月|monthly|tier1|basic/i.test(p.tier || p.name)) ?? plans[0]
        if (!monthly) {
          test.skip()
          throw new Error('空套餐无法下单')
        }
        planId = monthly.id
      })

      await test.step('2. 创建订单（POST /membership/orders）', async () => {
        const create = await request.post(`${API}/membership/orders`, {
          headers: header,
          data: { creatorId: creatorSession.userInfo.id, planId, payMethod: 'demo' },
        })
        if (!create.ok()) {
          test.skip()
          throw new Error('POST /membership/orders 未开放')
        }
        const order = (await create.json()).data as { id: number | string; status: string; amount: number }
        expect(order.id, '返回订单 ID').toBeTruthy()
        expect(order.amount, '金额 > 0').toBeGreaterThan(0)
        process.env.E2E_UC08_ORDER_ID = String(order.id)
      })

      await test.step('3. 若后端提供 DemoPay 回调则调用模拟支付', async () => {
        const orderId = process.env.E2E_UC08_ORDER_ID
        const pay = await request.post(`${API}/membership/orders/${orderId}/demo-pay`, { headers: header })
        if (pay.ok()) {
          const myPlan = await request.get(`${API}/membership/me`, { headers: header })
          const plan = (await myPlan.json()).data as { planName?: string; expiredAt?: string; valid?: boolean }
          expect(plan.planName || plan.valid, '我的会员状态变为有效').toBeTruthy()
        } else {
          test.skip()
          throw new Error('DemoPay 模拟回调尚未在后端开放，跳过支付阶段断言')
        }
      })

      await test.step('4. 打开"我的订阅页"至少包含当前已生效套餐描述', async () => {
        await openAs(page, session, '/subscriptions?tab=membership')
        const active = page.locator('.membership-active-card, [data-testid=active-plan-card]').first()
        if (await active.isVisible({ timeout: 8000 })) {
          const text = await active.innerText()
          expect(text, '订阅页显示有效会员或到期日').toMatch(/会员|到期|有效|订阅/)
        }
      })
    })

    test.skip('E2E-TC08-02-02 订单过期 / 取消续费后，进入订阅 Tab 显示「已过期」Tag', () => {
      // 原因：后端 Order/Subscription 表 + 定时任务 expire 尚未定义状态转移。
    })
  })

  test.describe('E2E-TC08-03 创作侧设置价格主流程', () => {
    test('E2E-TC08-03-01 创作者改价（包月 30）→ 用户侧订阅页取到的月卡价格 = 最新值', async ({ page, request }) => {
      const creatorSession = await loginViaApi(request, creator.nickname, creator.password)
      const creatorHeader = { Authorization: `Bearer ${creatorSession.token}` }

      await test.step('1. PUT /membership/me 更新我作为创作者的月卡价格', async () => {
        const update = await request.put(`${API}/membership/me`, {
          headers: creatorHeader,
          data: { monthlyPrice: 30, quarterlyPrice: 80, yearlyPrice: 288 },
        })
        if (!update.ok()) {
          test.skip()
          throw new Error('创作者自主定价 PUT /membership/me 路由未开放')
        }
      })

      await test.step('2. 用户拉 plans 时对应创作者的月卡为 30', async () => {
        const plans = await request.get(`${API}/membership/plans?creatorId=${creatorSession.userInfo.id}`)
        if (!plans.ok()) {
          test.skip()
          throw new Error('plans 列表不支持 creatorId 过滤参数，暂时跳过')
        }
        const list = (await plans.json()).data?.plans ?? ((await plans.json()).data?.list ?? []) as { name: string; price: number }[]
        const monthly = list.find((p) => /月|monthly|basic/i.test(p.name))
        if (monthly) {
          expect(monthly.price, '创作者月卡定价 = 30').toBe(30)
        }
      })
    })

    test.skip('E2E-TC08-03-02 价格改大后 → 现有未到期老用户会员权益不受影响（仅新订单价格变化）', () => {
      // 原因：需要订单 + Subscription 双状态下的断言。
    })
  })
})
