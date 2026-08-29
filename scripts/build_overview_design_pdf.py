from __future__ import annotations

from pathlib import Path
from datetime import date

from PIL import Image as PILImage
from reportlab.lib import colors
from reportlab.lib.enums import TA_CENTER, TA_LEFT
from reportlab.lib.pagesizes import A4, landscape
from reportlab.lib.styles import ParagraphStyle, getSampleStyleSheet
from reportlab.lib.units import mm
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.platypus import (
    BaseDocTemplate,
    Frame,
    Image,
    KeepTogether,
    PageBreak,
    PageTemplate,
    Paragraph,
    Spacer,
    Table,
    TableStyle,
)
from reportlab.platypus.tableofcontents import TableOfContents


ROOT = Path(__file__).resolve().parents[1]
OUTPUT = ROOT / "docs" / "delivery" / "软件概要设计说明书.pdf"
PAGE_SIZE = A4

NAVY = colors.HexColor("#17365D")
BLUE = colors.HexColor("#2F75B5")
LIGHT_BLUE = colors.HexColor("#D9EAF7")
PALE_BLUE = colors.HexColor("#EEF5FB")
GRAY = colors.HexColor("#5B6573")
LIGHT_GRAY = colors.HexColor("#F4F6F8")
RULE = colors.HexColor("#B7C5D3")
WHITE = colors.white
ORANGE = colors.HexColor("#C55A11")


def register_fonts() -> None:
    font_candidates = [
        ("CN", ["C:/Windows/Fonts/Deng.ttf",
                "/usr/share/fonts/truetype/wqy/wqy-microhei.ttc",
                "/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc"]),
        ("CN-Bold", ["C:/Windows/Fonts/Dengb.ttf",
                     "/usr/share/fonts/truetype/wqy/wqy-microhei.ttc",
                     "/usr/share/fonts/opentype/noto/NotoSansCJK-Bold.ttc"]),
    ]
    for name, candidates in font_candidates:
        for path in candidates:
            if Path(path).exists():
                pdfmetrics.registerFont(TTFont(name, path))
                break
        else:
            raise SystemExit(f"未找到可用中文字体，请安装文泉驿微米黑或 Noto Sans CJK: {name}")


class DesignDocTemplate(BaseDocTemplate):
    def __init__(self, filename: str, **kwargs):
        super().__init__(filename, **kwargs)
        self.section_no = ""

    def afterFlowable(self, flowable):
        if isinstance(flowable, Paragraph):
            style_name = flowable.style.name
            if style_name in ("H1", "H2"):
                level = 0 if style_name == "H1" else 1
                text = flowable.getPlainText()
                key = f"heading-{self.seq.nextf('heading')}"
                self.canv.bookmarkPage(key)
                self.canv.addOutlineEntry(text, key, level=level, closed=False)
                if level == 0:
                    self.notify("TOCEntry", (level, text, self.page, key))
                if level == 0:
                    self.section_no = text.split(" ", 1)[0]


def styles():
    base = getSampleStyleSheet()
    return {
        "CoverTitle": ParagraphStyle(
            "CoverTitle", parent=base["Title"], fontName="CN-Bold", fontSize=29,
            leading=34, textColor=NAVY, alignment=TA_CENTER, spaceAfter=5 * mm,
        ),
        "CoverSub": ParagraphStyle(
            "CoverSub", parent=base["Normal"], fontName="CN", fontSize=15,
            leading=23, textColor=GRAY, alignment=TA_CENTER,
        ),
        "H1": ParagraphStyle(
            "H1", parent=base["Heading1"], fontName="CN-Bold", fontSize=20,
            leading=27, textColor=NAVY, spaceBefore=2 * mm, spaceAfter=6 * mm,
            keepWithNext=True,
        ),
        "H2": ParagraphStyle(
            "H2", parent=base["Heading2"], fontName="CN-Bold", fontSize=14,
            leading=20, textColor=BLUE, spaceBefore=3 * mm, spaceAfter=3 * mm,
            keepWithNext=True,
        ),
        "H3": ParagraphStyle(
            "H3", parent=base["Heading3"], fontName="CN-Bold", fontSize=10.5,
            leading=15, textColor=BLUE, spaceBefore=2 * mm, spaceAfter=2 * mm,
            keepWithNext=True,
        ),
        "Body": ParagraphStyle(
            "Body", parent=base["BodyText"], fontName="CN", fontSize=9.5,
            leading=15, textColor=colors.HexColor("#263445"), spaceAfter=2.3 * mm,
        ),
        "Small": ParagraphStyle(
            "Small", parent=base["BodyText"], fontName="CN", fontSize=8,
            leading=12, textColor=GRAY,
        ),
        "Caption": ParagraphStyle(
            "Caption", parent=base["BodyText"], fontName="CN", fontSize=8.5,
            leading=12, textColor=GRAY, alignment=TA_CENTER, spaceBefore=2 * mm,
        ),
        "Callout": ParagraphStyle(
            "Callout", parent=base["BodyText"], fontName="CN", fontSize=9.2,
            leading=14, textColor=NAVY, borderColor=BLUE, borderWidth=0.7,
            borderPadding=7, backColor=PALE_BLUE, spaceBefore=2 * mm, spaceAfter=4 * mm,
        ),
        "Warning": ParagraphStyle(
            "Warning", parent=base["BodyText"], fontName="CN", fontSize=9.2,
            leading=14, textColor=colors.HexColor("#7F3300"), borderColor=ORANGE,
            borderWidth=0.7, borderPadding=7, backColor=colors.HexColor("#FFF4E8"),
            spaceBefore=2 * mm, spaceAfter=4 * mm,
        ),
        "TOCHeading": ParagraphStyle(
            "TOCHeading", parent=base["Heading1"], fontName="CN-Bold", fontSize=22,
            textColor=NAVY, spaceAfter=8 * mm,
        ),
    }


def p(text: str, style) -> Paragraph:
    return Paragraph(text, style)


