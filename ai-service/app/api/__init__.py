from fastapi import APIRouter

from app.api.summary import router as summary_router
from app.api.tags import router as tags_router

router = APIRouter()
router.include_router(summary_router, prefix="/summary", tags=["视频摘要"])
router.include_router(tags_router, prefix="/tags", tags=["智能标签"])
