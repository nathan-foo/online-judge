import uuid
import enum
from datetime import datetime
from typing import Optional
from sqlalchemy import (
    String,
    DateTime,
    Uuid,
    Integer,
    ForeignKey,
    Text,
    UniqueConstraint,
    Index,
    Boolean,
    Enum,
    CheckConstraint,
    text,
)
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.sql import func
from .database import Base


class ProblemType(str, enum.Enum):
    MULTIPLE_CHOICE = "multiple_choice"
    CODE = "code"


class Quiz(Base):
    __tablename__ = "quizzes"
    __table_args__ = (
        Index(
            "ix_quizzes_owner_active",
            "owner_id",
            text("created_at DESC"),
            postgresql_where=text("NOT is_deleted"),
        ),
        Index(
            "ix_quizzes_public_feed",
            text("created_at DESC"),
            postgresql_where=text("is_public AND is_published AND NOT is_deleted"),
        ),
    )

    id: Mapped[uuid.UUID] = mapped_column(Uuid, primary_key=True, default=uuid.uuid4)
    owner_id: Mapped[str] = mapped_column(String(255), nullable=False)

    title: Mapped[str] = mapped_column(String(255), nullable=False)
    description: Mapped[Optional[str]] = mapped_column(Text, nullable=True)

    is_published: Mapped[bool] = mapped_column(Boolean, default=False, nullable=False)
    is_public: Mapped[bool] = mapped_column(Boolean, default=False, nullable=False)
    is_deleted: Mapped[bool] = mapped_column(Boolean, default=False, nullable=False)

    problem_count: Mapped[int] = mapped_column(Integer, default=0, nullable=False)
    version: Mapped[int] = mapped_column(Integer, nullable=False)

    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now()
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), onupdate=func.now()
    )
    published_at: Mapped[Optional[datetime]] = mapped_column(
        DateTime(timezone=True), nullable=True
    )
    published_snapshot: Mapped[Optional[dict]] = mapped_column(JSONB, nullable=True)

    problems: Mapped[list["QuizProblem"]] = relationship(
        back_populates="quiz",
        order_by="QuizProblem.position",
        cascade="all, delete-orphan",
    )

    __mapper_args__ = {"version_id_col": version}

    def __repr__(self) -> str:
        return f"<Quiz(id={self.id}, title={self.title})>"


class QuizProblem(Base):
    __tablename__ = "quiz_problems"
    __table_args__ = (
        UniqueConstraint(
            "quiz_id",
            "position",
            name="uq_quiz_position",
            deferrable=True,
            initially="DEFERRED",
        ),
        CheckConstraint("position >= 1", name="ck_quiz_problem_position_positive"),
    )

    id: Mapped[uuid.UUID] = mapped_column(Uuid, primary_key=True, default=uuid.uuid4)
    quiz_id: Mapped[uuid.UUID] = mapped_column(
        Uuid, ForeignKey("quizzes.id", ondelete="CASCADE"), nullable=False
    )

    type: Mapped[ProblemType] = mapped_column(Enum(ProblemType), nullable=False)
    title: Mapped[str] = mapped_column(String(255), nullable=False)
    payload: Mapped[dict] = mapped_column(JSONB, nullable=False)

    position: Mapped[int] = mapped_column(Integer, nullable=False)
    points: Mapped[int] = mapped_column(Integer, nullable=False, default=1000)

    quiz: Mapped[Quiz] = relationship(back_populates="problems")

    def __repr__(self) -> str:
        return f"<QuizProblem(id={self.id}, title={self.title})>"
