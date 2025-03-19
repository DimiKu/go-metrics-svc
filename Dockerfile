FROM postgres:latest

ENV POSTGRES_DB=metric_db
ENV POSTGRES_USER=myuser
#ENV POSTGRES_PASSWORD=mypassword

# COPY init.sql /docker-entrypoint-initdb.d/

# Открываем порт PostgreSQL
EXPOSE 5432