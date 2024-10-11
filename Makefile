SHELL=/bin/bash

UID := 1000

env:
	cp .env.example .env

up:
	env UID=${UID} docker compose up -d --build --remove-orphans

compose:
	env UID=${UID} docker compose up --build --remove-orphans
