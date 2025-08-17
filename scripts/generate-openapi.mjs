import { OpenAPIRegistry, OpenApiGeneratorV3 } from '@asteasolutions/zod-to-openapi';
import { z } from 'zod';
import { createRequire } from 'node:module';
import { writeFileSync, mkdirSync } from 'node:fs';
import path from 'node:path';
const require = createRequire(import.meta.url);

require('dotenv-expand').expand(require('dotenv').config());

const openapi = await import('../dist/interfaces/http/openapi.js');

const registry = new OpenAPIRegistry();

const {
  ReadResponseSchema,
  IngestBodySchema,
  IngestResponseSchema,
  BrokersUpsertBodySchema,
  InstrumentsUpsertBodySchema,
  BrokersListResponseSchema,
  InstrumentsListResponseSchema,
  RemovedCountResponseSchema,
  TaskCompleteBodySchema,
  TaskCompleteResponseSchema,
  TasksListResponseSchema,
  ErrorResponseSchema,
  OkResponseSchema,
} = openapi;

// Register components
registry.register('ReadResponse', ReadResponseSchema);
registry.register('IngestBody', IngestBodySchema);
registry.register('IngestResponse', IngestResponseSchema ?? z.object({ inserted: z.number(), updated: z.number(), skipped: z.number() }));
registry.register('BrokersUpsertBody', BrokersUpsertBodySchema ?? z.array(z.object({ code: z.string(), name: z.string(), isActive: z.boolean().optional() })));
registry.register('InstrumentsUpsertBody', InstrumentsUpsertBodySchema ?? z.array(z.object({ symbol: z.string(), name: z.string().optional(), isActive: z.boolean().optional() })));
registry.register('BrokersListResponse', BrokersListResponseSchema ?? z.object({ items: z.array(z.object({ id: z.number(), code: z.string(), name: z.string(), is_active: z.boolean() })) }));
registry.register('InstrumentsListResponse', InstrumentsListResponseSchema ?? z.object({ items: z.array(z.object({ id: z.number(), symbol: z.string(), name: z.string().nullable().optional(), is_active: z.boolean() })) }));
registry.register('RemovedCountResponse', RemovedCountResponseSchema ?? z.object({ removed: z.number().int().nonnegative() }));
registry.register('TasksListResponse', TasksListResponseSchema ?? z.object({ items: z.array(z.any()) }));
registry.register('TaskCompleteBody', TaskCompleteBodySchema ?? z.object({ status: z.enum(['done','failed']) }));
registry.register('TaskCompleteResponse', TaskCompleteResponseSchema ?? z.object({ item: z.any() }));
registry.register('ErrorResponse', ErrorResponseSchema ?? z.object({ code: z.string(), message: z.string() }));
registry.register('OkResponse', OkResponseSchema ?? z.object({ ok: z.boolean() }));

const generator = new OpenApiGeneratorV3(registry.definitions);
const ref = (name) => ({ $ref: `#/components/schemas/${name}` });

const document = generator.generateDocument({
  openapi: '3.0.3',
  info: { title: 'm-data-storage API', version: '1.0.0' },
  servers: [{ url: process.env.SWAGGER_SERVER_URL || 'http://localhost:8080' }],
  paths: {},
});

// Read API with query params
document.paths['/v1/ohlcv'] = {
  get: {
    summary: 'Read OHLCV bars with optional aggregation',
    parameters: [
      { name: 'broker_id', in: 'query', required: true, schema: { type: 'integer' } },
      { name: 'instrument_id', in: 'query', required: true, schema: { type: 'integer' } },
      { name: 'tf', in: 'query', required: false, schema: { type: 'string', enum: ['1m','5m','15m','30m','1h','4h','1d'], default: '1m' } },
      { name: 'pageToken', in: 'query', required: false, schema: { type: 'string' } },
    ],
    responses: {
      '200': { description: 'OK', content: { 'application/json': { schema: ref('ReadResponse') } } },
      '400': { description: 'Bad Request', content: { 'application/json': { schema: ref('ErrorResponse') } } },
      '401': { description: 'Unauthorized', content: { 'application/json': { schema: ref('ErrorResponse') } } },
    },
  },
};

// Ingest API
document.paths['/v1/ingest/ohlcv'] = {
  post: {
    summary: 'Ingest OHLCV batch',
    requestBody: { required: true, content: { 'application/json': { schema: ref('IngestBody') } } },
    responses: {
      '200': { description: 'OK', content: { 'application/json': { schema: ref('IngestResponse') } } },
      '400': { description: 'Bad Request', content: { 'application/json': { schema: ref('ErrorResponse') } } },
      '401': { description: 'Unauthorized', content: { 'application/json': { schema: ref('ErrorResponse') } } },
    },
  },
};

