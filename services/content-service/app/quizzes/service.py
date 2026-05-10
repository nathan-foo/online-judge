import uuid
from datetime import datetime, timezone
from fastapi import HTTPException, status
from sqlalchemy import select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import load_only, selectinload
from .models import Quiz, QuizProblem
from ..problems.models import Problem
from .schemas import QuizCreate, QuizRead, QuizUpdate


async def create_quiz(
    session: AsyncSession,
    owner_id: str,
    quiz_in: QuizCreate
) -> Quiz:
    if quiz_in.problems:
        problem_ids = [p.problem_id for p in quiz_in.problems]
        result = await session.execute(
            select(Problem.id).where(
                Problem.id.in_(problem_ids),
                Problem.owner_id == owner_id,
                Problem.is_deleted == False,
            )
        )
        owned = set(result.scalars().all())
        if set(problem_ids) - owned:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="One or more problems not found or not owned by you",
            )
    quiz = Quiz(
        owner_id=owner_id,
        title=quiz_in.title,
        description=quiz_in.description,
        is_public=quiz_in.is_public,
        problem_count=len(quiz_in.problems),
        problems=[
            QuizProblem(problem_id=p.problem_id, position=p.position, points=p.points)
            for p in quiz_in.problems
        ],
    )
    session.add(quiz)
    try:
        await session.flush()
    except IntegrityError:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Duplicate problem or position in quiz",
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
    )
    quiz = result.scalar_one_or_none()
    if not quiz or (quiz.owner_id != viewer_id and not (quiz.is_public and quiz.is_published)):
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Quiz not found",
        )
    if quiz.owner_id != viewer_id:
        return QuizRead.model_validate(quiz.published_snapshot)
    return await get_owned_quiz(session, quiz.owner_id, quiz_id)


async def get_owned_quiz(
    session: AsyncSession,
    owner_id: str,
    quiz_id: uuid.UUID
) -> Quiz:
    result = await session.execute(
        select(Quiz)
        .where(Quiz.id == quiz_id, Quiz.owner_id == owner_id, Quiz.is_deleted == False)
        .options(selectinload(Quiz.problems).selectinload(QuizProblem.problem))
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
    data = quiz_in.model_dump(exclude_unset=True, exclude={"problems"})
    for field, value in data.items():
        setattr(quiz, field, value)
    if quiz_in.problems is not None:
        new_ids = [p.problem_id for p in quiz_in.problems]
        if new_ids:
            result = await session.execute(
                select(Problem.id).where(
                    Problem.id.in_(new_ids),
                    Problem.owner_id == quiz.owner_id,
                    Problem.is_deleted == False,
                )
            )
            owned = set(result.scalars().all())
            if set(new_ids) - owned:
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="One or more problems not found or not owned by you",
                )
        quiz.problems = [
            QuizProblem(problem_id=p.problem_id, position=p.position, points=p.points)
            for p in quiz_in.problems
        ]
        quiz.problem_count = len(quiz_in.problems)
    try:
        await session.flush()
    except IntegrityError:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Duplicate problem or position in quiz",
        )
    return await get_owned_quiz(session, quiz.owner_id, quiz.id)


async def delete_quiz(
    session: AsyncSession,
    quiz: Quiz
) -> None:
    quiz.is_deleted = True
    await session.flush()


async def publish_quiz(
    session: AsyncSession,
    quiz: Quiz,
) -> Quiz:
    quiz.is_published = True
    quiz.published_at = datetime.now(timezone.utc)
    quiz.published_snapshot = QuizRead.model_validate(quiz).model_dump(mode="json")
    await session.flush()
    return await get_owned_quiz(session, quiz.owner_id, quiz.id)
