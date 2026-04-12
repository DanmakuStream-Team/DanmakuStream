from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    app_name: str = "DanmakuStream AI Service"
    debug: bool = False
    host: str = "0.0.0.0"
    port: int = 8000

    # LLM
    openai_api_key: str = ""
    openai_base_url: str = "https://api.deepseek.com/v1"  # DeepSeek compatible
    llm_model: str = "deepseek-chat"

    # Redis
    redis_url: str = "redis://redis:6379/1"

    # MySQL (for reading video metadata)
    database_url: str = "mysql+pymysql://root:password@mysql:3306/danmakustream"

    # Internal auth (shared secret with Go backend)
    internal_secret: str = "internal-service-secret"

    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8")


settings = Settings()