def table(data, widths, header=True, font_size=8.3):
    converted = []
    normal = ParagraphStyle("Cell", fontName="CN", fontSize=font_size, leading=font_size + 4, textColor=colors.HexColor("#263445"))
    bold = ParagraphStyle("CellBold", parent=normal, fontName="CN-Bold", textColor=NAVY)
    for r, row in enumerate(data):
        converted.append([Paragraph(str(cell), bold if (header and r == 0) else normal) for cell in row])
    t = Table(converted, colWidths=widths, repeatRows=1 if header else 0, hAlign="LEFT")
    commands = [
        ("FONTNAME", (0, 0), (-1, -1), "CN"),
        ("VALIGN", (0, 0), (-1, -1), "TOP"),
        ("GRID", (0, 0), (-1, -1), 0.45, RULE),
        ("LEFTPADDING", (0, 0), (-1, -1), 6),
        ("RIGHTPADDING", (0, 0), (-1, -1), 6),
        ("TOPPADDING", (0, 0), (-1, -1), 5),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 5),
    ]
    if header:
        commands.extend([
            ("BACKGROUND", (0, 0), (-1, 0), LIGHT_BLUE),
            ("LINEBELOW", (0, 0), (-1, 0), 1.0, BLUE),
        ])
        if len(data) > 1:
            commands.append(("ROWBACKGROUNDS", (0, 1), (-1, -1), [WHITE, LIGHT_GRAY]))
    t.setStyle(TableStyle(commands))
    return t


def diagram(path: Path, max_width: float, max_height: float) -> Image:
    with PILImage.open(path) as im:
        width, height = im.size
    scale = min(max_width / width, max_height / height)
    return Image(str(path), width=width * scale, height=height * scale, hAlign="CENTER")


def page_decor(canvas, doc):
    canvas.saveState()
    width, height = canvas._pagesize
    if doc.page > 1:
        canvas.setStrokeColor(RULE)
        canvas.setLineWidth(0.5)
        canvas.line(18 * mm, height - 13 * mm, width - 18 * mm, height - 13 * mm)
        canvas.setFont("CN", 7.5)
        canvas.setFillColor(GRAY)
        canvas.drawString(18 * mm, height - 10 * mm, "DanmakuStream 软件概要设计说明书")
        canvas.drawRightString(width - 18 * mm, height - 10 * mm, "基线：origin/dev@03c0547")
        canvas.line(18 * mm, 12 * mm, width - 18 * mm, 12 * mm)
        canvas.drawString(18 * mm, 7.7 * mm, "版本 1.0  |  2026-08-28")
        canvas.drawRightString(width - 18 * mm, 7.7 * mm, f"第 {doc.page} 页")
    canvas.restoreState()


