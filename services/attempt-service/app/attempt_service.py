import uuid
from datetime import datetime, timezone
from fastapi import HTTPException, status
from sqlalchemy import select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload
from .models import Attempt, AttemptAnswer, AttemptStatus, EvalStatus, ProblemType
from .schemas import AnswerPayload, AttemptAnswerRead, AttemptRead
from . import quiz_client


def _taker_view_problem(problem: dict, include_answers: bool) -> dict:
    payload = dict(problem["payload"])
    if payload["type"] == ProblemType.MULTIPLE_CHOICE:
        if not include_answers:
            payload.pop("correct_choice_ids", None)
            payload.pop("explanation", None)
    else:
        payload.pop("reference_solutions", None)
        if not include_answers:
            payload["test_cases"] = [
                t for t in payload["test_cases"] if t["is_example"]
            ]
    return {**problem, "payload": payload}


def _taker_view(snapshot: dict, include_answers: bool) -> dict:
    return {
        "id": snapshot["id"],
        "title": snapshot["title"],
        "description": snapshot.get("description"),
        "problems": [
            _taker_view_problem(p, include_answers) for p in snapshot["problems"]
        ],
    }


def to_attempt_read(attempt: Attempt) -> AttemptRead:
    return AttemptRead(
        id=attempt.id,
        quiz_id=attempt.quiz_id,
        quiz_version=attempt.quiz_version,
        quiz_title=attempt.quiz_title,
        quiz=_taker_view(
            attempt.quiz_snapshot,
            include_answers=attempt.status == AttemptStatus.GRADED,
        ),
        status=attempt.status,
        score=attempt.score,
        max_score=attempt.max_score,
        started_at=attempt.started_at,
        submitted_at=attempt.submitted_at,
        answers=[AttemptAnswerRead.model_validate(a) for a in attempt.answers],
    )


def _problems_by_id(snapshot: dict) -> dict[str, dict]:
    return {p["id"]: p for p in snapshot["problems"]}


async def start_attempt(
    session: AsyncSession, user_id: str, quiz_id: uuid.UUID
) -> Attempt:
    snapshot = await quiz_client.fetch_quiz_snapshot(user_id, quiz_id)
    attempt = Attempt(
        user_id=user_id,
        quiz_id=quiz_id,
        quiz_version=snapshot["version"],
        quiz_title=snapshot["title"],
        quiz_snapshot=snapshot,
        max_score=sum(p["points"] for p in snapshot["problems"]),
        answers=[],
    )
    session.add(attempt)
    try:
        await session.flush()
    except IntegrityError as exc:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="An attempt for this quiz is already in progress",
        ) from exc
    return await get_owned_attempt(session, user_id, attempt.id)


async def list_attempts(
    session: AsyncSession, user_id: str, limit: int, offset: int
) -> list[Attempt]:
    result = await session.execute(
        select(Attempt)
        .where(Attempt.user_id == user_id)
        .order_by(Attempt.started_at.desc())
        .limit(limit)
        .offset(offset)
    )
    return list(result.scalars().all())


async def get_owned_attempt(
    session: AsyncSession, user_id: str, attempt_id: uuid.UUID
) -> Attempt:
    result = await session.execute(
        select(Attempt)
        .where(Attempt.id == attempt_id, Attempt.user_id == user_id)
        .options(selectinload(Attempt.answers))
    )
    attempt = result.scalar_one_or_none()
    if not attempt:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Attempt not found",
        )
    return attempt


def _validate_answer(problem: dict, answer_in: AnswerPayload) -> None:
    if answer_in.type != problem["type"]:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Answer type does not match problem type",
        )
    payload = problem["payload"]
    if answer_in.type == ProblemType.MULTIPLE_CHOICE:
        choice_ids = {c["id"] for c in payload["choices"]}
        unknown = set(answer_in.selected_choice_ids) - choice_ids
        if unknown:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Answer references unknown choices",
            )
        if not payload["multiple_correct"] and len(answer_in.selected_choice_ids) != 1:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Exactly one choice must be selected",
            )
    else:
        if answer_in.language not in payload["languages"]:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Language not allowed for this problem",
            )


