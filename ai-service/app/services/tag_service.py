"""
智能标签提取 Service
"""
import json

from langchain_openai import ChatOpenAI
from langchain.prompts import ChatPromptTemplate
from langchain.output_parsers import CommaSeparatedListOutputParser

from app.core.config import settings


async def extract_tags(title: str, description: str) -> list[str]:
    llm = ChatOpenAI(
        model=settings.llm_model,
        api_key=settings.openai_api_key,
        base_url=settings.openai_base_url,
        temperature=0.2,
    )
    parser = CommaSeparatedListOutputParser()

    prompt = ChatPromptTemplate.from_messages([
        ("system", "你是一位视频内容分类专家。根据视频标题和简介，提取3-6个最相关的内容标签，标签应简洁（2-5个汉字），用英文逗号分隔。"),
        ("human", "标题：{title}\n简介：{description}\n\n标签："),
    ])

    chain = prompt | llm | parser
    tags = await chain.ainvoke({"title": title, "description": description})
    return [t.strip() for t in tags if t.strip()][:6]
