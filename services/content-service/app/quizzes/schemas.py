import uuid
from datetime import datetime
from typing import Optional
from pydantic import BaseModel, ConfigDict, Field, model_validator
from ..problems.schemas import ProblemRead


class QuizProblemRef(BaseModel):
    problem_id: uuid.UUID
    position: int = Field(ge=1)
    points: int = Field(default=1, ge=1, le=100)


class QuizProblemRead(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    position: int
    points: int
    problem: ProblemRead


class QuizCreate(BaseModel):
    title: str = Field(min_length=1, max_length=255)
    description: Optional[str] = Field(default=None, max_length=2000)
    is_public: bool = False
    problems: list[QuizProblemRef] = []

    @model_validator(mode="after")
    def _validate(self):
        positions = [p.position for p in self.problems]
        if len(positions) != len(set(positions)):
            raise ValueError("positions must be unique")
        return self


class QuizUpdate(BaseModel):
    title: Optional[str] = Field(default=None, min_length=1, max_length=255)
    description: Optional[str] = Field(default=None, max_length=2000)
    is_public: Optional[bool] = None
    problems: Optional[list[QuizProblemRef]] = None

    @model_validator(mode="after")
    def _validate(self):
        if self.problems is None:
            return self
        positions = [p.position for p in self.problems]
        if len(positions) != len(set(positions)):
            raise ValueError("positions must be unique")
        return self


class QuizRead(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    owner_id: str
    title: str
    description: Optional[str] = None
    is_published: bool
    is_public: bool
    problems: list[QuizProblemRead]
    created_at: datetime
    updated_at: datetime


class QuizSummary(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    owner_id: str
    title: str
    description: Optional[str] = None
    is_published: bool
    is_public: bool
    problem_count: int
    created_at: datetime
    updated_at: datetime


class QuizPublicSummary(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    title: str
    description: Optional[str] = None
    is_published: bool
    is_public: bool
    problem_count: int
    created_at: datetime
    updated_at: datetime