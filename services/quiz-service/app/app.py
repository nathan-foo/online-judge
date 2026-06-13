import uuid
from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException, Query, status
from sqlalchemy import text
from .database import engine, Base
from .dependencies import AsyncSessionDep, CurrentUserIdDep
from .schemas import QuizCreate, QuizPublicRead, QuizPublish, QuizRead, QuizSummary, QuizUpdate
from . import models, quiz_service  # noqa: F401 — registers Quiz on Base.metadata


@asynccontextmanager
async def lifespan(app: FastAPI):
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    yield


app = FastAPI(lifespan=lifespan)


@app.get("/healthz", include_in_schema=False)
async def healthz():
    return {"status": "ok"}


@app.get("/readyz", include_in_schema=False)
async def readyz():
    try:
        async with engine.connect() as conn:
            await conn.execute(text("SELECT 1"))
    except Exception:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE, detail="database unavailable"
        )
    return {"status": "ready"}


@app.post("/", response_model=QuizRead, status_code=status.HTTP_201_CREATED)
async def create_quiz(
    quiz_in: QuizCreate,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    return await quiz_service.create_quiz(session, user_id, quiz_in)


@app.get("/", response_model=list[QuizSummary])
async def list_quizzes(
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep,
    limit: int = Query(default=20, ge=1, le=100),
    offset: int = Query(default=0, ge=0),
):
    return await quiz_service.list_quizzes(session, user_id, limit, offset)


@app.get("/public", response_model=list[QuizSummary])
async def list_public_quizzes(
    session: AsyncSessionDep,
    limit: int = Query(default=20, ge=1, le=100),
    offset: int = Query(default=0, ge=0),
):
    return await quiz_service.list_public_quizzes(session, limit, offset)


@app.get("/internal/{quiz_id}/snapshot", response_model=QuizRead)
async def get_quiz_snapshot(
    quiz_id: uuid.UUID,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    return await quiz_service.get_quiz_snapshot(session, user_id, quiz_id)


@app.get("/{quiz_id}", response_model=QuizRead | QuizPublicRead)
async def get_quiz(
    quiz_id: uuid.UUID,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    return await quiz_service.get_quiz(session, user_id, quiz_id)


@app.patch("/{quiz_id}", response_model=QuizRead)
async def update_quiz(
    quiz_id: uuid.UUID,
    quiz_in: QuizUpdate,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    quiz = await quiz_service.get_owned_quiz(session, user_id, quiz_id)
    return await quiz_service.update_quiz(session, quiz, quiz_in)


@app.delete("/{quiz_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_quiz(
    quiz_id: uuid.UUID,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    quiz = await quiz_service.get_owned_quiz(session, user_id, quiz_id)
    await quiz_service.delete_quiz(session, quiz)


@app.post("/{quiz_id}/publish", response_model=QuizRead)
async def publish_quiz(
    quiz_id: uuid.UUID,
    publish_in: QuizPublish,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    quiz = await quiz_service.get_owned_quiz(session, user_id, quiz_id)
    return await quiz_service.publish_quiz(session, quiz, publish_in)
