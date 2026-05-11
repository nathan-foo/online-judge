import uuid
from datetime import datetime, timezone
from fastapi import HTTPException, status
from sqlalchemy import select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import load_only, selectinload
from sqlalchemy.orm.exc import StaleDataError
from sqlalchemy.sql import func
from .models import Quiz, QuizProblem
from .schemas import QuizCreate, QuizProblemCreate, QuizProblemUpsert, QuizPublish, QuizRead, QuizUpdate


def _build_problem(p: QuizProblemCreate | QuizProblemUpsert) -> QuizProblem:
    return QuizProblem(
        type=p.payload.type,
        title=p.title,
        payload=p.payload.model_dump(mode="json"),
        position=p.position,
        points=p.points,
    )


def _apply_problem_diff(quiz: Quiz, desired: list[QuizProblemUpsert]) -> None:
    current_by_id = {p.id: p for p in quiz.problems}
    desired_ids = {d.id for d in desired if d.id is not None}
    unknown = desired_ids - current_by_id.keys()
    if unknown:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Unknown problem id(s) in request",
        )
    new_problems: list[QuizProblem] = []
    for d in desired:
        if d.id is None:
            new_problems.append(_build_problem(d))
            continue
        existing = current_by_id[d.id]
        payload_dict = d.payload.model_dump(mode="json")
        if existing.title != d.title:
            existing.title = d.title
        if existing.type != d.payload.type:
            existing.type = d.payload.type
        if existing.payload != payload_dict:
            existing.payload = payload_dict
        if existing.position != d.position:
            existing.position = d.position
        if existing.points != d.points:
            existing.points = d.points
        new_problems.append(existing)
    quiz.problems = new_problems
    quiz.problem_count = len(desired)


async def create_quiz(
    session: AsyncSession,
    owner_id: str,
    quiz_in: QuizCreate
) -> Quiz:
    quiz = Quiz(
        owner_id=owner_id,
        title=quiz_in.title,
        description=quiz_in.description,
        is_public=quiz_in.is_public,
        problem_count=len(quiz_in.problems),
        problems=[_build_problem(p) for p in quiz_in.problems],
    )
    session.add(quiz)
    try:
        await session.flush()
    except IntegrityError:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Duplicate position in quiz",
        )
    return await get_owned_quiz(session, owner_id, quiz.id)


_QUIZ_SUMMARY_LOAD = load_only(
    Quiz.id,
    Quiz.owner_id,
    Quiz.title,
    Quiz.description,
    Quiz.is_published,
    Quiz.is_public,
    Quiz.problem_count,
    Quiz.created_at,
    Quiz.updated_at,
    Quiz.published_at,
)


async def list_quizzes(
    session: AsyncSession,
    owner_id: str,
    limit: int,
    offset: int
) -> list[Quiz]:
    result = await session.execute(
        select(Quiz)
        .where(Quiz.owner_id == owner_id, Quiz.is_deleted == False)
        .options(_QUIZ_SUMMARY_LOAD)
        .order_by(Quiz.created_at.desc())
        .limit(limit)
        .offset(offset)
    )
    return list(result.scalars().all())


async def list_public_quizzes(
    session: AsyncSession,
    limit: int,
    offset: int
) -> list[Quiz]:
    result = await session.execute(
        select(Quiz)
        .where(Quiz.is_public == True, Quiz.is_published == True, Quiz.is_deleted == False)
        .options(_QUIZ_SUMMARY_LOAD)
        .order_by(Quiz.created_at.desc())
        .limit(limit)
        .offset(offset)
    )
    return list(result.scalars().all())


async def get_quiz(
    session: AsyncSession,
    viewer_id: str,
    quiz_id: uuid.UUID
) -> Quiz | QuizRead:
    result = await session.execute(
        select(Quiz)
        .where(Quiz.id == quiz_id, Quiz.is_deleted == False)
        .options(selectinload(Quiz.problems))
    )
    quiz = result.scalar_one_or_none()
    if not quiz or (quiz.owner_id != viewer_id and not (quiz.is_public and quiz.is_published)):
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Quiz not found",
        )
    if quiz.owner_id != viewer_id:
        return QuizRead.model_validate(quiz.published_snapshot)
    return quiz


async def get_owned_quiz(
    session: AsyncSession,
    owner_id: str,
    quiz_id: uuid.UUID
) -> Quiz:
    result = await session.execute(
        select(Quiz)
        .where(Quiz.id == quiz_id, Quiz.owner_id == owner_id, Quiz.is_deleted == False)
        .options(selectinload(Quiz.problems))
    )
    quiz = result.scalar_one_or_none()
    if not quiz:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Quiz not found",
        )
    return quiz


async def update_quiz(
    session: AsyncSession,
    quiz: Quiz,
    quiz_in: QuizUpdate
) -> Quiz:
    if quiz_in.version != quiz.version:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="Quiz was modified by another request",
        )
    data = quiz_in.model_dump(exclude_unset=True, exclude={"problems", "version"})
    for field, value in data.items():
        setattr(quiz, field, value)
    if quiz_in.problems is not None:
        _apply_problem_diff(quiz, quiz_in.problems)
        quiz.updated_at = func.now()
    try:
        await session.flush()
    except StaleDataError:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="Quiz was modified by another request",
        )
    except IntegrityError:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Duplicate position in quiz",
        )
    return await get_owned_quiz(session, quiz.owner_id, quiz.id)


async def delete_quiz(
    session: AsyncSession,
    quiz: Quiz
) -> None:
    quiz.is_deleted = True
    try:
        await session.flush()
    except StaleDataError:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="Quiz was modified by another request",
        )


async def publish_quiz(
    session: AsyncSession,
    quiz: Quiz,
    publish_in: QuizPublish,
) -> Quiz:
    if publish_in.version != quiz.version:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="Quiz was modified by another request",
        )
    quiz.is_published = True
    quiz.published_at = datetime.now(timezone.utc)
    quiz.published_snapshot = QuizRead.model_validate(quiz).model_dump(mode="json")
    try:
        await session.flush()
    except StaleDataError:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="Quiz was modified by another request",
        )
    return await get_owned_quiz(session, quiz.owner_id, quiz.id)
