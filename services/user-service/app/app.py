import os
from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException, Request, status
from svix.webhooks import Webhook, WebhookVerificationError
from .database import engine, Base
from .dependencies import AsyncSessionDep, CurrentUserDep
from .schemas import UserRead, UserCreate, UserUpdate
from . import models, user_service  # noqa: F401 — registers User on Base.metadata

try:
    with open("/run/secrets/clerk_webhook_secret") as f:
        CLERK_WEBHOOK_SECRET = f.read().strip()
except FileNotFoundError:
    CLERK_WEBHOOK_SECRET = os.environ["CLERK_WEBHOOK_SECRET"]


# TODO: Replace with Alembic migration later.
@asynccontextmanager
async def lifespan(app: FastAPI):
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    yield


app = FastAPI(lifespan=lifespan)


@app.post("/", status_code=status.HTTP_204_NO_CONTENT)
async def clerk_webhook(request: Request, session: AsyncSessionDep):
    try:
        event = Webhook(CLERK_WEBHOOK_SECRET).verify(await request.body(), dict(request.headers))
    except WebhookVerificationError:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid signature")

    if event["type"] == "user.created":
        data = event["data"]
        primary_id = data.get("primary_email_address_id")
        email = next(e["email_address"] for e in data["email_addresses"] if e["id"] == primary_id)
        await user_service.create_user(session, UserCreate(
            clerk_user_id=data["id"],
            email=email,
            avatar_url=data.get("image_url"),
            first_name=data.get("first_name"),
            last_name=data.get("last_name"),
        ))


@app.get("/me", response_model=UserRead)
async def get_me(current_user: CurrentUserDep):
    return current_user


@app.patch("/me", response_model=UserRead)
async def update_me(user_in: UserUpdate, current_user: CurrentUserDep, session: AsyncSessionDep):
    return await user_service.update_user(session, current_user, user_in)
