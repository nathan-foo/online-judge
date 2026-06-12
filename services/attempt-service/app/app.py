import uuid
from contextlib import asynccontextmanager
from fastapi import BackgroundTasks, FastAPI, Query, status
from .database import engine, Base, async_session_maker
from .dependencies import AsyncSessionDep, CurrentUserIdDep
from .messaging import broker
from .schemas import AnswerPayload, AttemptAnswerRead, AttemptCreate, AttemptRead, AttemptSummary
from . import attempt_service, models, quiz_client  # noqa: F401 — registers Attempt on Base.metadata


async def _handle_eval_result(payload: dict) -> None:
    async with async_session_maker() as session:
        await attempt_service.apply_eval_result(session, payload)
        await session.commit()


@asynccontextmanager
async def lifespan(app: FastAPI):
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    await broker.connect()
    await broker.consume_results(_handle_eval_result)
    yield
    await broker.close()
    await quiz_client.close()


app = FastAPI(lifespan=lifespan)


@app.post("/", response_model=AttemptRead, status_code=status.HTTP_201_CREATED)
async def start_attempt(
    attempt_in: AttemptCreate,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    attempt = await attempt_service.start_attempt(session, user_id, attempt_in.quiz_id)
    return attempt_service.to_attempt_read(attempt)


@app.get("/", response_model=list[AttemptSummary])
async def list_attempts(
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep,
    limit: int = Query(default=20, ge=1, le=100),
    offset: int = Query(default=0, ge=0),
):
    return await attempt_service.list_attempts(session, user_id, limit, offset)


@app.get("/{attempt_id}", response_model=AttemptRead)
async def get_attempt(
    attempt_id: uuid.UUID,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    attempt = await attempt_service.get_owned_attempt(session, user_id, attempt_id)
    return attempt_service.to_attempt_read(attempt)


@app.put("/{attempt_id}/answers/{problem_id}", response_model=AttemptAnswerRead)
async def save_answer(
    attempt_id: uuid.UUID,
    problem_id: uuid.UUID,
    answer_in: AnswerPayload,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    attempt = await attempt_service.get_owned_attempt(session, user_id, attempt_id)
    return await attempt_service.save_answer(session, attempt, problem_id, answer_in)


@app.post("/{attempt_id}/submit", response_model=AttemptRead)
async def submit_attempt(
    attempt_id: uuid.UUID,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep,
    background_tasks: BackgroundTasks
):
    attempt = await attempt_service.get_owned_attempt(session, user_id, attempt_id)
    attempt, eval_requests = await attempt_service.submit_attempt(session, attempt)
    
    for request in eval_requests:
        background_tasks.add_task(broker.publish_eval_request, request)
    return attempt_service.to_attempt_read(attempt)
