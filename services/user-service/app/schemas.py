import uuid
from datetime import datetime
from typing import Optional
from pydantic import BaseModel, ConfigDict


class UserRead(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    clerk_user_id: str
    email: str
    first_name: Optional[str]
    last_name: Optional[str]
    is_admin: bool
    created_at: datetime
    updated_at: datetime


class UserCreate(BaseModel):
    clerk_user_id: str
    email: str
    first_name: Optional[str] = None
    last_name: Optional[str] = None
