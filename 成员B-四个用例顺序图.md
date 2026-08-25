# 成员 B：四个用例顺序图

## UC001 用户注册、登录与资料维护

### 需求说明书：系统级顺序图

```mermaid
sequenceDiagram
    autonumber
    actor U as 用户
    participant S as DanmakuStream系统
    alt 注册
        U->>S: 提交昵称和密码
        S->>S: 校验格式与昵称唯一性
        alt 信息合法且昵称可用
            S-->>U: 注册成功（登录凭证、用户资料）
        else 信息非法或昵称重复
            S-->>U: 注册失败原因
        end
    else 登录
        U->>S: 提交昵称和密码
        S->>S: 校验用户凭据
        alt 凭据正确
            S-->>U: 登录成功（登录凭证、用户资料）
        else 凭据错误
            S-->>U: 登录失败
        end
    else 资料维护
        U->>S: 提交昵称、简介或头像修改
        S->>S: 校验身份与资料
        alt 校验通过
            S-->>U: 修改成功（最新资料）
        else 身份失效或资料非法
            S-->>U: 修改失败原因
        end
    end
```

### 概要设计说明书：组件级顺序图

```mermaid
sequenceDiagram
    autonumber
    actor U as 用户
    participant UI as «component» Web前端
    participant GW as «component» API网关
    participant AUTH as «component» 认证管理
    participant PROFILE as «component» 资料管理
    U->>UI: 提交账户操作
    UI->>GW: 注册／登录／资料请求
    alt 注册或登录
        GW->>AUTH: register／login
        AUTH-->>GW: JWT与用户资料／失败原因
    else 修改资料
        GW->>AUTH: validateToken
        AUTH-->>GW: userId／401
        opt 身份有效
            GW->>PROFILE: updateProfile(userId, patch)
            PROFILE-->>GW: 最新资料／校验失败
        end
    end
    GW-->>UI: 操作结果
    UI-->>U: 显示结果
```

### 详细设计说明书：对象级顺序图

```mermaid
sequenceDiagram
    autonumber
    actor U as 用户
    participant H as «boundary» UserHandler
    participant A as «control» AuthService
    participant P as «control» ProfileService
    participant R as «repository» UserRepository
    participant T as «service» TokenService
    U->>H: submit(action, data)
    alt 注册
        H->>A: register(data)
        A->>R: existsByNickname(nickname)
        R-->>A: exists
        alt 昵称可用且数据合法
            A->>R: create(hashedUser)
            R-->>A: savedUser
            A->>T: issue(savedUser)
            T-->>A: JWT
            A-->>H: session
        else 注册数据非法
            A-->>H: validationError
        end
    else 登录
        H->>A: login(data)
        A->>R: findByNickname(nickname)
        R-->>A: user／notFound
        A->>A: comparePassword()
        A->>T: issue(user)
        T-->>A: JWT
        A-->>H: session／invalidCredentials
    else 修改资料
        H->>P: updateMe(userId, patch)
        P->>R: updateProfile(userId, patch)
        R-->>P: updatedUser／error
        P-->>H: profileDTO／validationError
    end
    H-->>U: 操作结果
```

## UC007 关注关系、分组、屏蔽与内容通知

### 需求说明书：系统级顺序图

```mermaid
sequenceDiagram
    autonumber
    actor U as 用户
    participant S as DanmakuStream系统
    U->>S: 提交关系操作（关注／取消／分组／特别关注／屏蔽）
    S->>S: 校验身份、目标用户和操作规则
    alt 不能关注自己／目标不存在
        S-->>U: 拒绝操作并说明原因
    else 请求合法
        S->>S: 保存关系且避免重复记录
        opt 新增关注或特别关注
            S->>S: 生成内容通知订阅关系
        end
        S-->>U: 返回最新关系状态
    end
```

### 概要设计说明书：组件级顺序图

