import uuid
import enum
from sqlalchemy import String, DateTime, Uuid, Integer, ForeignKey, UniqueConstraint, Index, Boolean, Enum, text
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.sql import func
from typing import Optional
from datetime import datetime
from .database import Base


class ProblemType(str, enum.Enum):
    MULTIPLE_CHOICE = "multiple_choice"
    CODE = "code"


class AttemptStatus(str, enum.Enum):
    IN_PROGRESS = "in_progress"
    GRADING = "grading"
    GRADED = "graded"


class EvalStatus(str, enum.Enum):
    PENDING = "pending"
    DONE = "done"
    ERROR = "error"


class Attempt(Base):
    __tablename__ = "attempts"
    __table_args__ = (
        Index(
            "uq_attempts_user_quiz_active",
            "user_id",
            "quiz_id",
            unique=True,
            postgresql_where=text("status = 'IN_PROGRESS'"),
        ),
        Index(
            "ix_attempts_user",
            "user_id",
            text("started_at DESC"),
        ),
    )

    id: Mapped[uuid.UUID] = mapped_column(Uuid, primary_key=True, default=uuid.uuid4)
    user_id: Mapped[str] = mapped_column(String(255), nullable=False)

    quiz_id: Mapped[uuid.UUID] = mapped_column(Uuid, nullable=False)
    quiz_version: Mapped[int] = mapped_column(Integer, nullable=False)
    quiz_title: Mapped[str] = mapped_column(String(255), nullable=False)
    quiz_snapshot: Mapped[dict] = mapped_column(JSONB, nullable=False)

    status: Mapped[AttemptStatus] = mapped_column(
        Enum(AttemptStatus), default=AttemptStatus.IN_PROGRESS, nullable=False
    )
    score: Mapped[Optional[int]] = mapped_column(Integer, nullable=True)
    max_score: Mapped[int] = mapped_column(Integer, nullable=False)

    started_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now()
    )
    submitted_at: Mapped[Optional[datetime]] = mapped_column(
        DateTime(timezone=True), nullable=True
    )

    answers: Mapped[list["AttemptAnswer"]] = relationship(
        back_populates="attempt",
        cascade="all, delete-orphan",
    )

    def __repr__(self) -> str:
        return f"<Attempt(id={self.id}, quiz_id={self.quiz_id}, status={self.status})>"


class AttemptAnswer(Base):
    __tablename__ = "attempt_answers"
    __table_args__ = (
        UniqueConstraint("attempt_id", "problem_id", name="uq_attempt_problem"),
    )

    id: Mapped[uuid.UUID] = mapped_column(Uuid, primary_key=True, default=uuid.uuid4)
    attempt_id: Mapped[uuid.UUID] = mapped_column(
        Uuid, ForeignKey("attempts.id", ondelete="CASCADE"), nullable=False
    )
    problem_id: Mapped[uuid.UUID] = mapped_column(Uuid, nullable=False)

    problem_type: Mapped[ProblemType] = mapped_column(Enum(ProblemType), nullable=False)
    answer: Mapped[dict] = mapped_column(JSONB, nullable=False)

    is_correct: Mapped[Optional[bool]] = mapped_column(Boolean, nullable=True)
    points_awarded: Mapped[Optional[int]] = mapped_column(Integer, nullable=True)

    eval_status: Mapped[Optional[EvalStatus]] = mapped_column(Enum(EvalStatus), nullable=True)
    eval_result: Mapped[Optional[dict]] = mapped_column(JSONB, nullable=True)

    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), onupdate=func.now()
    )

    attempt: Mapped[Attempt] = relationship(back_populates="answers")

    def __repr__(self) -> str:
        return f"<AttemptAnswer(id={self.id}, problem_id={self.problem_id})>"
