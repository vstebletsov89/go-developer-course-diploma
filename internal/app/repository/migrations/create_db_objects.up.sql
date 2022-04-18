CREATE TYPE order_status AS ENUM (
    'REGISTERED',
    'INVALID',
    'PROCESSING',
    'PROCESSED'
);

CREATE TABLE users (
 id uuid    UNIQUE PRIMARY KEY,
 login      text UNIQUE NOT NULL,
 password   text NOT NULL,
 balance    decimal default 0,
 token      text,
 created_at timestamp
);

CREATE TABLE orders (
  id          text UNIQUE PRIMARY KEY NOT NULL,
  userID      uuid,
  status      order_status NOT NULL,
  accrual     decimal,
  uploaded_at timestamp
);

CREATE TABLE accruals (
  orderID   text UNIQUE NOT NULL,
  userID    uuid,
  processed boolean DEFAULT false,
  sum       decimal
);

CREATE TABLE withdrawals (
  orderID      text UNIQUE NOT NULL,
  userID       uuid,
  sum          decimal,
  status       order_status NOT NULL,
  processed_at timestamp
);
