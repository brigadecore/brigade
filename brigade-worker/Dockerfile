FROM node:8-alpine

WORKDIR /home/src
COPY brigade-worker/ /home/src/
RUN yarn build

CMD yarn -s start
