FROM debian:jessie-slim

RUN apt-get update && apt-get install -y git

COPY acid /usr/bin/acid
COPY runner.js /runner.js

CMD /usr/bin/acid
