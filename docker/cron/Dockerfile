FROM willfarrell/crontab AS base

ENV TZ=Europe/Moscow
RUN apk add tzdata && ln -s /usr/share/zoneinfo/$TZ /etc/localtime

FROM base AS deploy

COPY docker/cron/config.json /opt/crontab/config.json
