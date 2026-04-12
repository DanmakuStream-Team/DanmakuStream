"""
视频摘要生成 Service（异步后台任务）
"""
from langchain_openai import ChatOpenAI
from langchain.prompts import ChatPromptTemplate

from app.core.config import settings


async def generate_video_summary(rdb, video_id: int, title: str, description: str):
    await rdb.hset(f"ai:summary:{video_id}", "status", "processing")

    try:
        llm = ChatOpenAI(
            model=settings.llm_model,
            api_key=settings.openai_api_key,
            base_url=settings.openai_base_url,
            temperature=0.3,
        )

        prompt = ChatPromptTemplate.from_messages([
            ("system", "你是一位专业的视频内容摘要助手（视频课代表）。请根据视频标题和简介，生成简洁的三点式摘要，帮助观众快速了解视频核心内容。"),
            ("human", "视频标题：{title}\n视频简介：{description}\n\n请生成摘要（三条要点，每条不超过50字）："),
        ])

        chain = prompt | llm
        result = await chain.ainvoke({"title": title, "description": description})
        summary = result.content

        await rdb.hset(f"ai:summary:{video_id}", mapping={
            "status": "done",
            "summary": summary,
        })
    except Exception as e:
        await rdb.hset(f"ai:summary:{video_id}", mapping={
            "status": "failed",
            "summary": f"生成失败：{str(e)}",
        })
