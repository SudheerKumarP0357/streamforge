# syntax=docker/dockerfile:1

ARG NODE_VERSION=24.11.0

FROM node:24-alpine AS base

WORKDIR /app

FROM base AS build

RUN --mount=type=bind,source=package.json,target=package.json \
    --mount=type=bind,source=package-lock.json,target=package-lock.json \
    --mount=type=cache,target=/root/.npm \
    npm ci

COPY . .
RUN npm run build

FROM base AS runner

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV HOSTNAME="::"
ENV PORT=3000

RUN addgroup --system --gid 1001 nodejs \
    && adduser  --system --uid 1001 nextjs

WORKDIR /app

COPY --from=build /app/build/standalone /app/
COPY --from=build /app/build/static/ /app/build/static/

EXPOSE 3000

ENTRYPOINT [ "node" ]
CMD [ "/app/server.js" ]
