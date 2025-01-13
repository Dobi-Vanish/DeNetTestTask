FROM alpine:latest

RUN mkdir /app

COPY rewardApp /app

CMD [ "/app/rewardApp"]