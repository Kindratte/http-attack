FROM ubuntu:18.04

RUN mkdir /http-attack
COPY http-attack /http-attack/http-attack
COPY entrypoint.sh /http-attack/entrypoint.sh
RUN chmod 777 /http-attack/http-attack
RUN chmod 777 /http-attack/entrypoint.sh

ENTRYPOINT ["/http-attack/entrypoint.sh"]
