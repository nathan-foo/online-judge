import uuid
from typing import Optional
from fastapi import HTTPException, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from .models import Problem, ProblemType
from .schemas import ProblemCreate, ProblemUpdate


async def create_problem(
    session: AsyncSession,
    owner_id: str,
    problem_in: ProblemCreate
) -> Problem:
    problem = Problem(
        owner_id=owner_id,
        type=problem_in.payload.type,
        title=problem_in.title,
        payload=problem_in.payload.model_dump(mode="json"),
    )
    session.add(problem)
    await session.flush()
    await session.refresh(problem)
    return problem


async def list_problems(
    session: AsyncSession,
    owner_id: str,
    type: Optional[ProblemType] = None,
    limit: int = 20,
    offset: int = 0,
) -> list[Problem]:
    stmt = select(Problem).where(Problem.owner_id == owner_id, Problem.is_deleted == False)
    if type is not None:
        stmt = stmt.where(Problem.type == type)
    stmt = stmt.order_by(Problem.created_at.desc()).limit(limit).offset(offset)
    result = await session.execute(stmt)
    return list(result.scalars().all())


async def get_problem(
    session: AsyncSession,
    owner_id: str,
    problem_id: uuid.UUID
) -> Problem:
    result = await session.execute(
        select(Problem).where(
            Problem.id == problem_id,
            Problem.owner_id == owner_id,
            Problem.is_deleted == False,
        )
    )
    problem = result.scalar_one_or_none()
    if not problem:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Problem not found",
        )
    return problem


async def update_problem(
    session: AsyncSession,
    problem: Problem,
    problem_in: ProblemUpdate
) -> Problem:
    if problem_in.title is not None:
        problem.title = problem_in.title
    if problem_in.payload is not None:
        if problem_in.payload.type != problem.type:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Cannot change problem type",
            )
        problem.payload = problem_in.payload.model_dump(mode="json")
    await session.flush()
    await session.refresh(problem)
    return problem


async def delete_problem(
    session: AsyncSession,
    problem: Problem
) -> None:
    problem.is_deleted = True
    await session.flush()
