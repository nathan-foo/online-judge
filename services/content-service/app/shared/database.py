import os
from sqlalchemy.ext.asyncio import (
    AsyncSession,
    create_async_engine,
    async_sessionmaker
)
from sqlalchemy.orm import DeclarativeBase
from typing import AsyncGenerator

try:
    with open("/run/secrets/postgres_password") as f:
        password = f.read().strip()
except FileNotFoundError:
    password = os.environ["POSTGRES_PASSWORD"]

DATABASE_URL = f"postgresql+asyncpg://postgres:{password}@postgres:5432/content"

engine = create_async_engine(
    DATABASE_URL
)

async_session_maker = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False,
)

class Base(DeclarativeBase):
    pass

async def get_async_session() -> AsyncGenerator[AsyncSession, None]:
    async with async_session_maker() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
        finally:
            await session.close()
