import uuid
from sqlalchemy import String, DateTime, Uuid, Integer, ForeignKey, Text, UniqueConstraint, Index, Boolean, text
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.sql import func
from typing import Optional
from datetime import datetime
from ..shared.database import Base
from ..problems.models import Problem

class Quiz(Base):
    __tablename__ = "quizzes"
    __table_args__ = (
        Index(
            "ix_quizzes_owner_active",
            "owner_id",
            postgresql_where=text("is_deleted = false"),
        ),
    )

    id: Mapped[uuid.UUID] = mapped_column(Uuid, primary_key=True, default=uuid.uuid4)
    owner_id: Mapped[str] = mapped_column(String(255), nullable=False)

    title: Mapped[str] = mapped_column(String(255), nullable=False)
    description: Mapped[Optional[str]] = mapped_column(Text, nullable=True)

    is_published: Mapped[bool] = mapped_column(Boolean, default=False, index=True, nullable=False)
    is_public: Mapped[bool] = mapped_column(Boolean, default=False, index=True, nullable=False)
    is_deleted: Mapped[bool] = mapped_column(Boolean, default=False, nullable=False)

    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now()
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), onupdate=func.now()
    )

    problems: Mapped[list["QuizProblem"]] = relationship(
        back_populates="quiz",
        order_by="QuizProblem.position",
        cascade="all, delete-orphan"
    )

    def __repr__(self) -> str:
        return f"<Quiz(id={self.id}, title={self.title})>"

class QuizProblem(Base):
    __tablename__ = "quiz_problems"
    __table_args__ = (
        UniqueConstraint("quiz_id", "position", name="uq_quiz_position"),
        Index("ix_quiz_problems_problem_id", "problem_id"),
    )

    quiz_id: Mapped[uuid.UUID] = mapped_column(Uuid, ForeignKey("quizzes.id", ondelete="CASCADE"), primary_key=True)
    problem_id: Mapped[uuid.UUID] = mapped_column(Uuid, ForeignKey("problems.id", ondelete="RESTRICT"), primary_key=True)

    position: Mapped[int] = mapped_column(Integer, nullable=False)
    points: Mapped[int] = mapped_column(Integer, nullable=False, default=1)

    quiz: Mapped[Quiz] = relationship(back_populates="problems")
    problem: Mapped[Problem] = relationship()
