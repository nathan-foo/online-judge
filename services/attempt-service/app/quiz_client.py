import os
import uuid
import httpx
from fastapi import HTTPException, status

QUIZ_SERVICE_URL = os.environ.get("QUIZ_SERVICE_URL", "http://quiz-service:8000")

client = httpx.AsyncClient(base_url=QUIZ_SERVICE_URL, timeout=5.0)


async def fetch_quiz_snapshot(user_id: str, quiz_id: uuid.UUID) -> dict:
    try:
        response = await client.get(
            f"/internal/{quiz_id}/snapshot",
            headers={"X-User-Id": user_id},
        )
    except httpx.HTTPError:
        raise HTTPException(
            status_code=status.HTTP_502_BAD_GATEWAY,
            detail="Quiz service unavailable",
        )
    if response.status_code == status.HTTP_404_NOT_FOUND:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Quiz not found",
        )
    if response.status_code != status.HTTP_200_OK:
        raise HTTPException(
            status_code=status.HTTP_502_BAD_GATEWAY,
            detail="Quiz service unavailable",
        )
    return response.json()


async def close() -> None:
    await client.aclose()
