"""
智能标签提取 API
- POST /api/tags/extract  从标题+描述提取标签
"""
from fastapi import APIRouter
from pydantic import BaseModel

from app.services.tag_service import extract_tags

router = APIRouter()


class TagRequest(BaseModel):
    title: str
    description: str


class TagResponse(BaseModel):
    tags: list[str]


@router.post("", response_model=TagResponse)
async def extract_video_tags(req: TagRequest):
    tags = await extract_tags(req.title, req.description)
    return TagResponse(tags=tags)
