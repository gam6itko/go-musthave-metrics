CREATE TABLE IF NOT EXISTS public.counter
(
    "name" varchar NOT NULL,
    value  bigint  NOT NULL DEFAULT 0,
    CONSTRAINT counter_pk PRIMARY KEY ("name")
);

CREATE TABLE IF NOT EXISTS public.gauge
(
    "key" varchar          NOT NULL,
    value double precision NULL,
    CONSTRAINT gauge_pk PRIMARY KEY ("key")
);
