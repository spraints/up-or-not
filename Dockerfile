FROM python

WORKDIR /app
COPY . .

ENTRYPOINT ["python3", "main.py"]