async def save_answer(
    session: AsyncSession,
    attempt: Attempt,
    problem_id: uuid.UUID,
    answer_in: AnswerPayload,
) -> AttemptAnswer:
    if attempt.status != AttemptStatus.IN_PROGRESS:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="Attempt has already been submitted",
        )
    problem = _problems_by_id(attempt.quiz_snapshot).get(str(problem_id))
    if not problem:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Problem not found in attempt",
        )
    _validate_answer(problem, answer_in)
    answer = next((a for a in attempt.answers if a.problem_id == problem_id), None)
    if answer is None:
        answer = AttemptAnswer(
            problem_id=problem_id,
            problem_type=answer_in.type,
            answer=answer_in.model_dump(mode="json"),
        )
        attempt.answers.append(answer)
    else:
        answer.answer = answer_in.model_dump(mode="json")
    await session.flush()
    await session.refresh(answer)
    return answer


def _grade_multiple_choice(answer: AttemptAnswer, problem: dict) -> int:
    correct = set(answer.answer["selected_choice_ids"]) == set(
        problem["payload"]["correct_choice_ids"]
    )
    answer.is_correct = correct
    answer.points_awarded = problem["points"] if correct else 0
    return answer.points_awarded


def _build_eval_request(attempt: Attempt, answer: AttemptAnswer, problem: dict) -> dict:
    payload = problem["payload"]
    return {
        "attempt_id": str(attempt.id),
        "problem_id": str(answer.problem_id),
        "language": answer.answer["language"],
        "source_code": answer.answer["source_code"],
        "test_cases": payload["test_cases"],
        "time_limit_ms": payload["time_limit_ms"],
        "memory_limit_mb": payload["memory_limit_mb"],
        "points": problem["points"],
    }


async def submit_attempt(
    session: AsyncSession, attempt: Attempt
) -> tuple[Attempt, list[dict]]:
    if attempt.status != AttemptStatus.IN_PROGRESS:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="Attempt has already been submitted",
        )
    problems = _problems_by_id(attempt.quiz_snapshot)
    score = 0
    eval_requests: list[dict] = []
    for answer in attempt.answers:
        problem = problems[str(answer.problem_id)]
        if answer.problem_type == ProblemType.MULTIPLE_CHOICE:
            score += _grade_multiple_choice(answer, problem)
        else:
            answer.eval_status = EvalStatus.PENDING
            eval_requests.append(_build_eval_request(attempt, answer, problem))
    attempt.submitted_at = datetime.now(timezone.utc)
    if eval_requests:
        attempt.status = AttemptStatus.GRADING
    else:
        attempt.status = AttemptStatus.GRADED
        attempt.score = score
    await session.flush()
    return await get_owned_attempt(session, attempt.user_id, attempt.id), eval_requests


async def apply_eval_result(session: AsyncSession, payload: dict) -> None:
    result = await session.execute(
        select(Attempt)
        .where(Attempt.id == uuid.UUID(payload["attempt_id"]))
        .options(selectinload(Attempt.answers))
    )
    attempt = result.scalar_one_or_none()
    if not attempt or attempt.status != AttemptStatus.GRADING:
        return
    problem_id = uuid.UUID(payload["problem_id"])
    answer = next((a for a in attempt.answers if a.problem_id == problem_id), None)
    if not answer or answer.eval_status != EvalStatus.PENDING:
        return
    if payload.get("status") == "done":
        answer.eval_status = EvalStatus.DONE
        answer.is_correct = payload.get("passed") == payload.get("total")
        answer.points_awarded = payload.get("points_awarded", 0)
    else:
        answer.eval_status = EvalStatus.ERROR
        answer.is_correct = False
        answer.points_awarded = 0
    answer.eval_result = payload
    code_answers = [a for a in attempt.answers if a.problem_type == ProblemType.CODE]
    if all(a.eval_status in (EvalStatus.DONE, EvalStatus.ERROR) for a in code_answers):
        attempt.status = AttemptStatus.GRADED
        attempt.score = sum(a.points_awarded or 0 for a in attempt.answers)
    await session.flush()
