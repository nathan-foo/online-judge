from fastapi import Depends, Header, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession
from typing import Annotated
from .database import get_async_session
from .models import User
from . import user_service

AsyncSessionDep = Annotated[AsyncSession, Depends(get_async_session)]


async def get_current_user(
    session: AsyncSessionDep,
    x_user_id: Annotated[str | None, Header()] = None,
) -> User:
    if not x_user_id:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Missing user identity",
        )
    return await user_service.get_user(session, x_user_id)


CurrentUserDep = Annotated[User, Depends(get_current_user)]
