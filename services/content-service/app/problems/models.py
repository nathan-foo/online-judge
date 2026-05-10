import uuid
import enum
from sqlalchemy import String, DateTime, Uuid, Enum, Boolean, Index, text
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import Mapped, mapped_column
from sqlalchemy.sql import func
from datetime import datetime
from ..shared.database import Base

class ProblemType(enum.Enum):
    MULTIPLE_CHOICE = "multiple_choice"
    CODE = "code"

class Problem(Base):
    __tablename__ = "problems"
    __table_args__ = (
        Index(
            "ix_problems_owner_active",
            "owner_id",
            text("created_at DESC"),
            postgresql_where=text("NOT is_deleted"),
        ),
    )

    id: Mapped[uuid.UUID] = mapped_column(Uuid, primary_key=True, default=uuid.uuid4)
    owner_id: Mapped[str] = mapped_column(String(255), nullable=False)

    type: Mapped[ProblemType] = mapped_column(Enum(ProblemType), nullable=False)

    title: Mapped[str] = mapped_column(String(255), nullable=False)
    payload: Mapped[dict] = mapped_column(JSONB, nullable=False)

    is_deleted: Mapped[bool] = mapped_column(Boolean, default=False, nullable=False)

    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now()
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), onupdate=func.now()
    )

    def __repr__(self) -> str:
        return f"<Problem(id={self.id}, title={self.title})>"
