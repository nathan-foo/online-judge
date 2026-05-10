import uuid
from typing import Optional
from fastapi import APIRouter, Depends, Query, status
from ..shared.dependencies import get_current_user_id, CurrentUserIdDep, AsyncSessionDep
from .models import ProblemType
from .schemas import ProblemCreate, ProblemUpdate, ProblemRead, ProblemSummary
from . import service

router = APIRouter(
    prefix="/problems",
    dependencies=[Depends(get_current_user_id)],
)


@router.post("/", response_model=ProblemRead, status_code=status.HTTP_201_CREATED)
async def create_problem(
    problem_in: ProblemCreate,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    return await service.create_problem(session, user_id, problem_in)


@router.get("/", response_model=list[ProblemSummary])
async def list_problems(
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep,
    type: Optional[ProblemType] = None,
    limit: int = Query(default=20, ge=1, le=100),
    offset: int = Query(default=0, ge=0),
):
    return await service.list_problems(session, user_id, type, limit, offset)


@router.get("/{problem_id}", response_model=ProblemRead)
async def get_problem(
    problem_id: uuid.UUID,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    return await service.get_problem(session, user_id, problem_id)


@router.patch("/{problem_id}", response_model=ProblemRead)
async def update_problem(
    problem_id: uuid.UUID,
    problem_in: ProblemUpdate,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    problem = await service.get_problem(session, user_id, problem_id)
    return await service.update_problem(session, problem, problem_in)


@router.delete("/{problem_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_problem(
    problem_id: uuid.UUID,
    user_id: CurrentUserIdDep,
    session: AsyncSessionDep
):
    problem = await service.get_problem(session, user_id, problem_id)
    await service.delete_problem(session, problem)