// Tasks
document.paths['/v1/tasks/next'] = {
  get: {
    summary: 'Reserve next tasks',
    parameters: [
      { name: 'broker_id', in: 'query', required: false, schema: { type: 'integer' } },
      { name: 'instrument_ids', in: 'query', required: false, schema: { type: 'string', description: 'Comma-separated instrument ids' } },
      { name: 'limit', in: 'query', required: false, schema: { type: 'integer', default: 10, maximum: 100 } },
      { name: 'leaseSeconds', in: 'query', required: false, schema: { type: 'integer', default: 60, maximum: 3600 } },
    ],
    responses: {
      '200': { description: 'OK', content: { 'application/json': { schema: ref('TasksListResponse') } } },
      '400': { description: 'Bad Request', content: { 'application/json': { schema: ref('ErrorResponse') } } },
      '401': { description: 'Unauthorized', content: { 'application/json': { schema: ref('ErrorResponse') } } },
    },
  },
};

document.paths['/v1/tasks/{id}/complete'] = {
  post: {
    summary: 'Complete reserved task',
    parameters: [{ name: 'id', in: 'path', required: true, schema: { type: 'string', format: 'uuid' } }],
    requestBody: { required: true, content: { 'application/json': { schema: ref('TaskCompleteBody') } } },
    responses: {
      '200': { description: 'OK', content: { 'application/json': { schema: ref('TaskCompleteResponse') } } },
      '400': { description: 'Bad Request', content: { 'application/json': { schema: ref('ErrorResponse') } } },
      '401': { description: 'Unauthorized', content: { 'application/json': { schema: ref('ErrorResponse') } } },
      '409': { description: 'Conflict', content: { 'application/json': { schema: ref('ErrorResponse') } } },
    },
  },
};

// Registry
document.paths['/v1/admin/brokers'] = {
  get: { summary: 'List brokers', responses: { '200': { description: 'OK', content: { 'application/json': { schema: ref('BrokersListResponse') } } } } },
  post: { summary: 'Upsert brokers', requestBody: { required: true, content: { 'application/json': { schema: ref('BrokersUpsertBody') } } }, responses: { '200': { description: 'OK', content: { 'application/json': { schema: ref('OkResponse') } } }, '400': { description: 'Bad Request', content: { 'application/json': { schema: ref('ErrorResponse') } } } } },
};

document.paths['/v1/admin/instruments'] = {
  get: { summary: 'List instruments', responses: { '200': { description: 'OK', content: { 'application/json': { schema: ref('InstrumentsListResponse') } } } } },
  post: { summary: 'Upsert instruments', requestBody: { required: true, content: { 'application/json': { schema: ref('InstrumentsUpsertBody') } } }, responses: { '200': { description: 'OK', content: { 'application/json': { schema: ref('OkResponse') } } }, '400': { description: 'Bad Request', content: { 'application/json': { schema: ref('ErrorResponse') } } } } },
};

document.paths['/v1/admin/watch/broker/{broker_id}'] = {
  delete: { summary: 'Unwatch all instruments for broker', parameters: [{ name: 'broker_id', in: 'path', required: true, schema: { type: 'integer' } }], responses: { '200': { description: 'OK', content: { 'application/json': { schema: ref('RemovedCountResponse') } } } } },
};

document.paths['/v1/admin/watch/instrument/{instrument_id}'] = {
  delete: { summary: 'Unwatch instrument globally', parameters: [{ name: 'instrument_id', in: 'path', required: true, schema: { type: 'integer' } }], responses: { '200': { description: 'OK', content: { 'application/json': { schema: ref('RemovedCountResponse') } } } } },
};

document.paths['/v1/admin/watch/{broker_id}/{instrument_id}'] = {
  delete: { summary: 'Unwatch specific broker-instrument pair', parameters: [
    { name: 'broker_id', in: 'path', required: true, schema: { type: 'integer' } },
    { name: 'instrument_id', in: 'path', required: true, schema: { type: 'integer' } },
  ], responses: { '200': { description: 'OK', content: { 'application/json': { schema: ref('RemovedCountResponse') } } } } },
};

const outDir = path.join(process.cwd(), 'Docs');
mkdirSync(outDir, { recursive: true });
const outFile = path.join(outDir, 'openapi.json');
writeFileSync(outFile, JSON.stringify(document, null, 2));
console.log(`OpenAPI written to ${outFile}`);


