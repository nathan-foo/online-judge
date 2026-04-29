from contextlib import asynccontextmanager
from fastapi import FastAPI, status
from .database import engine, Base
from .dependencies import AsyncSessionDep
from .schemas import UserRead, UserCreate
from . import models, user_service  # noqa: F401 — registers User on Base.metadata


# TODO: Replace with Alembic migration later.
@asynccontextmanager
async def lifespan(app: FastAPI):
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    yield


app = FastAPI(lifespan=lifespan)


@app.get("/", response_model=list[UserRead])
async def get_users(session: AsyncSessionDep):
    return await user_service.get_users(session)


@app.post("/", response_model=UserRead, status_code=status.HTTP_201_CREATED)
async def create_user(user_in: UserCreate, session: AsyncSessionDep):
    return await user_service.create_user(session, user_in)


@app.get("/{clerk_user_id}", response_model=UserRead)
async def get_user(clerk_user_id: str, session: AsyncSessionDep):
    return await user_service.get_user(session, clerk_user_id)
