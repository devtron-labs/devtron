# syntax=docker/dockerfile:1
FROM python:3.10-slim-buster
RUN pip3 install pipenv
WORKDIR /app
COPY . .
RUN pipenv install --system --deploy --ignore-pipfile
CMD ["python","kafka_producer.py"]
