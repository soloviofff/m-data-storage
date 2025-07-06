-- Rollback for: Initial Schema
-- Description: Drop all initial tables for time series data
-- Created: 2024-07-04

-- Drop tables
DROP TABLE IF EXISTS aggregates;
DROP TABLE IF EXISTS trades;
DROP TABLE IF EXISTS order_books;
DROP TABLE IF EXISTS candles;
DROP TABLE IF EXISTS tickers;
