from contextlib import asynccontextmanager
from fastapi import FastAPI
from .shared.database import engine, Base
from .problems import models as problem_models  # noqa: F401 — register on Base
from .quizzes import models as quiz_models      # noqa: F401 — register on Base
from .problems.routes import router as problems_router
from .quizzes.routes import router as quizzes_router


@asynccontextmanager
async def lifespan(app: FastAPI):
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    yield


app = FastAPI(lifespan=lifespan)
app.include_router(problems_router)
app.include_router(quizzes_router)
