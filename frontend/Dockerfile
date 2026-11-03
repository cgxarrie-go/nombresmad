FROM node:18-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm install --production=false
EXPOSE 3000
CMD ["npm", "start"]

FROM deps AS build
ARG REACT_APP_GOOGLE_CLIENT_ID
ENV REACT_APP_GOOGLE_CLIENT_ID=$REACT_APP_GOOGLE_CLIENT_ID
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["node", "server.js"]