UC = [
    dict(no="01", title="用户注册、登录与资料维护", actor="游客、普通用户", trigger="创建账号、登录或维护个人资料", pre="系统可访问；注册昵称未占用；登录账号有效", components="Login/Register/Profile 页面；Axios；路由与鉴权；认证逻辑；资料管理；MySQL；头像存储", endpoints="POST /auth/register；POST /auth/login；GET /auth/me；PUT /users/me；POST /users/me/avatar", data="User、JWT/登录身份、Avatar", flow="提交注册并校验唯一性；登录签发身份；读取并更新本人资料；上传头像后持久化并刷新展示。", exceptions="昵称冲突、错误密码、无效身份、非法头像均返回明确错误，不产生错误写入。"),
    dict(no="02", title="视频发现、搜索与播放", actor="游客、普通用户", trigger="浏览、搜索或选择视频播放", pre="存在审核通过且媒体可访问的视频", components="Home/VideoDetail 页面；videoApi；Gin Router；列表/详情 Handler 与 Logic；MySQL；静态媒体服务", endpoints="GET /videos；GET /videos/:id；GET /search/users；GET /media/*", data="Video、User、播放地址、互动统计", flow="查询公开视频；按关键词筛选；读取详情和作者；浏览器加载媒体资源并播放。", exceptions="空结果正常返回；未发布/不存在视频不可播放；媒体不可用时给出稳定失败状态。"),
    dict(no="03", title="创作者投稿与状态跟踪", actor="创作者", trigger="上传视频并提交审核", pre="已登录；文件满足要求；存储与转码可用", components="VideoUpload/CreatorDashboard；videoApi；鉴权路由；Upload/MeVideos Handler；MySQL；文件系统；FFmpeg/FFprobe", endpoints="POST /videos/upload；POST /videos/:id/cover；GET /users/me/videos；DELETE /videos/:id", data="Video、MediaFile、Cover、ReviewStatus", flow="接收视频与元数据；保存投稿；转码/探测；生成封面；进入待审核；创作中心查询状态。", exceptions="非法格式、取消上传、网络中断和转码失败不应形成可公开播放的残留投稿。"),
    dict(no="04", title="视频审核与发布", actor="审核员、管理员", trigger="处理待审核投稿", pre="操作者已登录并具有审核权限", components="AdminVideosPage；videoApi；Router；Auth/Staff Middleware；AdminList/UpdateStatus Handler；MySQL；媒体存储", endpoints="GET /admin/videos；PUT /admin/videos/:id/status", data="Video、ReviewStatus、Reviewer", flow="列出待审核视频；查看媒体；提交通过或拒绝；保存终态；公开视频查询立即遵循审核状态。", exceptions="普通用户越权被拒绝；重复终态更新受保护；媒体不可访问时不能发布。"),
    dict(no="05", title="视频观看互动", actor="普通用户", trigger="观看视频并发送弹幕、评论、点赞或收藏", pre="已登录；视频存在且审核通过", components="VideoDetail/VideoPlayer；HTTP Client；Gin Router/Auth；Video/Danmaku/Comment Handler；MySQL", endpoints="POST /danmaku；GET /danmaku/:videoId；POST /comments；POST /videos/:id/like；POST /videos/:id/collect", data="Danmaku、Comment、VideoLike、VideoCollect", flow="按播放时间发送弹幕；发布评论；切换点赞与收藏；更新状态和统计；刷新后保持一致。", exceptions="空/超长内容拒绝；屏蔽弹幕不可见；取消和重试采用幂等约束避免重复计数。"),
    dict(no="06", title="个人视频资料库管理", actor="普通用户", trigger="查看或整理个人视频记录", pre="已登录；产生过观看、收藏或稍后再看记录", components="UserLibrary/VideoPlayer；libraryApi/videoApi；Router/Auth；Library/Video Handler；MySQL；本地缓存", endpoints="GET/PUT/DELETE /users/me/history；GET/POST/DELETE /users/me/watch-later；GET /users/me/collections", data="WatchHistory、WatchLater、Like、Collect、Collection", flow="记录进度；加入分类；查询资料库；恢复进度；移除单项或清空列表。", exceptions="重复加入不产生重复记录；不可用视频标记异常；多设备以服务端记录为准。"),
    dict(no="07", title="关注关系与内容通知", actor="普通用户、创作者", trigger="关注创作者并接收相关内容", pre="已登录；目标创作者存在", components="用户主页/订阅页；逻辑路由边界；用户与关系管理；通知管理；MySQL", endpoints="POST /users/:id/follow；GET/POST/PUT/DELETE /users/follow-groups；PUT /users/:id/follow-settings；GET /notifications", data="Follow、FollowGroup、UserBlock、Notification", flow="建立关注；分组或特别关注；内容发布后生成动态/通知；订阅页查询。", exceptions="不能关注自己；重复关注幂等；屏蔽后限制交互；无效目标拒绝。"),
    dict(no="08", title="创作者会员订阅", actor="普通用户、创作者", trigger="配置会员方案或订阅创作者", pre="双方已登录；创作者已配置可用方案", components="Subscription/Profile 页面；逻辑路由边界；会员、订单、支付管理；到期任务；MySQL", endpoints="PUT /creator/membership-plan；POST /subscriptions/orders；POST /subscriptions/orders/:orderNo/demo-pay；PUT /subscriptions/:creatorId/auto-renew", data="MembershipPlan、SubscriptionOrder、CreatorSubscription", flow="配置方案；创建订单；演示支付；事务内激活订阅；查询状态并设置自动续订。", exceptions="未开放方案、自我订阅、失效订单、金额不符被拒绝；支付请求保持幂等。"),
    dict(no="09", title="直播预约与用户预约", actor="主播、普通用户", trigger="发布未来直播计划或预约直播", pre="双方已登录；计划时间合法", components="LiveListPage；liveApi；Router/Auth；Schedule Handler；Schedule Worker；MySQL", endpoints="POST/GET /live-schedules；DELETE /live-schedules/:id；POST /live-schedules/:id/reserve", data="LiveSchedule、LiveReservation、Notification、LiveRoom", flow="主播发布计划；系统展示；用户预约；更新人数；定时任务在到点时处理计划。", exceptions="过去时间和同主播冲突拒绝；重复预约幂等；双方可按权限取消。"),
    dict(no="10", title="直播发布、观看与实时互动", actor="主播、普通用户、SRS", trigger="主播开播或观众进入直播间", pre="已登录；SRS 可用；主播拥有直播间", components="LiveStudio/LiveRoom；liveApi；Live/Interaction Handler；WebSocket Handler；Live Hub；MySQL；SRS", endpoints="POST/GET/PUT /live*；GET /ws/live/:id；GET /ws/live-publish/:id；RTMP/HLS", data="LiveRoom、LiveLike、LiveGift、Danmaku、在线连接", flow="创建直播；获取推流信息；SRS 接收媒体；观众播放；WebSocket 互动；下播清理状态。", exceptions="未就绪时等待；非房主禁止管理；连接重连不重复计数；持久化失败不广播。"),
    dict(no="11", title="用户私信与媒体分享", actor="普通用户", trigger="向另一用户发送私信或媒体", pre="发送者已登录；双方用户有效", components="ChatPage；逻辑路由边界；消息管理；媒体服务；Chat Hub；通知管理；MySQL/附件存储", endpoints="GET /messages/conversations；GET/POST /messages；PUT /messages/:userId/read；POST /messages/media；GET /ws/chat", data="ChatMessage、Conversation、Attachment、UnreadCount", flow="打开会话；上传可选附件；校验并持久化消息；在线实时推送；读取后更新未读数。", exceptions="空消息、未知类型、越权附件、无效/屏蔽用户被拒绝；断线后从历史恢复。"),
    dict(no="12", title="创作者数据分析", actor="创作者", trigger="查看内容与粉丝表现", pre="已登录并拥有可统计内容或历史数据", components="CreatorDashboard/MetricLineChart；Axios；Router/Auth；AnalyticsHandler；MySQL 聚合表", endpoints="GET /creator/analytics?range=...", data="Video、CreatorDailyStat、VideoDailyStat、汇总指标", flow="校验身份；按范围聚合作品、播放、互动和粉丝指标；返回汇总与趋势；图表渲染。", exceptions="无数据返回零值和空趋势；越权拒绝；部分数据不可用时明确降级。"),
    dict(no="13", title="平台审核、权限、运营与基础设施管理", actor="审核员、管理员", trigger="治理内容、调整权限、配置运营信息或检查状态", pre="已登录并具有对应后台角色", components="后台 Vue 页面；Nginx；Gin Router；Auth/Staff/Admin Middleware；Video/Danmaku/Admin Handler；Metrics；MySQL", endpoints="/admin/videos；/admin/danmaku；/admin/users；/admin/banners；/admin/announcements；/admin/infrastructure", data="Video、Danmaku、UserRole、Banner、Announcement、InfrastructureMetric", flow="按角色进入后台；审核内容；管理角色与运营配置；采集并展示基础设施指标。", exceptions="角色边界严格区分；非法参数拒绝；指标不可用时返回明确状态。"),
]


