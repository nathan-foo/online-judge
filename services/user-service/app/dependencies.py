from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession
from typing import Annotated
from .database import get_async_session

AsyncSessionDep = Annotated[AsyncSession, Depends(get_async_session)]