```mermaid
sequenceDiagram
    autonumber
    actor U as 用户
    participant UI as «component» 用户主页
    participant GW as «component» API网关
    participant REL as «component» 关注关系管理
    participant USER as «component» 用户管理
    participant NOTIFY as «component» 通知管理
    U->>UI: 选择关系操作
    UI->>GW: POST /users/{id}/relationship
    GW->>USER: validateUsers(currentId, targetId)
    USER-->>GW: 有效／无效
    alt 用户有效且不是本人
        GW->>REL: changeRelation(command)
        REL-->>GW: 最新关系状态／幂等结果
        opt 新关注或特别关注
            REL->>NOTIFY: subscribeContent(currentId, targetId)
            NOTIFY-->>REL: subscriptionCreated
        end
    end
    GW-->>UI: 操作结果
    UI-->>U: 刷新关注、分组或屏蔽状态
```

### 详细设计说明书：对象级顺序图

```mermaid
sequenceDiagram
    autonumber
    actor U as 用户
    participant H as «boundary» RelationshipHandler
    participant S as «control» RelationshipService
    participant UR as «repository» UserRepository
    participant RR as «repository» RelationshipRepository
    participant NS as «control» NotificationService
    U->>H: changeRelationship(targetId, type, groupId)
    H->>S: execute(currentId, command)
    S->>UR: exists(targetId)
    UR-->>S: exists
    alt 目标不存在或目标是自己
        S-->>H: ruleViolation
    else 请求合法
        S->>RR: find(currentId, targetId, type)
        RR-->>S: relation／none
        S->>RR: upsertOrDelete(command)
        RR-->>S: relationState
        opt 需要内容通知
            S->>NS: createSubscription(relationState)
            NS-->>S: created
        end
        S-->>H: relationshipDTO
    end
    H-->>U: 最新关系状态／失败原因
```

## UC008 创作者会员订阅

### 需求说明书：系统级顺序图

```mermaid
sequenceDiagram
    autonumber
    actor U as 用户
    participant S as DanmakuStream系统
    U->>S: 选择会员方案并提交订阅
    S->>S: 校验创作者、方案、金额和订阅资格
    alt 订阅自己／方案无效／金额不一致
        S-->>U: 拒绝订阅并说明原因
    else 校验通过
        S->>S: 创建待支付订单
        S-->>U: 返回支付信息
        U->>S: 提交支付结果
        S->>S: 幂等校验订单并确认支付
        alt 支付成功且订单有效
            S->>S: 创建或续期有效订阅
            S-->>U: 返回订阅状态和有效期
        else 支付失败或订单失效
            S-->>U: 返回失败原因
        end
    end
```

### 概要设计说明书：组件级顺序图

```mermaid
sequenceDiagram
    autonumber
    actor U as 用户
    participant UI as «component» 会员订阅页
    participant GW as «component» API网关
    participant MEMBER as «component» 会员订阅管理
    participant ORDER as «component» 订单管理
    participant PAY as «component» 支付管理
    U->>UI: 选择方案并确认购买
    UI->>GW: 创建订阅订单
    GW->>MEMBER: validatePlan(userId, creatorId, planId)
    MEMBER-->>GW: 方案与应付金额／规则错误
    opt 校验通过
        GW->>ORDER: createOrder(plan, amount, idempotencyKey)
        ORDER-->>GW: 待支付订单
        GW-->>UI: 支付信息
        UI->>PAY: 演示支付(orderNo)
        PAY->>ORDER: confirmPayment(orderNo, amount)
        ORDER->>MEMBER: activateOrRenew(subscription)
        MEMBER-->>ORDER: 订阅状态与有效期
        ORDER-->>PAY: 支付处理结果
        PAY-->>UI: 最终订阅结果
    end
    UI-->>U: 显示结果
```

### 详细设计说明书：对象级顺序图

