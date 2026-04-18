FROM node:25-alpine3.22 AS builder

WORKDIR /user/src/app

COPY package*.json ./

RUN npm install

COPY . . 

RUN npm run build

FROM nginx:stable-alpine

WORKDIR /usr/share/nginx/html
RUN rm -rf ./*

COPY --from=builder /user/src/app/dist .

COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]