from fastapi import APIRouter, Depends
from ..shared.dependencies import get_current_user_id, CurrentUserIdDep

router = APIRouter(
    prefix="/quizzes",
    dependencies=[Depends(get_current_user_id)],
)


@router.get("/")
async def list_quizzes(user_id: CurrentUserIdDep):
    return None
