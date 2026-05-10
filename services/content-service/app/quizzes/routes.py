import uuid
from fastapi import APIRouter, Depends, Query, status
from ..shared.dependencies import get_current_user_id, CurrentUserIdDep, AsyncSessionDep
from .schemas import QuizCreate, QuizRead, QuizSummary, QuizUpdate
from . import service

router = APIRouter(
    prefix="/quizzes",
    dependencies=[Depends(get_current_user_id)],
)


@router.post("/", response_model=QuizRead, status_code=status.HTTP_201_CREATED)
async def create_quiz(
    quiz_in: QuizCreate,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    return await service.create_quiz(session, user_id, quiz_in)


@router.get("/", response_model=list[QuizSummary])
async def list_quizzes(
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep,
    limit: int = Query(default=20, ge=1, le=100),
    offset: int = Query(default=0, ge=0),
):
    return await service.list_quizzes(session, user_id, limit, offset)


@router.get("/public", response_model=list[QuizSummary])
async def list_public_quizzes(
    session: AsyncSessionDep,
    limit: int = Query(default=20, ge=1, le=100),
    offset: int = Query(default=0, ge=0),
):
    return await service.list_public_quizzes(session, limit, offset)


@router.get("/{quiz_id}", response_model=QuizRead)
async def get_quiz(
    quiz_id: uuid.UUID,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    return await service.get_quiz(session, user_id, quiz_id)


@router.patch("/{quiz_id}", response_model=QuizRead)
async def update_quiz(
    quiz_id: uuid.UUID,
    quiz_in: QuizUpdate,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    quiz = await service.get_owned_quiz(session, user_id, quiz_id)
    return await service.update_quiz(session, quiz, quiz_in)


@router.delete("/{quiz_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_quiz(
    quiz_id: uuid.UUID,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    quiz = await service.get_owned_quiz(session, user_id, quiz_id)
    await service.delete_quiz(session, quiz)


@router.post("/{quiz_id}/publish", response_model=QuizRead)
async def publish_quiz(
    quiz_id: uuid.UUID,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    quiz = await service.get_owned_quiz(session, user_id, quiz_id)
    return await service.publish_quiz(session, quiz)
