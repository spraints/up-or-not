FROM python

WORKDIR /app
COPY . .

RUN pip install pyping

ENTRYPOINT ["python3", "main.py"]
