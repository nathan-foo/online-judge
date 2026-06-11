from fastapi.testclient import TestClient

from app import app

client = TestClient(app)

class TestHelloWorld:
    def test_hello_world(self) -> None:
        response = client.get("/")
        assert response.status_code == 200
        assert response.json() == {"message": "Hello from attempt-service"}
