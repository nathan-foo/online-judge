from fastapi import HTTPException, status
from sqlalchemy import select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.ext.asyncio import AsyncSession
from .models import User
from .schemas import UserCreate, UserUpdate


async def sync_user(session: AsyncSession, user_in: UserCreate) -> None:
    result = await session.execute(select(User).where(User.clerk_user_id == user_in.clerk_user_id))
    user = result.scalar_one_or_none()
    if user:
        for field, value in user_in.model_dump().items():
            setattr(user, field, value)
    else:
        session.add(User(**user_in.model_dump()))
    await session.flush()


async def get_user(session: AsyncSession, clerk_user_id: str) -> User:
    result = await session.execute(select(User).where(User.clerk_user_id == clerk_user_id))
    user = result.scalar_one_or_none()
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="User not found",
        )
    return user


async def deactivate_user(session: AsyncSession, clerk_user_id: str) -> None:
    result = await session.execute(select(User).where(User.clerk_user_id == clerk_user_id))
    user = result.scalar_one_or_none()
    if user:
        user.is_active = False
        await session.flush()


async def update_user(session: AsyncSession, user: User, user_in: UserUpdate) -> User:
    for field, value in user_in.model_dump(exclude_unset=True).items():
        setattr(user, field, value)
    try:
        await session.flush()
    except IntegrityError:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="User with this username already exists",
        )
    await session.refresh(user)
    return user
