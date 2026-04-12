"""
视频摘要 API
- POST /api/summary/video  异步生成视频摘要（课代表功能）
- GET  /api/summary/video/{video_id}  查询摘要结果
"""
import json
import uuid

from fastapi import APIRouter, BackgroundTasks, HTTPException, Request
from pydantic import BaseModel

from app.services.summary_service import generate_video_summary

router = APIRouter()


class SummaryRequest(BaseModel):
    video_id: int
    title: str
    description: str
    duration: int  # seconds


class SummaryResponse(BaseModel):
    task_id: str
    status: str  # pending | processing | done | failed
    summary: str | None = None


@router.post("", response_model=SummaryResponse)
async def request_summary(req: SummaryRequest, background_tasks: BackgroundTasks, request: Request):
    task_id = str(uuid.uuid4())
    rdb = request.app.state.redis

    # Store initial task state
    await rdb.hset(f"ai:summary:{req.video_id}", mapping={
        "task_id": task_id,
        "status": "pending",
        "summary": "",
    })

    # Run in background — never blocks the response
    background_tasks.add_task(
        generate_video_summary,
        rdb=rdb,
        video_id=req.video_id,
        title=req.title,
        description=req.description,
    )

    return SummaryResponse(task_id=task_id, status="pending")


@router.get("/{video_id}", response_model=SummaryResponse)
async def get_summary(video_id: int, request: Request):
    rdb = request.app.state.redis
    data = await rdb.hgetall(f"ai:summary:{video_id}")
    if not data:
        raise HTTPException(status_code=404, detail="摘要任务不存在")
    return SummaryResponse(
        task_id=data.get("task_id", ""),
        status=data.get("status", "pending"),
        summary=data.get("summary") or None,
    )
