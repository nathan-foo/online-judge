import uuid
import enum
from datetime import datetime
from typing import Annotated, Literal, Optional, Union
from pydantic import BaseModel, ConfigDict, Field
from .models import AttemptStatus, EvalStatus, ProblemType


class Language(str, enum.Enum):
    PYTHON = "python"
    C = "c"
    CPP = "cpp"
    JAVA = "java"
    JAVASCRIPT = "javascript"
    GO = "go"
    TYPESCRIPT = "typescript"
    KOTLIN = "kotlin"
    RUST = "rust"
    CSHARP = "csharp"


class MultipleChoiceAnswer(BaseModel):
    type: Literal[ProblemType.MULTIPLE_CHOICE]
    selected_choice_ids: list[str] = Field(min_length=1, max_length=10)


class CodeAnswer(BaseModel):
    type: Literal[ProblemType.CODE]
    language: Language
    source_code: str = Field(min_length=1, max_length=100_000)


AnswerPayload = Annotated[
    Union[MultipleChoiceAnswer, CodeAnswer],
    Field(discriminator="type"),
]


class AttemptCreate(BaseModel):
    quiz_id: uuid.UUID


class AttemptAnswerRead(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    problem_id: uuid.UUID
    problem_type: ProblemType
    answer: dict
    is_correct: Optional[bool] = None
    points_awarded: Optional[int] = None
    eval_status: Optional[EvalStatus] = None
    eval_result: Optional[dict] = None
    updated_at: datetime


class AttemptRead(BaseModel):
    id: uuid.UUID
    quiz_id: uuid.UUID
    quiz_version: int
    quiz_title: str
    quiz: dict
    status: AttemptStatus
    score: Optional[int] = None
    max_score: int
    started_at: datetime
    submitted_at: Optional[datetime] = None
    answers: list[AttemptAnswerRead]


class AttemptSummary(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    quiz_id: uuid.UUID
    quiz_version: int
    quiz_title: str
    status: AttemptStatus
    score: Optional[int] = None
    max_score: int
    started_at: datetime
    submitted_at: Optional[datetime] = None
