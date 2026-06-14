from fastapi import FastAPI

app = FastAPI()


@app.get("/healthz", include_in_schema=False)
async def healthz():
    return {"status": "ok"}


@app.get("/readyz", include_in_schema=False)
async def readyz():
    return {"status": "ready"}


@app.get("/")
async def root():
    return {"message": "Hello, world!"}
