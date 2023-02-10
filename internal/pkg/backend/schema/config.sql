CREATE TABLE "config" (
  "id" INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  "version" int NOT NULL,
  "path" varchar(2048) NOT NULL,
  "value" varchar(2048) NOT NULL,
  "value_provider_id" varchar(128),
  "created_at" timestamp DEFAULT (now()),  
  "updated_at" timestamp DEFAULT (now())
);

CREATE TABLE "value_providers" (
  "id" varchar(128) PRIMARY KEY,
  "method" varchar(128) NOT NULL,
  "payload" jsonb NOT NULL,
  "created_at" timestamp DEFAULT (now()),
  "updated_at" timestamp DEFAULT (now())
);

CREATE TABLE "config_metadata" (
  "id" INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  "path" varchar(2048) NOT NULL UNIQUE,
  "latest_version" int NOT NULL DEFAULT 0,
  "current_version" int NOT NULL DEFAULT 0,
  "created_at" timestamp DEFAULT (now()),  
  "updated_at" timestamp DEFAULT (now()) 
);

ALTER TABLE "config" ADD FOREIGN KEY ("value_provider_id") REFERENCES "value_providers" ("id");