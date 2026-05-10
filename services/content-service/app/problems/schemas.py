import uuid
import enum
from datetime import datetime
from typing import Annotated, Literal, Optional, Union
from pydantic import BaseModel, ConfigDict, Field, HttpUrl, model_validator
from .models import ProblemType


class MCChoice(BaseModel):
    id: str = Field(min_length=1, max_length=64)
    text: str = Field(min_length=1, max_length=200)


class MultipleChoicePayload(BaseModel):
    type: Literal[ProblemType.MULTIPLE_CHOICE]
    prompt: str = Field(min_length=1, max_length=1000)
    choices: list[MCChoice] = Field(min_length=2, max_length=10)
    correct_choice_ids: list[str] = Field(min_length=1)
    multiple_correct: bool = False
    shuffle_choices: bool = False
    image_url: Optional[HttpUrl] = Field(default=None, max_length=2000)
    explanation: Optional[str] = Field(default=None, max_length=2000)
    
    @model_validator(mode="after")
    def _validate(self):
        ids = [c.id for c in self.choices]
        if len(ids) != len(set(ids)):
            raise ValueError("choice ids must be unique")
        unknown = set(self.correct_choice_ids) - set(ids)
        if unknown:
            raise ValueError(f"correct_choice_ids reference unknown choices: {unknown}")
        if not self.multiple_correct and len(self.correct_choice_ids) != 1:
            raise ValueError("single-answer questions must have exactly one correct choice")
        return self


class Language(str, enum.Enum):
    PYTHON = "python"
    JAVA = "java"
    CPP = "cpp"
    JAVASCRIPT = "javascript"


class TestCase(BaseModel):
    id: str = Field(min_length=1, max_length=64)
    stdin: str = Field(default="", max_length=10_000)
    expected_stdout: str = Field(max_length=10_000)
    is_example: bool = False


class CodePayload(BaseModel):
    type: Literal[ProblemType.CODE]
    prompt: str = Field(min_length=1, max_length=2000)
    languages: list[Language] = Field(min_length=1)
    starter_code: dict[Language, str] = Field(default_factory=dict)
    test_cases: list[TestCase] = Field(min_length=1, max_length=50)
    time_limit_ms: int = Field(default=2000, ge=1000, le=10_000)
    memory_limit_mb: int = Field(default=256, ge=16, le=1024)
    reference_solutions: dict[Language, str] = Field(default_factory=dict)
    
    @model_validator(mode="after")
    def _validate(self):
        ids = [t.id for t in self.test_cases]
        if len(ids) != len(set(ids)):
            raise ValueError("test_case ids must be unique")
        for field_name in ("starter_code", "reference_solutions"):
            extra = set(getattr(self, field_name).keys()) - set(self.languages)
            if extra:
                raise ValueError(f"{field_name} references languages not in `languages`: {extra}")
        if not any(t.is_example for t in self.test_cases):
            raise ValueError("at least one test case must an example")
        return self


ProblemPayload = Annotated[
    Union[MultipleChoicePayload, CodePayload],
    Field(discriminator="type"),
]


class ProblemCreate(BaseModel):
    title: str = Field(min_length=1, max_length=255)
    payload: ProblemPayload


class ProblemUpdate(BaseModel):
    title: Optional[str] = Field(default=None, min_length=1, max_length=255)
    payload: Optional[ProblemPayload] = None


class ProblemRead(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    owner_id: str
    type: ProblemType
    title: str
    payload: ProblemPayload
    created_at: datetime
    updated_at: datetime