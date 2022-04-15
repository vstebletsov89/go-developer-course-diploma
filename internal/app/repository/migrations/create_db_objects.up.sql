CREATE TABLE IF NOT EXISTS users
(
    id              serial not null primary key,
    userID          text,
    password        text
);
CREATE UNIQUE INDEX IF NOT EXISTS login_ix ON users(userID);

CREATE TABLE IF NOT EXISTS orders
(
    id              serial not null primary key,
    userID          text,
    orderID         text,
    status          text,
    accrual         numeric,
    uploaded_at     TIMESTAMP
);

CREATE TABLE IF NOT EXISTS withdrawals
(
    id              serial not null primary key,
    orderID         text,
    sum             numeric,
    processed_at    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS balance
(
    id              serial not null primary key,
    user_uid        text,
    balance         numeric,
    withdrawn       numeric
);
