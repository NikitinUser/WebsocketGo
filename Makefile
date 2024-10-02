SHELL=/bin/bash

UID := 1000

up:
	env UID=${UID} docker-compose up -d --build --remove-orphans