def build_story(st):
    story = []
    width, height = PAGE_SIZE

    story += [Spacer(1, 5 * mm), p("DanmakuStream", st["CoverSub"]), Spacer(1, 3 * mm),
              p("软件概要设计说明书", st["CoverTitle"]),
              p("UC01-UC13 全业务场景概要设计", st["CoverSub"]), Spacer(1, 6 * mm)]
    story.append(table([
        ["文档属性", "内容"],
        ["版本", "1.0"],
        ["编制日期", "2026-08-28"],
        ["代码与文档基线", "origin/dev@03c0547"],
        ["覆盖范围", "REQ01/UC01 至 REQ13/UC13"],
        ["图形规范", "PlantUML 源文件及同名 PNG/SVG；本文嵌入 PNG"],
    ], [42 * mm, 92 * mm]))
    story += [Spacer(1, 5 * mm), p("本说明书描述当前 Vue 3 + Gin 单体运行基线，并给出改造后 Kubernetes 目标部署。用例级概要设计以 COMP-SEQ01 至 COMP-SEQ13 为统一编号基线。", st["Callout"]), Spacer(1, 3 * mm)]

    story.append(p("目录", st["TOCHeading"]))
    toc = TableOfContents()
    toc.levelStyles = [
        ParagraphStyle("TOC0", fontName="CN-Bold", fontSize=9.2, leading=13, textColor=NAVY, leftIndent=0, firstLineIndent=0, spaceBefore=1),
        ParagraphStyle("TOC1", fontName="CN", fontSize=8, leading=11, textColor=GRAY, leftIndent=10, firstLineIndent=0),
    ]
    story += [toc, Spacer(1, 5 * mm)]

    story += [p("1 引言", st["H1"]), p("1.1 编写目的", st["H2"]),
              p("本文说明 DanmakuStream 程序系统的基本处理流程、组织结构、模块划分、功能分配、接口、数据、运行、安全和异常处理设计，并把经确认的 UC01-UC13 落实为组件级协作方案，为详细设计、编码、测试、验收和维护提供概要层基线。预期读者包括课程教师、开发组成员、测试人员和后续维护人员。", st["Body"]),
              p("1.2 背景", st["H2"]),
              table([
                  ["项目项", "说明"],
                  ["系统名称", "DanmakuStream 在线视频与直播平台。"],
                  ["任务提出者", "软件工程基础实践课程教学方。"],
                  ["开发者", "DanmakuStream-Team 项目小组。"],
                  ["主要用户", "游客、普通用户、创作者/主播、审核员和管理员。"],
                  ["运行位置", "本地开发机、Docker Compose 演示环境及目标 Kubernetes 集群。"],
              ], [37 * mm, 137 * mm]),
              p("1.3 定义", st["H2"]),
              table([["术语", "含义"], ["SDD", "Software Design Description，软件概要设计说明书。"], ["组件", "具有明确职责和接口边界的页面、后端模块、中间件或外部服务。"], ["逻辑路由边界", "顺序图中标注的 API Gateway 表示统一 API 路由职责；当前单体基线由 Nginx + Gin Router 承担，并非独立 Gateway 容器。"], ["组件级顺序图", "描述页面、路由、中间件、Handler/Logic、数据库和外部服务之间的调用顺序。"], ["SRS", "Simple Realtime Server，负责 RTMP/HLS 直播媒体处理。"]], [37 * mm, 137 * mm]),
              p("1.4 参考设计", st["H2"]),
              table([
                  ["编号", "参考文件", "用途"],
                  ["REF-01", "docs/project/use-case-catalog.md", "UC01-UC13 最终验收范围与用例说明。"],
                  ["REF-02", "软件需求规格说明书.pdf", "参与者、功能需求、系统级模型与概念模型。"],
                  ["REF-03", "软件详细设计说明书.pdf", "实现类、对象协作和方法级设计。"],
                  ["REF-04", "软件工程基础实践-2026夏.pdf", "课程交付、测试、CI/CD 与验收要求。"],
                  ["REF-05", "docs/models/ 与 docs/traceability/master.md", "正式 PlantUML 图件和统一追溯关系。"],
              ], [28 * mm, 71 * mm, 75 * mm]), Spacer(1, 5 * mm)]

    story += [p("2 系统需求概述与总体设计", st["H1"]), p("2.1 需求规定", st["H2"]),
              p("系统接收账号资料、视频和图片文件、搜索条件、播放行为、互动消息、直播计划、推流状态及后台管理操作，输出公开视频列表、播放资源、互动结果、直播状态、通知、分析指标和管理处理结果。所有输出必须与权限、审核状态和持久化结果一致。", st["Body"]),
              table([
                  ["类别", "主要项目", "处理要求"],
                  ["账号与关系", "注册资料、登录凭据、关注/屏蔽、会员订单", "唯一性、身份校验、关系幂等、订单事务。"],
                  ["内容与媒体", "视频、封面、头像、消息附件、搜索关键词", "类型/大小校验、审核状态、媒体归属和可访问性。"],
                  ["互动与直播", "弹幕、评论、点赞、预约、礼物、RTMP/WebSocket 事件", "实时性、去重、先持久化后广播、断线恢复。"],
                  ["管理与分析", "审核决定、角色、横幅公告、日期范围", "角色边界、终态保护、统计口径稳定。"],
              ], [34 * mm, 78 * mm, 62 * mm]),
              p("2.2 业务目标", st["H2"]),
              p("系统围绕视频消费、创作者投稿、社区互动、直播互动、用户关系、会员订阅和平台治理形成完整业务闭环。最终验收范围固定为 UC01-UC13，每个用例均应具备可运行流程、组件设计、代码实现、测试及追溯证据。", st["Body"]),
              p("2.3 运行环境", st["H2"]),
              p("2.3.1 设备", st["H3"]),
              table([
                  ["设备", "建议配置", "说明"],
                  ["开发/演示主机", "4 核 CPU、8 GB 内存、20 GB 以上可用磁盘", "运行前后端、MySQL、SRS 和自动化测试。"],
                  ["用户终端", "现代桌面浏览器及稳定网络", "访问页面、点播 HLS/MP4 和建立 WebSocket。"],
                  ["主播终端", "OBS 或兼容 RTMP 推流工具", "向 SRS 推送直播媒体流。"],
              ], [39 * mm, 68 * mm, 67 * mm]),
              p("2.3.2 支持软件", st["H3"]),
              table([
                  ["类别", "当前运行基线", "用途"],
                  ["客户端", "Chrome / Edge；Vue 3 + Vite", "页面、播放器、HTTP 与 WebSocket 客户端。"],
                  ["服务端", "Go、Gin、GORM", "REST API、鉴权、业务处理、实时连接。"],
                  ["数据与媒体", "MySQL 8、文件卷、FFmpeg/FFprobe、SRS", "数据持久化、转码探测、RTMP/HLS 分发。"],
                  ["部署与构建", "Docker Compose、Nginx、GitHub Actions", "本地演示、反向代理、编译测试和镜像构建。"],
              ], [34 * mm, 78 * mm, 62 * mm]),
              p("2.4 设计约束", st["H2"]),
              p("前端统一使用 /api/v1 同源路径；受保护接口经过 JWT 中间件；媒体大文件不进入数据库；实时消息采用 WebSocket；当前可运行基线仍是 Gin 单体，不把目标 Kubernetes 服务图表述成已经落地的微服务。", st["Body"]),
              p("2.5 功能需求", st["H2"]),
              table([
                  ["功能域", "覆盖用例", "核心处理"],
                  ["用户与关系", "UC01、UC06、UC07、UC08、UC11", "身份、资料库、关注分组、会员订单、私信与通知。"],
                  ["视频内容", "UC02、UC03、UC04、UC05、UC12", "发现播放、投稿处理、审核发布、互动和创作者分析。"],
                  ["直播", "UC09、UC10", "计划预约、推流播放、在线状态和实时互动。"],
                  ["平台治理", "UC13（并复用 UC04 审核）", "角色、内容治理、运营配置和基础设施指标。"],
              ], [39 * mm, 48 * mm, 87 * mm]),
              p("2.6 非功能需求", st["H2"]),
              p("系统应具备分级授权、事务一致性、稳定错误响应、列表分页、实时连接恢复、日志可观察性和容器化部署能力。性能指标按演示数据量设定：登录与普通写操作 1 秒级，列表/搜索 2 秒级，页面首屏 3 秒内形成可操作状态，实时消息保持可感知实时。", st["Body"]),
              p("2.7 系统总体架构", st["H2"]),
              p("2.7.1 当前单体部署基线", st["H3"]),
              p("浏览器从 frontend/Nginx 获取 Vue 静态资源；/api、/media、/ws 由反向代理转发至 Gin 单体应用；MySQL 保存业务数据；SRS 负责 RTMP/HLS 媒体链路；共享卷保存视频、封面、头像及消息附件。", st["Body"]),
              diagram(ROOT / "docs/models/deployment/DEPLOY-MONO-DETAILED.png", 174 * mm, 104 * mm),
              p("图 2-1 改造前 Docker Compose 单体部署图（DEPLOY-MONO-DETAILED）", st["Caption"]), PageBreak(),
              p("2.7.2 改造后 Kubernetes 部署", st["H3"]),
              p("目标部署以 Ingress/LoadBalancer 为统一入口，API Gateway 负责 JWT 与路由，业务拆为 user、content、engagement 三个服务；Redis 承担实时 Hub、缓存和限流，持久化数据按 Schema 所有权隔离。该图是演进设计，不表示当前仓库已经具备独立 Gateway 容器或微服务版 E2E 环境。", st["Body"]),
              diagram(ROOT / "docs/models/deployment/DEPLOY-K8S.png", 174 * mm, 116 * mm),
              p("图 2-2 改造后 Kubernetes 部署图（DEPLOY-K8S）", st["Caption"]), PageBreak()]

    main_story = story
    story = []
    story += [p("6 用例与组件概要设计", st["H1"]),
              p("6.1 组件职责", st["H2"]),
              table([
                  ["层次", "核心组件", "职责"],
                  ["表现层", "Vue 页面、通用组件、Axios、WebSocket Client", "采集输入、展示业务状态、媒体播放及实时事件。"],
                  ["接入层", "Nginx、Gin Router、Auth/Staff/Admin Middleware", "统一路径、身份校验、角色鉴权、请求分派。"],
                  ["业务层", "auth/user/video/live/message/membership/admin Handler 与 Logic", "执行 UC01-UC13 的业务规则和事务边界。"],
                  ["实时与任务", "Chat/Live Hub、Schedule/Membership Worker", "实时广播、在线状态、预约启动与订阅到期处理。"],
                  ["数据与媒体", "MySQL、文件卷、FFmpeg、SRS", "业务持久化、媒体处理、RTMP/HLS 分发。"],
              ], [28 * mm, 63 * mm, 83 * mm]),
              Spacer(1, 4 * mm),
              p("6.2 全系统总体组件图", st["H2"]),
              p("总体组件图以当前可运行实现为准：前端通过 Nginx 进入 Gin Router，业务在同一 Gin 进程内按用户与社交、视频内容、直播、平台治理四个逻辑组件组织，并共享鉴权中间件、实时 Hub、后台任务和 GORM 数据访问层。图中的组件边界用于说明职责与依赖，不表示已经拆成可独立部署的微服务。", st["Body"]),
              diagram(ROOT / "docs/models/component/COMPONENT-OVERVIEW.png", 174 * mm, 81 * mm),
              p("图 6-1 DanmakuStream 全系统总体组件图（UC01-UC13，COMPONENT-OVERVIEW）", st["Caption"]),
              p("6.3 用户、关系、会员与消息域组件", st["H2"]),
              diagram(ROOT / "docs/models/component/COMPONENT-B.png", 174 * mm, 96 * mm),
              p("图 6-2 当前单体组件图（UC01/07/08/11，COMPONENT-B）", st["Caption"]),
              p("6.4 视频互动与直播域组件", st["H2"]),
              diagram(ROOT / "docs/models/component/COMPONENT-D.png", 174 * mm, 78 * mm),
              p("图 6-3 当前单体组件图（UC05/09/10，COMPONENT-D）", st["Caption"]),
              p("6.5 关键设计原则", st["H2"])]
    story.append(table([
        ["原则", "设计约束"],
        ["鉴权集中", "受保护接口统一经过 AuthMiddleware；后台接口继续经过 Staff/Admin 权限中间件。"],
        ["事务一致性", "点赞、收藏、预约、支付和已读更新在数据库事务或唯一约束下保持幂等。"],
        ["媒体与业务分离", "SRS/FFmpeg 处理媒体；Gin Handler/Hub 管理身份、状态和互动。"],
        ["实时先持久化", "私信和直播弹幕在持久化成功后再广播；失败时返回错误事件。"],
        ["接口稳定", "前端使用 /api/v1 同源相对路径；部署层决定其转发至 Gin Router 或未来 Gateway。"],
        ["可观测失败", "权限、参数、媒体、数据库与实时连接异常均提供可验证结果。"],
    ], [35 * mm, 139 * mm]))
    story.append(PageBreak())

    story += [p("6.6 用例概要设计（UC01-UC13）", st["H2"]),
              p("每个用例给出参与者、触发/前置条件、组件、接口、数据、关键流程和组件级顺序图。图中 API Gateway 字样按“逻辑路由边界”解释；当前运行实现仍是 Nginx + Gin Router。", st["Callout"])]

    for index, uc in enumerate(UC, start=1):
        no = uc["no"]
        story += [p(f"6.6.{index} UC{no} {uc['title']}", st["H2"])]
        story.append(table([
            ["设计项", "概要说明"],
            ["需求/用例编号", f"REQ{no} / UC{no} / COMP-SEQ{no}"],
            ["参与者", uc["actor"]],
            ["触发条件", uc["trigger"]],
            ["前置条件", uc["pre"]],
            ["参与组件", uc["components"]],
            ["主要接口", uc["endpoints"]],
            ["主要数据", uc["data"]],
            ["主成功流程", uc["flow"]],
            ["备选/异常与质量约束", uc["exceptions"]],
        ], [37 * mm, 137 * mm], font_size=8.0))
        story += [Spacer(1, 2 * mm),
                  p(f"6.6.{index}.1 COMP-SEQ{no} 组件级顺序图", st["H3"]),
                  diagram(ROOT / f"docs/models/component/COMP-SEQ{no}.png", 174 * mm, 104 * mm),
                  p(f"图 6-{index + 3} UC{no} {uc['title']}组件级顺序图（源文件：docs/models/component/COMP-SEQ{no}.puml）", st["Caption"])]
        if index != len(UC):
            story.append(PageBreak())

    design_story = story
    story = main_story
    story += [p("3 系统接口与数据结构设计", st["H1"]), p("3.1 用户接口", st["H2"])]
    story.append(table([
        ["界面类别", "主要页面", "对应业务"],
        ["通用与用户端", "LoginPage、RegisterPage、HomePage、UserProfilePage、SubscriptionPage", "登录注册、视频发现、资料、关注和会员。"],
        ["视频与创作者端", "VideoDetailPage、VideoUploadPage、UserLibraryPage、CreatorDashboardPage", "播放互动、投稿、资料库和数据分析。"],
        ["直播与消息端", "LiveListPage、LiveStudioPage、LiveRoomPage、ChatPage", "预约、开播观看、实时互动和私信。"],
        ["管理端", "AdminVideosPage、AdminUsersPage、AdminOperationsPage、AdminInfrastructurePage、AdminDanmakuPage", "内容审核、权限、运营和基础设施管理。"],
    ], [36 * mm, 82 * mm, 56 * mm]))
    story += [p("页面统一通过 Axios 访问 /api/v1，通过 WebSocket Client 访问 /ws；成功、参数错误、权限不足、资源不存在和服务异常均应显示明确反馈，不把后端原始错误堆栈暴露给用户。", st["Body"]),
              p("3.2 与其他软件、硬件接口", st["H2"]),
              p("3.2.1 支持软件与外部接口", st["H3"])]
    story.append(table([
        ["接口", "协议/方式", "用途"],
        ["MySQL", "GORM / TCP", "业务实体、关系、订单、消息、直播及运营数据持久化。"],
        ["文件存储与 FFmpeg", "文件系统 / 子进程", "视频、封面、头像和附件存储，媒体探测与转码。"],
        ["SRS", "RTMP / HTTP-HLS", "主播推流接收和观众直播播放。"],
        ["Nginx", "HTTP / HTTPS / WebSocket", "静态资源服务及 /api、/media、/ws 反向代理。"],
        ["Docker Compose / Kubernetes", "容器网络与配置", "当前单体编排和目标集群部署。"],
    ], [43 * mm, 48 * mm, 83 * mm]))
    story += [p("3.2.2 系统内部接口分组", st["H3"])]
    story.append(table([
        ["接口组", "服务范围", "覆盖用例"],
        ["/api/v1/auth、/users", "认证、资料、关系、资料库", "UC01、UC06、UC07"],
        ["/api/v1/videos、/danmaku、/comments", "视频发现、投稿、审核、互动", "UC02-UC05"],
        ["/api/v1/subscriptions、/creator", "会员方案、订单、分析", "UC08、UC12"],
        ["/api/v1/live、/live-schedules", "直播计划、房间、互动", "UC09、UC10"],
        ["/api/v1/messages、/notifications", "私信、附件、未读与通知", "UC07、UC11"],
        ["/api/v1/admin", "审核、角色、运营与基础设施", "UC04、UC13"],
        ["/ws/live、/ws/live-publish、/ws/chat", "直播与私信实时通道", "UC10、UC11"],
    ], [44 * mm, 85 * mm, 45 * mm]))
    story += [Spacer(1, 4 * mm), p("3.3 系统数据结构设计", st["H2"]), p("3.3.1 逻辑结构设计要点", st["H3"])]
    story.append(table([
        ["领域", "核心数据", "一致性策略"],
        ["用户与关系", "User、Follow、FollowGroup、UserBlock、Notification", "唯一关系约束；本人/角色权限校验。"],
        ["内容与媒体", "Video、MediaFile、WatchHistory、Collection", "审核状态控制公开性；媒体文件与记录协同清理。"],
        ["互动", "Danmaku、Comment、VideoLike、VideoCollect", "联合唯一约束；计数与关系事务更新。"],
        ["直播", "LiveRoom、LiveSchedule、LiveReservation、LiveGift", "计划冲突校验；在线用户去重；下播状态清理。"],
        ["会员与消息", "MembershipPlan、Order、Subscription、ChatMessage", "支付幂等；先持久化后实时广播。"],
        ["运营与指标", "Banner、Announcement、Role、DailyStat", "管理员权限；统计口径和日期范围固定。"],
    ], [35 * mm, 88 * mm, 51 * mm]))
    story += [p("3.3.2 物理结构设计要点", st["H3"]),
              p("关系表使用 InnoDB 和 utf8mb4；主键采用稳定编号；高频筛选字段建立普通或联合索引；关注、点赞、收藏、预约等关系建立联合唯一约束；密码只保存安全哈希；大文件只在数据库保存受控路径和元数据；删除与审核状态采用明确枚举，避免把不可公开内容错误返回给游客。", st["Body"]),
              p("3.3.3 数据结构与程序关系", st["H3"]),
              table([
                  ["数据域", "主要程序", "访问方式"],
                  ["用户/关系/会员", "auth、user、membership、notification Handler/Logic", "GORM 查询与事务；唯一关系和订单幂等约束。"],
                  ["视频/互动", "video、danmaku、comment、collection Handler/Logic", "GORM + 文件系统；审核状态控制媒体公开。"],
                  ["直播", "live Handler、Schedule Worker、Live Hub", "GORM + SRS + WebSocket；状态迁移和房间广播。"],
                  ["消息", "message/media Handler、Chat Hub", "GORM + 附件目录；持久化后实时推送。"],
                  ["管理/分析", "admin、creator Handler", "角色限定查询、聚合统计和系统指标采集。"],
              ], [39 * mm, 75 * mm, 60 * mm]), PageBreak()]

    story += [p("4 系统数据库设计", st["H1"]),
              p("系统使用 MySQL 8，业务模型由 GORM 管理。数据库设计以当前代码模型为准，不沿用旧版 Spring/MyBatis 或虚构 Redis 表结构。媒体文件存储在受控目录，数据库只保存路径、归属、状态和业务元数据。", st["Body"]),
              p("4.1 主要数据表与关系", st["H2"]),
              table([
                  ["数据组", "代表性表/模型", "关键关系与约束"],
                  ["用户与权限", "users、follows、follow_groups、user_blocks", "用户编号为关系中心；关注/屏蔽组合唯一；角色决定后台权限。"],
                  ["视频与资料库", "videos、watch_histories、watch_laters、collections", "视频归属创作者；审核状态决定公开性；用户-视频记录去重。"],
                  ["互动", "danmakus、comments、video_likes、video_collects", "均关联用户和视频；点赞/收藏采用联合唯一约束。"],
                  ["直播", "live_rooms、live_schedules、live_reservations、live_likes、live_gifts", "房间归属主播；计划冲突校验；预约和点赞关系去重。"],
                  ["会员与消息", "membership_plans、subscription_orders、creator_subscriptions、chat_messages", "订单号唯一、支付幂等；消息保存发送者/接收者和附件归属。"],
                  ["运营与分析", "notifications、banners、announcements、creator_daily_stats", "通知按用户读取；运营内容受管理员控制；统计按日期和主体唯一。"],
              ], [36 * mm, 70 * mm, 68 * mm]),
              p("4.2 索引、事务与保密", st["H2"]),
              table([
                  ["设计点", "概要规定"],
                  ["索引", "主键索引覆盖所有实体；外键、状态、创建时间和常用列表条件建立索引；关系表建立联合唯一索引。"],
                  ["事务", "支付、关系切换、预约、互动统计和已读更新在同一事务边界完成；失败回滚。"],
                  ["并发", "使用唯一约束和条件更新抵御重复请求；定时任务按状态筛选并保持幂等。"],
                  ["保密", "密码仅保存安全哈希；JWT 密钥和数据库凭据通过环境变量/Secret 注入；私信只允许会话双方访问。"],
                  ["备份恢复", "演示环境保存 MySQL 数据卷和媒体卷；恢复时先数据库后应用，并校验媒体路径一致性。"],
              ], [39 * mm, 135 * mm]), PageBreak()]

    story += [p("5 运行设计", st["H1"]), p("5.1 非功能与运行质量", st["H2"])]
    story.append(table([
        ["属性", "概要设计"],
        ["安全性", "JWT 身份校验；普通/审核员/管理员分级授权；上传媒体归属与类型校验；Secret 不进入镜像。"],
        ["可靠性", "数据库失败不伪造成功；实时连接断开后可恢复历史；定时任务幂等；媒体未就绪时有限重试。"],
        ["性能", "列表分页；统计聚合表；WebSocket Hub 广播；目标部署使用 Redis 与 HPA。"],
        ["可维护性", "统一 /api/v1；Handler/Logic/Model 分层；PlantUML 源图与导出物同步；编号贯穿追溯。"],
        ["可测试性", "组件边界可由单元、API/集成和 Playwright E2E 验证；CI 失败阻断镜像构建。"],
        ["可部署性", "当前 Docker Compose 单体可运行；Kubernetes 图为目标形态，Gateway 与微服务 E2E 尚待落地。"],
    ], [37 * mm, 137 * mm]))
    story += [p("5.2 运行模块组合", st["H2"]),
              table([
                  ["运行组合", "参与模块", "主要职责"],
                  ["用户基础组合", "auth/user Handler、AuthMiddleware、MySQL、头像存储", "注册登录、资料、关系和身份校验。"],
                  ["视频消费组合", "video/danmaku/comment Handler、媒体存储、FFmpeg", "发现、播放、投稿、审核和观看互动。"],
                  ["直播互动组合", "live Handler、WebSocket Handler、Hub、SRS、Worker", "预约、推流、播放、实时互动与在线状态。"],
                  ["会员消息组合", "membership/message/notification、Chat Hub", "方案订单、订阅、私信和通知。"],
                  ["管理分析组合", "admin/creator Handler、权限中间件、统计查询", "内容治理、角色运营、指标和创作者分析。"],
              ], [38 * mm, 79 * mm, 57 * mm]),
              p("5.3 运行控制", st["H2"]),
              p("Docker Compose 按 MySQL、SRS、backend、frontend 的依赖顺序启动；backend 初始化配置、数据库连接和路由后提供 /api/v1/health；frontend/Nginx 在后端健康后提供同源入口。请求先经过路由与身份/角色中间件，再进入 Handler、Logic、数据库或外部媒体组件。", st["Body"]),
              table([
                  ["控制点", "运行规则"],
                  ["身份与权限", "公开接口允许游客访问；业务写接口要求 JWT；后台操作继续校验 Staff/Admin。"],
                  ["事务与幂等", "订单支付、关系切换、预约和互动统计在事务/唯一约束下完成。"],
                  ["实时连接", "连接建立后登记用户和房间；断线清理；重连不重复统计在线用户。"],
                  ["定时任务", "预约启动和会员到期处理采用可重复执行的状态条件，避免重复迁移。"],
              ], [43 * mm, 131 * mm]),
              p("5.4 故障处理", st["H2"]),
              table([
                  ["故障", "处理方式"],
                  ["鉴权或权限失败", "返回 401/403；前端引导登录或显示无权限，不执行写操作。"],
                  ["参数/业务规则失败", "返回稳定错误码和可读原因；冲突、重复、终态覆盖等不写入数据。"],
                  ["数据库失败", "回滚事务；实时消息不广播伪成功；记录服务端日志。"],
                  ["媒体或推流失败", "显示处理中/等待/失败状态；不可用媒体不进入公开可播放状态。"],
                  ["WebSocket 中断", "客户端有限重连；已持久化消息可从历史恢复；在线计数最终清理。"],
              ], [43 * mm, 131 * mm]),
              p("5.5 运行时间", st["H2"]),
              table([
                  ["操作", "目标", "说明"],
                  ["登录与普通写操作", "正常网络下 1 秒级", "包含 JWT 校验和单次事务。"],
                  ["列表、搜索与分析", "正常数据量下 2 秒级", "分页、索引和限定日期范围。"],
                  ["页面首屏与视频详情", "3 秒内形成可操作页面", "媒体加载时间单独反馈。"],
                  ["直播弹幕和私信", "可感知实时", "WebSocket 传输，持久化成功后广播。"],
                  ["演示环境恢复", "容器重启后 2 分钟内恢复", "依赖健康检查和确定的启动顺序。"],
              ], [53 * mm, 48 * mm, 73 * mm]),
              p("5.6 当前设计风险", st["H2"]),
              table([
                  ["风险", "影响", "处理建议"],
                  ["总体组件图与代码不同步", "路由、Handler 或外部依赖变化后，图中职责和依赖可能过时。", "修改组件边界时同步更新 COMPONENT-OVERVIEW，并复核 13 张顺序图。"],
                  ["逻辑 Gateway 与运行实现命名不一致", "可能误认为独立网关已经部署。", "图注区分当前 Nginx + Gin Router 与未来 api-gateway 容器。"],
                  ["微服务版 E2E 尚未建立", "单体 E2E 全绿不能证明目标部署链路。", "Gateway 和服务容器就绪后增加独立 E2E 作业。"],
              ], [40 * mm, 60 * mm, 74 * mm]), PageBreak()]

    story += design_story
    story.append(PageBreak())
    story += [p("7 图件与追溯索引", st["H1"]), p("7.1 图件完整性", st["H2"])]
    story.append(table([
        ["图件类型", "现状", "结论"],
        ["组件级顺序图", "COMP-SEQ01 至 COMP-SEQ13 均有 .puml/.svg/.png", "13/13 齐全"],
        ["领域组件图", "COMPONENT-B、COMPONENT-D 均有 .puml/.svg/.png", "已覆盖 7 个用例域"],
        ["全系统总体组件图", "COMPONENT-OVERVIEW 有 .puml/.svg/.png，覆盖 UC01-UC13", "齐全"],
        ["部署图", "DEPLOY-MONO、DEPLOY-MONO-DETAILED、DEPLOY-K8S 均有源文件与导出物", "齐全"],
    ], [38 * mm, 102 * mm, 34 * mm]))
    story += [Spacer(1, 5 * mm), p("7.2 统一编号映射", st["H2"])]
    mapping = [["需求", "用例", "概要设计图", "详细设计图", "测试编号"]]
    for uc in UC:
        n = uc["no"]
        mapping.append([f"REQ{n}", f"UC{n}", f"COMP-SEQ{n}", f"OBJ-SEQ{n}", f"UNIT-TC{n} / INT-TC{n} / E2E-TC{n}"])
    story.append(table(mapping, [22 * mm, 22 * mm, 31 * mm, 31 * mm, 68 * mm], font_size=7.6))
    story += [Spacer(1, 5 * mm), p("说明：测试实际文件和执行结果以 docs/traceability/master.md 及 docs/testing/reports/ 为准；编号映射不等同于测试已经通过。", st["Warning"])]
    return story


def main() -> None:
    register_fonts()
    OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    st = styles()
    width, height = PAGE_SIZE
    frame = Frame(18 * mm, 16 * mm, width - 36 * mm, height - 31 * mm, id="content-frame")
    template = PageTemplate(id="content", frames=[frame], onPage=page_decor, pagesize=PAGE_SIZE)
    doc = DesignDocTemplate(
        str(OUTPUT), pagesize=PAGE_SIZE, leftMargin=18 * mm, rightMargin=18 * mm,
        topMargin=16 * mm, bottomMargin=16 * mm,
        title="DanmakuStream 软件概要设计说明书",
        author="DanmakuStream-Team",
        subject="UC01-UC13 概要设计",
    )
    doc.addPageTemplates([template])
    doc.multiBuild(build_story(st))
    print(OUTPUT)


if __name__ == "__main__":
    main()
