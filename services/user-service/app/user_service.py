from fastapi import HTTPException, status
from sqlalchemy import select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.ext.asyncio import AsyncSession
from .models import User
from .schemas import UserCreate


async def get_users(session: AsyncSession) -> list[User]:
    result = await session.execute(select(User))
    return list(result.scalars().all())


async def create_user(session: AsyncSession, user_in: UserCreate) -> User:
    user = User(**user_in.model_dump())
    session.add(user)
    try:
        await session.flush()
    except IntegrityError:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="User with this clerk_user_id or email already exists",
        )
    await session.refresh(user)
    return user


async def get_user(session: AsyncSession, clerk_user_id: str) -> User:
    result = await session.execute(select(User).where(User.clerk_user_id == clerk_user_id))
    user = result.scalar_one_or_none()
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"User not found",
        )
    return user
