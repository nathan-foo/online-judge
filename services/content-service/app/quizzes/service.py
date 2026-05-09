from fastapi import HTTPException, status
from sqlalchemy import select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.ext.asyncio import AsyncSession
from .models import Quiz, QuizProblem
from ..problems.models import Problem
from .schemas import QuizCreate, QuizUpdate