```mermaid
sequenceDiagram
    autonumber
    actor U as 用户
    participant H as «boundary» MembershipHandler
    participant S as «control» MembershipService
    participant PR as «repository» PlanRepository
    participant OR as «repository» OrderRepository
    participant PG as «service» PaymentGateway
    participant SR as «repository» SubscriptionRepository
    U->>H: subscribe(creatorId, planId, idempotencyKey)
    H->>S: createSubscriptionOrder(command)
    S->>PR: findActivePlan(planId, creatorId)
    PR-->>S: plan／notFound
    S->>S: validateBuyerAndAmount()
    alt 规则校验失败
        S-->>H: subscriptionRuleError
    else 校验通过
        S->>OR: createOnce(order, idempotencyKey)
        OR-->>S: pendingOrder
        S->>PG: pay(pendingOrder)
        PG-->>S: paymentResult
        alt 支付成功且订单有效
            S->>OR: markPaidOnce(orderNo)
            S->>SR: activateOrRenew(userId, plan)
            SR-->>S: activeSubscription
            S-->>H: subscriptionDTO
        else 支付失败或订单失效
            S->>OR: markFailed(orderNo)
            S-->>H: paymentError
        end
    end
    H-->>U: 订阅状态／失败原因
```

## UC011 用户私信与媒体分享

### 需求说明书：系统级顺序图

```mermaid
sequenceDiagram
    autonumber
    actor A as 发送方
    participant S as DanmakuStream系统
    actor B as 接收方
    A->>S: 发送文字、图片、短视频或平台视频
    S->>S: 校验双方、屏蔽规则、附件与重复请求
    alt 用户无效／被屏蔽／附件非法
        S-->>A: 拒绝发送并说明原因
    else 消息合法
        S->>S: 保存消息并更新会话未读数
        S-->>A: 返回已发送消息
        opt 接收方在线
            S-->>B: 实时推送新消息
        end
        B->>S: 打开会话并读取消息
        S->>S: 更新已读状态和未读数
        S-->>B: 返回会话历史
        S-->>A: 推送已读回执
    end
```

### 概要设计说明书：组件级顺序图

```mermaid
sequenceDiagram
    autonumber
    actor A as 发送方
    participant UI as «component» Chat页面
    participant GW as «component» API网关
    participant CHAT as «component» 私信管理
    participant MEDIA as «component» 媒体管理
    participant PUSH as «component» 实时推送
    actor B as 接收方
    A->>UI: 编辑并发送消息
    UI->>GW: POST /conversations/{id}/messages
    opt 包含附件
        GW->>MEDIA: validateAttachment(file／videoId)
        MEDIA-->>GW: mediaRef／invalid
    end
    GW->>CHAT: sendMessage(senderId, command)
    CHAT-->>GW: message／ruleError
    opt 消息保存成功
        CHAT->>PUSH: pushToRecipient(message)
        PUSH-->>B: 实时新消息
    end
    GW-->>UI: 发送结果
    UI-->>A: 显示消息状态
    B->>PUSH: 读取会话
    PUSH->>CHAT: markRead(conversationId, readerId)
    CHAT-->>PUSH: unreadCount=0
    PUSH-->>A: 已读回执
```

### 详细设计说明书：对象级顺序图

```mermaid
sequenceDiagram
    autonumber
    actor A as 发送方
    participant H as «boundary» MessageHandler
    participant S as «control» MessageService
    participant UR as «repository» UserRepository
    participant BR as «repository» BlockRepository
    participant MR as «repository» MessageRepository
    participant FS as «service» AttachmentStore
    participant WS as «service» RealtimeGateway
    actor B as 接收方
    A->>H: send(recipientId, content, attachment, requestId)
    H->>S: sendMessage(senderId, command)
    S->>UR: exists(recipientId)
    UR-->>S: exists
    S->>BR: isBlocked(senderId, recipientId)
    BR-->>S: blocked
    alt 用户无效或存在屏蔽关系
        S-->>H: messageRuleError
    else 允许发送
        opt 包含附件
            S->>FS: validateAndStore(attachment)
            FS-->>S: mediaRef／invalidAttachment
        end
        S->>MR: createOnce(message, requestId)
        MR-->>S: savedMessage
        S->>MR: incrementUnread(conversationId, recipientId)
        S->>WS: push(recipientId, savedMessage)
        WS-->>B: 新消息事件
        S-->>H: messageDTO
    end
    H-->>A: 发送结果
    B->>H: markRead(conversationId)
    H->>S: markRead(recipientId, conversationId)
    S->>MR: markReadAndClearUnread()
    MR-->>S: readReceipt
    S->>WS: push(senderId, readReceipt)
    WS-->>A: 已读回执
    S-->>H: unreadCount=0
    H-->>B: 已读状态
```
