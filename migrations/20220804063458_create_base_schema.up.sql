CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE "Users"(
                        "id" UUID NOT NULL DEFAULT uuid_generate_v4(),
                        "username" VARCHAR(32) NOT NULL UNIQUE,
                        "email" VARCHAR(64) NOT NULL UNIQUE,
                        "hash" VARCHAR(255) NOT NULL,
                        "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE
    "Users" ADD PRIMARY KEY("id");

CREATE TABLE "UserToBoard"(
                              "id" UUID NOT NULL DEFAULT uuid_generate_v4(),
                              "board_id" UUID NOT NULL,
                              "user_id" UUID NOT NULL,
                              "access_type" UUID NOT NULL,
                              "description" VARCHAR(255) NOT NULL,
                              "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE
    "UserToBoard" ADD PRIMARY KEY("id");
COMMENT
    ON COLUMN
    "UserToBoard"."description" IS '"Person gave you a $1 privilege. Description: $2"';

CREATE TABLE "Board"(
                        "id" UUID NOT NULL DEFAULT uuid_generate_v4(),
                        "title" VARCHAR(255) NOT NULL,
                        "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE
    "Board" ADD PRIMARY KEY("id");

CREATE TABLE "BoardPrivileges"(
                                  "id" UUID NOT NULL DEFAULT uuid_generate_v4(),
                                  "privilege" VARCHAR(255) NOT NULL
);
ALTER TABLE
    "BoardPrivileges" ADD PRIMARY KEY("id");
CREATE TABLE "Note"(
                       "id" INTEGER NOT NULL,
                       "read_only" BOOLEAN NOT NULL,
                       "title" VARCHAR(255) NOT NULL,
                       "content" text NOT NULL,
                       "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       "board_bridge_id" UUID NOT NULL
);
CREATE INDEX "note_id_created_at_index" ON
    "Note"("id", "created_at");
CREATE INDEX "note_id_board_bridge_id_index" ON
    "Note"("id", "board_bridge_id");
ALTER TABLE
    "Note" ADD PRIMARY KEY("id");

CREATE TABLE "BoardToBoard"(
                               "id" UUID NOT NULL DEFAULT uuid_generate_v4(),
                               "root_board_id" UUID NOT NULL,
                               "subboard_id"
                                   UUID NOT NULL
);
ALTER TABLE
    "BoardToBoard" ADD PRIMARY KEY("id");
ALTER TABLE
    "UserToBoard" ADD CONSTRAINT "usertoboard_board_id_foreign" FOREIGN KEY("board_id") REFERENCES "Board"("id");
ALTER TABLE
    "UserToBoard" ADD CONSTRAINT "usertoboard_user_id_foreign" FOREIGN KEY("user_id") REFERENCES "Users"("id");
ALTER TABLE
    "UserToBoard" ADD CONSTRAINT "usertoboard_access_type_foreign" FOREIGN KEY("access_type") REFERENCES "BoardPrivileges"("id");
ALTER TABLE
    "BoardToBoard" ADD CONSTRAINT "boardtoboard_root_board_id_foreign" FOREIGN KEY("root_board_id") REFERENCES "Board"("id");
ALTER TABLE
    "BoardToBoard" ADD CONSTRAINT "boardtoboard_subboard_id_foreign" FOREIGN KEY("subboard_id") REFERENCES "Board"("id");
ALTER TABLE
    "Note" ADD CONSTRAINT "note_board_bridge_id_foreign" FOREIGN KEY("board_bridge_id") REFERENCES "BoardToBoard"("id");