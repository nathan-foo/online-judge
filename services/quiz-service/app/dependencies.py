from fastapi import Depends, Header, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession
from typing import Annotated, Optional
from .database import get_async_session

AsyncSessionDep = Annotated[AsyncSession, Depends(get_async_session)]


async def get_current_user_id(
    x_user_id: Annotated[Optional[str], Header()] = None,
) -> str:
    if not x_user_id:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Missing user identity",
        )
    return x_user_id


CurrentUserIdDep = Annotated[str, Depends(get_current_user_id)]
