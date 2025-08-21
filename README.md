## m-data-storage

TypeScript/Node server for historical OHLCV (1m) storage on PostgreSQL/TimescaleDB, following clean architecture.

### Quick start (dev)

1. Install deps: `yarn`
2. Copy `.env.example` to `.env` and set `API_TOKEN` and DB settings. Example:
   - `API_TOKEN=changeme`
   - `DB_HOST=localhost`, `DB_PORT=5466`, `DB_USER=user`, `DB_PASSWORD=password`, `DB_NAME=m-data-storage`
   - `DATABASE_URL=postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}`
3. Start dev: `yarn dev`
4. Build and run: `yarn build` then `yarn start`

### API and Docs

- Health: `GET /health`, Ready: `GET /ready`
- Read API: `GET /v1/ohlcv?broker_system_name=&instrument_symbol=&tf=&pageToken=`
- Ingest API: `POST /v1/ingest/ohlcv` (requires bearer token)
  - Body: `{ broker_system_name, instrument_symbol, items: [...] }`
- Task API: `GET /v1/tasks/next?broker_system_name=&instrument_symbols=s1,s2`, `POST /v1/tasks/:id/complete`
- Swagger UI: `/docs` (server URL configured via `SWAGGER_SERVER_URL`)

#### OpenAPI/Swagger schema

- JSON schema endpoint: `GET /openapi.json`
  - Note: there is no `/swagger.json`. Use `/openapi.json` instead.
- Examples:

```bash
# View schema in terminal (pretty-printed with jq)
curl -s http://localhost:8080/openapi.json | jq .

# Save schema to a file
curl -s http://localhost:8080/openapi.json -o openapi.json
```

- Swagger UI is available at `http://localhost:8080/docs`.
- The server URL shown in the docs is controlled by `SWAGGER_SERVER_URL` (defaults to `http://localhost:8080`).

All protected endpoints under `/v1/*` require `Authorization: Bearer <API_TOKEN>` or `X-Api-Key`.

### Scripts

- `yarn dev`: run `ts-node src/main.ts`
- `yarn build`: compile to `dist`
- `yarn start`: run `node dist/main.js`
- `yarn lint`: ESLint
- `yarn format`: Prettier
- `yarn drizzle:generate`: generate SQL migrations
- `yarn drizzle:migrate`: apply migrations
- `yarn db:seed`: seed demo brokers/instruments
- `yarn test`: run unit/integration tests

### Docker (dev)

- Start TimescaleDB locally: `docker compose -f docker-compose.dev.yml up -d`
- Init scripts create schemas `registry` and `timeseries`


### Drizzle Studio

```bash
npx drizzle-kit studio
```
