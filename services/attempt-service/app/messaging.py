import json
import os
from typing import Awaitable, Callable, Optional
import aio_pika

RABBITMQ_URL = os.environ.get("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")

CODE_EVAL_EXCHANGE = "code_eval"
EVAL_REQUESTS_QUEUE = "code_eval.requests"
EVAL_RESULTS_QUEUE = "code_eval.results"


class Broker:
    def __init__(self) -> None:
        self.connection: Optional[aio_pika.abc.AbstractRobustConnection] = None
        self.channel: Optional[aio_pika.abc.AbstractChannel] = None
        self.exchange: Optional[aio_pika.abc.AbstractExchange] = None
        self.results_queue: Optional[aio_pika.abc.AbstractQueue] = None

    async def connect(self) -> None:
        self.connection = await aio_pika.connect_robust(RABBITMQ_URL)
        self.channel = await self.connection.channel()
        await self.channel.set_qos(prefetch_count=10)
        self.exchange = await self.channel.declare_exchange(
            CODE_EVAL_EXCHANGE, aio_pika.ExchangeType.DIRECT, durable=True
        )
        requests_queue = await self.channel.declare_queue(EVAL_REQUESTS_QUEUE, durable=True)
        await requests_queue.bind(self.exchange, routing_key=EVAL_REQUESTS_QUEUE)
        self.results_queue = await self.channel.declare_queue(EVAL_RESULTS_QUEUE, durable=True)
        await self.results_queue.bind(self.exchange, routing_key=EVAL_RESULTS_QUEUE)

    async def close(self) -> None:
        if self.connection:
            await self.connection.close()

    async def publish_eval_request(self, message: dict) -> None:
        await self.exchange.publish(
            aio_pika.Message(
                body=json.dumps(message).encode(),
                content_type="application/json",
                delivery_mode=aio_pika.DeliveryMode.PERSISTENT,
            ),
            routing_key=EVAL_REQUESTS_QUEUE,
        )

    async def consume_results(self, handler: Callable[[dict], Awaitable[None]]) -> None:
        async def _on_message(message: aio_pika.abc.AbstractIncomingMessage) -> None:
            async with message.process(requeue=False):
                await handler(json.loads(message.body))

        await self.results_queue.consume(_on_message)


broker = Broker()
