from contextlib import asynccontextmanager
from fastapi import FastAPI
from .database import engine, Base


@asynccontextmanager
async def lifespan(app: FastAPI):
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    yield


app = FastAPI(lifespan=lifespan)


@app.get("/")
async def hello():
    return {"message": "Hello from attempt-service"}
