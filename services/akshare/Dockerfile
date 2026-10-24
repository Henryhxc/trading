FROM python:3.12-slim

WORKDIR /app
COPY services/akshare/requirements.lock ./requirements.lock
RUN pip install --no-cache-dir -r requirements.lock
COPY services/akshare/ ./services/akshare/

USER nobody
ENTRYPOINT ["python", "-m", "services.akshare.scheduler"]
